package value_object

type FilterSortType int

const (
	UnknownFilterType FilterSortType = iota
	Asc
	Desc
)

func (fST FilterSortType) String() string {
	switch fST {
	case Asc:
		return "asc"
	case Desc:
		return "desc"
	default:
		return "unknown"
	}
}
