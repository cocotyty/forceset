package forceset

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Set(dst interface{}, src interface{}, opts ...Option) error {
	return ForceSet(reflect.ValueOf(dst).Elem(), src, opts...)
}

func ForceSet(value reflect.Value, i interface{}, opts ...Option) error {
	var opt SetOption
	opt.Mappers = map[MapperType]Mapper{}
	opt.Tag = "json"
	opt.Decoder = json.Unmarshal
	if len(opts) != 0 {
		for _, fn := range opts {
			fn(&opt)
		}
	}
	return forceSet(value, i, opt, "")
}

func forceSet(value reflect.Value, i interface{}, opt SetOption, tag string) error {
	if i == nil {
		return nil
	}
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	var bErr error
	iv := reflect.ValueOf(i)
	if m, ok := opt.Mappers[MapperType{value.Type(), iv.Type()}]; ok {
		return m(value, iv, tag)
	}
	switch value.Kind() {
	case reflect.String:
		value.SetString(toString(i, opt))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := toInt(i, opt)
		if err == nil {
			value.SetInt(i64)
			return nil
		}
		bErr = err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i64, err := toUint(i, opt)
		if err == nil {
			value.SetUint(i64)
			return nil
		}
		bErr = err
	case reflect.Bool:
		b, err := toBool(i, opt)
		if err == nil {
			value.SetBool(b)
			return nil
		}
		bErr = err
	case reflect.Float32:
		f, err := toFloat(i, opt)
		if err == nil {
			value.SetFloat(f)
			return nil
		}
		bErr = err
	case reflect.Float64:
		f, err := toFloat(i, opt)
		if err == nil {
			value.SetFloat(f)
			return nil
		}
		bErr = err

	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			data, err := toBytes(i, opt)
			if err == nil {
				value.SetBytes(data)
				return nil
			}
		}
	}
	if iv.Type() == value.Type() {
		value.Set(iv)
		return nil
	}
	if value.Type().Kind() == reflect.Interface {
		if iv.Type().Implements(value.Type()) {
			value.Set(iv)
			return nil
		}
	}

	if iv.Type().ConvertibleTo(value.Type()) {
		converted := iv.Convert(value.Type())
		value.Set(converted)
		return nil
	}

	if iv.Type().AssignableTo(value.Type()) {
		value.Set(iv)
		return nil
	}

	if bErr != nil {
		return bErr
	}

	switch value.Kind() {
	case reflect.Slice:
		for iv.Kind() == reflect.Ptr {
			if iv.IsNil() {
				return nil
			}
			iv = iv.Elem()
		}
		if iv.Type().Kind() == reflect.Slice || iv.Type().Kind() == reflect.Array {
			proxyValue := reflect.MakeSlice(value.Type(), iv.Len(), iv.Len())
			size := iv.Len()
			for n := 0; n < size; n++ {
				elm := iv.Index(n)
				err := forceSet(proxyValue.Index(n), elm.Interface(), opt, "")
				if err != nil {
					return err
				}
			}
			value.Set(proxyValue)
			return nil
		}
		switch iv.Type().Kind() {
		case reflect.Map:
			return map2slice(value, iv, opt)
		case reflect.Struct:
			return struct2slice(value, iv, opt)
		}
	case reflect.Struct:
		for iv.Kind() == reflect.Ptr {
			if iv.IsNil() {
				return nil
			}
			iv = iv.Elem()
		}
		if iv.Type() == value.Type() {
			value.Set(iv)
			return nil
		}
		if iv.Type().ConvertibleTo(value.Type()) {
			converted := iv.Convert(value.Type())
			value.Set(converted)
			return nil
		}

		if iv.Type().AssignableTo(value.Type()) {
			value.Set(iv)
			return nil
		}
		switch iv.Kind() {
		case reflect.Struct:
			return struct2Struct(value, iv, opt)
		case reflect.Map:
			_, err := map2Struct(value, iv, opt)
			return err
		case reflect.String:
			if opt.Decoder != nil {
				return decodeStruct(value, iv, opt)
			}
		case reflect.Slice:
			if opt.Decoder != nil && iv.Type().Elem().Kind() == reflect.Uint8 {
				return decodeStruct(value, iv, opt)
			}
		}
	case reflect.Map:
		for iv.Kind() == reflect.Ptr {
			if iv.IsNil() {
				return nil
			}
			iv = iv.Elem()
		}
		if iv.Type() == value.Type() {
			value.Set(iv)
			return nil
		}
		if iv.Type().ConvertibleTo(value.Type()) {
			converted := iv.Convert(value.Type())
			value.Set(converted)
			return nil
		}

		if iv.Type().AssignableTo(value.Type()) {
			value.Set(iv)
			return nil
		}
		switch iv.Kind() {
		case reflect.Struct:
			return struct2map(value, iv, opt)
		case reflect.Map:
			return map2map(value, iv, opt)

			//case reflect.String:
			//
		}
	}

	return errors.New("force set type(" + iv.Type().String() + ") into type(" + value.Type().String() + ") failed")
}

