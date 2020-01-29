package forceset

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
}

type Option func(opt *SetOption)

func MapAsPairs(opt *SetOption) {
	opt.MapToSliceOption = Pairs
}

func MapAsArrayLike(opt *SetOption) {
	opt.MapToSliceOption = ArrayLike
}
