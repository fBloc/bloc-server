package value_object

type RepositoryFilterOption struct {
	Limit  int64
	OffSet int64
	Asc    bool
	Desc   bool
}

func NewRepositoryFilterOption() *RepositoryFilterOption {
	return &RepositoryFilterOption{}
}

func (fo *RepositoryFilterOption) SetLimit(val int) {
	fo.Limit = int64(val)
}

func (fo *RepositoryFilterOption) SetOffset(val int) {
	fo.OffSet = int64(val)
}

func (fo *RepositoryFilterOption) SetAsc() {
	fo.Asc = true
}

func (fo *RepositoryFilterOption) SetDesc() {
	fo.Desc = true
}