var empty = reflect.Value{}

func decodeStruct(dst, src reflect.Value, opt SetOption) error {
	if src.Kind() == reflect.String {
		data := src.Convert(reflect.ValueOf(``).Type()).Interface().(string)
		return opt.Decoder([]byte(data), dst.Addr().Interface())

	}
	data := src.Convert(reflect.ValueOf([]byte(``)).Type()).Interface().([]byte)
	return opt.Decoder(data, dst.Addr().Interface())
}

func struct2Struct(dst, src reflect.Value, opt SetOption) error {
	num := dst.NumField()
	for i := 0; i < num; i++ {
		df := dst.Field(i)
		dt := dst.Type().Field(i)
		if dt.Anonymous {
			err := struct2Struct(df, src, opt)
			if err != nil {
				return err
			}
			continue
		}
		if dt.PkgPath != "" {
			continue
		}
		sf := src.FieldByName(dt.Name)
		if sf == empty {
			continue
		}
		if dt.Type == sf.Type() {
			df.Set(sf)
			continue
		}
		err := setPtr(df, sf, opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func setPtr(dst reflect.Value, src reflect.Value, opt SetOption) error {
	typ := dst.Type()
	val := dst
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		v := reflect.New(typ)
		val.Set(v)
		val = val.Elem()
	}
	return forceSet(val, src.Interface(), opt, "")
}

// dst struct
// src map
func map2Struct(dst, src reflect.Value, opt SetOption) (count int, err error) {
	srcType := src.Type()
	kt := srcType.Key()
	if kt.Kind() != reflect.String {
		return 0, errors.New("map key type must be string")
	}
	fn := dst.NumField()
	dstType := dst.Type()
	for i := 0; i < fn; i++ {
		sf := dst.Field(i)
		st := dstType.Field(i)
		fieldValue := sf
		if st.Anonymous {
			typ := st.Type
			var tempValue reflect.Value
			var rootValue reflect.Value
			for typ.Kind() == reflect.Ptr {
				value := reflect.New(typ.Elem())
				if rootValue == empty {
					rootValue = value
					tempValue = rootValue.Elem()
					typ = tempValue.Type()
					continue
				}
				tempValue.Set(value)
				tempValue = tempValue.Elem()
				typ = tempValue.Type()
			}
			if tempValue.Kind() == reflect.Struct {
				cnt, err := map2Struct(tempValue, src, opt)
				if err != nil {
					return 0, err
				}
				if cnt == 0 {
					continue
				}
				count++
				fieldValue.Set(rootValue)
			}
			continue
		}
		if st.PkgPath != "" {
			continue
		}
		tag := st.Tag.Get(opt.Tag)
		names := strings.Split(strings.Split(tag, ";")[0], " ")
		if len(names) == 1 && names[0] == "" {
			names = []string{st.Name}
		}

		var value reflect.Value
		for _, name := range names {
			value = src.MapIndex(reflect.ValueOf(name))
			if value != empty {
				break
			}
		}
		if value == empty {
			continue
		}
		err := forceSet(fieldValue, value.Interface(), opt, tag)
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

// dst map
// src struct
func struct2map(dst, src reflect.Value, opt SetOption) error {
	valueType := dst.Type().Elem()
	keyType := dst.Type().Key()
	numField := src.NumField()
	srcType := src.Type()
	for i := 0; i < numField; i++ {
		field := src.Field(i)
		structField := srcType.Field(i)
		if structField.Anonymous {
			f := field
			for f.Kind() == reflect.Ptr {
				if f.IsNil() {
					break
				}
				f = field.Elem()
			}
			err := struct2map(dst, f, opt)
			if err != nil {
				return err
			}
			continue
		}
		if structField.PkgPath != "" {
			continue
		}
		var tag = structField.Tag.Get(opt.Tag)
		root, val := ptrValue(valueType)
		err := forceSet(val, field.Interface(), opt, tag)
		if err != nil {
			return err
		}
		keyName := strings.Split(strings.Split(tag, ";")[0], " ")[0]
		if keyName == "" {
			keyName = structField.Name
		}
		k := reflect.New(keyType)
		err = forceSet(k.Elem(), keyName, opt, tag)
		if err != nil {
			return err
		}
		dst.SetMapIndex(k.Elem(), root.Elem())
	}
	return nil
}

func ptrValue(typ reflect.Type) (root reflect.Value, val reflect.Value) {
	root = reflect.New(typ)

	val = root.Elem()

	for typ.Kind() == reflect.Ptr {
		v := reflect.New(typ.Elem())
		val.Set(v)
		typ = typ.Elem()
		val = val.Elem()
	}
	return root, val
}

func map2map(dst, src reflect.Value, opt SetOption) error {
	valueType := dst.Type().Elem()
	keyType := dst.Type().Key()
	iter := src.MapRange()
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		k, kr := ptrValue(keyType)
		err := forceSet(kr, key.Interface(), opt, "")
		if err != nil {
			return err
		}
		v, vr := ptrValue(valueType)
		err = forceSet(vr, val.Interface(), opt, "")
		if err != nil {
			return err
		}
		dst.SetMapIndex(k.Elem(), v.Elem())
	}
	return nil
}

// dst slice
// src struct
func struct2slice(dst, src reflect.Value, opt SetOption) error {
	itemType := dst.Type().Elem()
	n := src.NumField()
	slice := reflect.MakeSlice(dst.Type(), 0, n)
	srcType := src.Type()
	for i := 0; i < n; i++ {
		field := src.Field(i)
		structField := srcType.Field(i)
		if structField.Anonymous {
			for field.Kind() == reflect.Ptr {
				if field.IsNil() {
					break
				}
				field = field.Elem()
			}
			if field.Type().Kind() == reflect.Struct {
				err := struct2slice(dst, field, opt)
				if err != nil {
					return err
				}
			}
			continue
		}
		if structField.PkgPath != "" {
			continue
		}
		value := reflect.New(itemType)
		err := setPtr(value.Elem(), field, opt)
		if err != nil {
			return nil
		}
		slice = reflect.Append(slice, value.Elem())
	}
	dst.Set(slice)
	return nil
}

// dst slice
// src map
func map2slice(dst, src reflect.Value, opt SetOption) error {
	if opt.MapToSliceOption == Pairs {
		return map2slice2(dst, src, opt)
	}
	valueType := dst.Type().Elem()

	keys := src.MapKeys()
	if len(keys) == 0 {
		return nil
	}
	var max int
	for _, key := range keys {
		var k int
		err := forceSet(reflect.ValueOf(&k), key.Interface(), opt, "")
		if err != nil {
			return err
		}
		if k > max {
			max = k
		}
	}

	var l = max + 1
	slice := reflect.MakeSlice(dst.Type(), l, l)

	iter := src.MapRange()
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		var k int
		err := forceSet(reflect.ValueOf(&k), key.Interface(), opt, "")
		if err != nil {
			return err
		}
		v, vr := ptrValue(valueType)
		err = forceSet(vr, val.Interface(), opt, "")
		if err != nil {
			return err
		}
		slice.Index(k).Set(v.Elem())
	}
	dst.Set(slice)
	return nil
}

// dst []pair
// src map[key]val
func map2slice2(dst, src reflect.Value, opt SetOption) error {
	elmType := dst.Type().Elem()
	elmStructType := elmType
	for elmStructType.Kind() == reflect.Ptr {
		elmStructType = elmStructType.Elem()
	}
	if elmStructType.Kind() != reflect.Struct {
		return errors.New("cannot convert map to slice of pair because slice's element type is not struct:" + elmType.String())
	}
	if elmStructType.NumField() < 2 {
		return errors.New("cannot convert map to slice of pair because slice's element numField less than 2:" + elmType.String())
	}
	kf := elmStructType.Field(0)
	if kf.Anonymous || kf.PkgPath != "" {
		return errors.New("pair struct invalid")
	}
	vf := elmStructType.Field(1)
	if vf.Anonymous || vf.PkgPath != "" {
		return errors.New("pair struct invalid")
	}
	l := src.Len()
	slice := reflect.MakeSlice(dst.Type(), l, l)
	iter := src.MapRange()
	var i int
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		root, val := ptrValue(elmType)
		err := forceSet(val.Field(0), k.Interface(), opt, "")
		if err != nil {
			return err
		}
		err = forceSet(val.Field(1), v.Interface(), opt, "")
		if err != nil {
			return err
		}
		slice.Index(i).Set(root.Elem())
		i++
	}
	dst.Set(slice)
	return nil
}

func toBytes(i interface{}, opt SetOption) ([]byte, error) {
	switch o := i.(type) {
	case []byte:
		return o, nil
	case string:
		switch opt.BytesOption {
		case Base64:
			data, err := base64.StdEncoding.DecodeString(o)
			if err != nil {
				return nil, err
			}
			return data, nil
		default:
			return []byte(o), nil
		}
	}
	if opt.BytesOption == AsString {
		return []byte(toString(i, opt)), nil
	}
	switch i.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, uintptr:
		buffer := bytes.NewBuffer(nil)
		_ = binary.Write(buffer, binary.LittleEndian, i)
		return buffer.Bytes(), nil
	}
	return nil, errors.New("type (" + reflect.TypeOf(i).String() + ") to []byte invalid")
}

