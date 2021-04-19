package forceset

import "reflect"

type BytesOption uint8

const (
	AsString BytesOption = iota
	Base64
	Binary
)

type MapToSliceOption uint8

const (
	ArrayLike MapToSliceOption = iota
	Pairs
)

type SetOption struct {
	BytesOption      BytesOption
	Tag              string
	MapToSliceOption MapToSliceOption
	Mappers          map[MapperType]Mapper
	Decoder          func([]byte, interface{}) error
}

type Mapper func(dst reflect.Value, src reflect.Value, tag string) error

type MapperType struct {
	Destination reflect.Type
	Source      reflect.Type
}

type Option func(opt *SetOption)

func MapAsPairs(opt *SetOption) {
	opt.MapToSliceOption = Pairs
}

func MapAsArrayLike(opt *SetOption) {
	opt.MapToSliceOption = ArrayLike
}
