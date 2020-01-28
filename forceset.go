package forceset

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ForceSet(value reflect.Value, i interface{}, opts ...option) error {
	var opt option
	if len(opts) != 0 {
		opt = opts[0]
	}
	var bErr error
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
	o := reflect.ValueOf(i)
	if o.Type() == value.Type() {
		value.Set(o)
		return nil
	}

	if value.Type().Kind() == reflect.Interface {
		if o.Type().Implements(value.Type()) {
			value.Set(o)
			return nil
		}
	}

	if o.Type().ConvertibleTo(value.Type()) {
		converted := o.Convert(value.Type())
		value.Set(converted)
		return nil
	}

	if o.Type().AssignableTo(value.Type()) {
		value.Set(o)
		return nil
	}

	if bErr != nil {
		return bErr
	}

	if value.Kind() == reflect.Slice {
		iv := reflect.ValueOf(i)
		if iv.Type().Kind() == reflect.Slice || iv.Type().Kind() == reflect.Array {
			proxyValue := reflect.MakeSlice(value.Type(), iv.Len(), iv.Len())
			size := iv.Len()
			for n := 0; n < size; n++ {
				elm := iv.Index(n)
				err := ForceSet(proxyValue.Index(n), elm.Interface(), opt)
				if err != nil {
					return err
				}
			}
			value.Set(proxyValue)
			return nil
		}
	}
	return errors.New("force set type(" + o.Type().String() + ") into type(" + value.Type().String() + ") failed")
}

type BytesOption int8

const (
	AsString BytesOption = iota
	Base64
	Binary
)

type option struct {
	BytesOption BytesOption
}

func toBytes(i interface{}, opt option) ([]byte, error) {
	switch o := i.(type) {
	case []byte:
		return o, nil
	case string:
		switch opt.BytesOption {
		case Base64:
			bytes, err := base64.StdEncoding.DecodeString(o)
			if err != nil {
				return nil, err
			}
			return bytes, nil
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

func toBool(i interface{}, opt option) (bool, error) {
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

func toInt(i interface{}, opt option) (int64, error) {
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
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toFloat(i interface{}, opt option) (float64, error) {
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
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toUint(i interface{}, opt option) (uint64, error) {
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
	}
	return 0, errors.New("type (" + reflect.TypeOf(i).String() + ") to int invalid")
}

func toString(i interface{}, opt option) string {
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