func toBool(i interface{}, opt SetOption) (bool, error) {
	if i == nil {
		return false, nil
	}
	switch o := i.(type) {
	case bool:
		return o, nil
	case int:
		return o != 0, nil
	case int8:
		return o != 0, nil
	case int16:
		return o != 0, nil
	case int32:
		return o != 0, nil
	case int64:
		return o != 0, nil

	case uint:
		return o != 0, nil
	case uint8:
		return o != 0, nil
	case uint16:
		return o != 0, nil
	case uint32:
		return o != 0, nil
	case uint64:
		return o != 0, nil
	case float32:
		return o != 0, nil
	case float64:
		return o != 0, nil
	case string:
		switch strings.ToLower(o) {
		case "true", "1":
			return true, nil
		case "", "false", "0", "null", "nil":
			return false, nil
		default:
			return true, nil
		}
	case []byte:
		return strconv.ParseBool(string(o))
	}
	return false, errors.New("type (" + reflect.TypeOf(i).String() + ") to bool invalid")
}

func toInt(i interface{}, opt SetOption) (int64, error) {
	switch o := i.(type) {
	case int:
		return int64(o), nil
	case int8:
		return int64(o), nil
	case int16:
		return int64(o), nil
	case int32:
		return int64(o), nil
	case int64:
		return int64(o), nil

	case uint:
		return int64(o), nil
	case uint8:
		return int64(o), nil
	case uint16:
		return int64(o), nil
	case uint32:
		return int64(o), nil
	case uint64:
		return int64(o), nil
	case []byte:
		switch opt.BytesOption {
		case AsString:
			return strconv.ParseInt(string(o), 10, 0)
		case Binary:
			i64, _ := binary.Varint(o)
			return i64, nil
		}
	case string:
		return strconv.ParseInt(string(o), 10, 0)
	case bool:
		if o {
			return 1, nil
		} else {
			return 0, nil
		}
	case float64:
		return int64(o), nil
	case float32:
		return int64(o), nil
	case json.Number:
		return o.Int64()
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toFloat(i interface{}, opt SetOption) (float64, error) {
	switch o := i.(type) {
	case int:
		return float64(o), nil
	case int8:
		return float64(o), nil
	case int16:
		return float64(o), nil
	case int32:
		return float64(o), nil
	case int64:
		return float64(o), nil

	case uint:
		return float64(o), nil
	case uint8:
		return float64(o), nil
	case uint16:
		return float64(o), nil
	case uint32:
		return float64(o), nil
	case uint64:
		return float64(o), nil

	case []byte:
		return strconv.ParseFloat(string(o), 64)

	case string:
		return strconv.ParseFloat(string(o), 64)
	case bool:
		if o {
			return 1, nil
		} else {
			return 0, nil
		}
	case float64:
		return o, nil
	case float32:
		return float64(o), nil
	case json.Number:
		return o.Float64()
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toUint(i interface{}, opt SetOption) (uint64, error) {
	switch o := i.(type) {

	case int:
		return uint64(o), nil
	case int8:
		return uint64(o), nil
	case int16:
		return uint64(o), nil
	case int32:
		return uint64(o), nil
	case int64:
		return uint64(o), nil

	case uint:
		return uint64(o), nil
	case uint8:
		return uint64(o), nil
	case uint16:
		return uint64(o), nil
	case uint32:
		return uint64(o), nil
	case uint64:
		return o, nil

	case []byte:
		switch opt.BytesOption {
		case AsString:
			return strconv.ParseUint(string(o), 10, 0)
		case Binary:
			i64, _ := binary.Uvarint(o)
			return i64, nil
		}
	case string:
		return strconv.ParseUint(string(o), 10, 0)
	case bool:
		if o {
			return 1, nil
		} else {
			return 0, nil
		}
	case float64:
		return uint64(o), nil
	case float32:
		return uint64(o), nil
	case json.Number:
		ii, err := o.Int64()
		return uint64(ii), err
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toString(i interface{}, opt SetOption) string {
	switch o := i.(type) {
	case string:
		return o
	case int:
		return strconv.FormatInt(int64(o), 10)
	case int8:
		return strconv.FormatInt(int64(o), 10)
	case int16:
		return strconv.FormatInt(int64(o), 10)
	case int32:
		return strconv.FormatInt(int64(o), 10)
	case int64:
		return strconv.FormatInt(o, 10)

	case uint:
		return strconv.FormatUint(uint64(o), 10)
	case uint8:
		return strconv.FormatUint(uint64(o), 10)
	case uint16:
		return strconv.FormatUint(uint64(o), 10)
	case uint32:
		return strconv.FormatUint(uint64(o), 10)
	case uint64:
		return strconv.FormatUint(o, 10)
	case uintptr:
		return strconv.FormatUint(uint64(o), 10)
	case bool:
		return strconv.FormatBool(o)
	case float32:
		return strconv.FormatFloat(float64(o), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(o, 'f', -1, 64)

	case []byte:
		switch opt.BytesOption {
		case Base64:
			return base64.StdEncoding.EncodeToString(o)
		}
		return string(o)
	case fmt.Stringer:
		return o.String()
	}
	return fmt.Sprint(i)
}
