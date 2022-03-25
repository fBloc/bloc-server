package filter_options

import "strconv"

type FilterOption struct {
	Limit          int64
	OffSet         int64
	SortAscFields  []string
	SortDescFields []string
	NaturalAsc     *bool
	NaturalDesc    *bool
}

func NewFilterOption() *FilterOption {
	return &FilterOption{}
}

func (fo *FilterOption) SetLimit(val string) {
	intVar, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	if intVar > 0 {
		fo.Limit = int64(intVar)
	}
}

func (fo *FilterOption) SetOffset(val string) {
	intVar, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	if intVar > 0 {
		fo.OffSet = int64(intVar)
	}
}

func (fo *FilterOption) SetSortByNaturalAsc() *FilterOption {
	aTrue := true
	fo.NaturalAsc = &aTrue
	return fo
}

func (fo *FilterOption) SetSortByNaturalDesc() *FilterOption {
	aTrue := true
	fo.NaturalDesc = &aTrue
	return fo
}
