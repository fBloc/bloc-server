package value_object

import "github.com/spf13/cast"

type RepositoryFilter struct {
	equal               map[string]interface{}
	notEqual            map[string]interface{}
	strValueContains    map[string]string
	strValueStartsWith  map[string]string
	strValueEndsWith    map[string]string
	strValueNotContains map[string]string
	valueIn             map[string][]interface{}
	valueNotIn          map[string][]interface{}
	gt                  map[string]interface{}
	gte                 map[string]interface{}
	lt                  map[string]interface{}
	lte                 map[string]interface{}
	fieldExist          []string
	fieldNotExist       []string
	fieldValueNotNull   []string
}

func NewRepositoryFilter() *RepositoryFilter {
	return &RepositoryFilter{}
}

func (rf *RepositoryFilter) GetGt() map[string]interface{} {
	return rf.gt
}

func (rf *RepositoryFilter) AddGt(key string, val interface{}) *RepositoryFilter {
	if rf.gt == nil {
		rf.gt = make(map[string]interface{})
	}
	rf.gt[key] = val
	return rf
}

func (rf *RepositoryFilter) GetGte() map[string]interface{} {
	return rf.gte
}

func (rf *RepositoryFilter) AddGte(key string, val interface{}) *RepositoryFilter {
	if rf.gte == nil {
		rf.gte = make(map[string]interface{})
	}
	rf.gte[key] = val
	return rf
}

func (rf *RepositoryFilter) GetLt() map[string]interface{} {
	return rf.lt
}

func (rf *RepositoryFilter) AddLt(key string, val interface{}) *RepositoryFilter {
	if rf.lt == nil {
		rf.lt = make(map[string]interface{})
	}
	rf.lt[key] = val
	return rf
}

func (rf *RepositoryFilter) GetLte() map[string]interface{} {
	return rf.lte
}

func (rf *RepositoryFilter) AddLte(key string, val interface{}) *RepositoryFilter {
	if rf.lte == nil {
		rf.lte = make(map[string]interface{})
	}
	rf.lte[key] = val
	return rf
}

func (rf *RepositoryFilter) GetEqual() map[string]interface{} {
	return rf.equal
}

func (rf *RepositoryFilter) AddEqual(key string, val interface{}) *RepositoryFilter {
	if rf.equal == nil {
		rf.equal = make(map[string]interface{})
	}
	rf.equal[key] = val
	return rf
}

func (rf *RepositoryFilter) GetStrValueStartsWith() map[string]string {
	return rf.strValueStartsWith
}

func (rf *RepositoryFilter) AddStrValueStartsWith(key string, val interface{}) *RepositoryFilter {
	if rf.strValueStartsWith == nil {
		rf.strValueStartsWith = make(map[string]string)
	}
	strVal := cast.ToString(val)
	rf.strValueStartsWith[key] = strVal
	return rf
}

func (rf *RepositoryFilter) GetStrValueEndsWith() map[string]string {
	return rf.strValueEndsWith
}

func (rf *RepositoryFilter) AddStrValueEndsWith(key string, val interface{}) *RepositoryFilter {
	if rf.strValueEndsWith == nil {
		rf.strValueEndsWith = make(map[string]string)
	}
	strVal := cast.ToString(val)
	rf.strValueEndsWith[key] = strVal
	return rf
}

func (rf *RepositoryFilter) GetNotEqual() map[string]interface{} {
	return rf.notEqual
}

func (rf *RepositoryFilter) AddNotEqual(key string, val interface{}) *RepositoryFilter {
	if rf.notEqual == nil {
		rf.notEqual = make(map[string]interface{})
	}
	rf.notEqual[key] = val
	return rf
}

func (rf *RepositoryFilter) AddNotNull(key string) *RepositoryFilter {
	if rf.fieldValueNotNull == nil {
		rf.fieldValueNotNull = make([]string, 0)
	}
	rf.fieldValueNotNull = append(rf.fieldValueNotNull, key)
	return rf
}

func (rf *RepositoryFilter) AddNotExist(key string) *RepositoryFilter {
	if rf.fieldNotExist == nil {
		rf.fieldNotExist = make([]string, 0)
	}
	rf.fieldNotExist = append(rf.fieldNotExist, key)
	return rf
}

func (rf *RepositoryFilter) GetStrContains() map[string]string {
	return rf.strValueContains
}

func (rf *RepositoryFilter) AddContains(key string, val interface{}) *RepositoryFilter {
	if rf.strValueContains == nil {
		rf.strValueContains = make(map[string]string)
	}
	strVal := cast.ToString(val)
	rf.strValueContains[key] = strVal
	return rf
}

func (rf *RepositoryFilter) AddNotContains(key string, val interface{}) *RepositoryFilter {
	if rf.strValueNotContains == nil {
		rf.strValueNotContains = make(map[string]string)
	}
	strVal := cast.ToString(val)
	rf.strValueNotContains[key] = strVal
	return rf
}

func (rf *RepositoryFilter) GetIn() map[string][]interface{} {
	return rf.valueIn
}

func (rf *RepositoryFilter) AddIn(key string, val interface{}) *RepositoryFilter {
	if rf.valueIn == nil {
		rf.valueIn = make(map[string][]interface{})
	}
	valSlice := cast.ToSlice(val)
	rf.valueIn[key] = valSlice
	return rf
}

func (rf *RepositoryFilter) AddNotIn(key string, val interface{}) *RepositoryFilter {
	if rf.valueNotIn == nil {
		rf.valueNotIn = make(map[string][]interface{})
	}
	valSlice := cast.ToSlice(val)
	rf.valueNotIn[key] = valSlice
	return rf
}

func (rf *RepositoryFilter) GetFiledExist() []string {
	return rf.fieldExist
}

func (rf *RepositoryFilter) GetFiledNotExist() []string {
	return rf.fieldNotExist
}

func (rf *RepositoryFilter) AddExist(field string) *RepositoryFilter {
	if rf.fieldExist == nil {
		rf.fieldExist = make([]string, 0)
	}
	rf.fieldExist = append(rf.fieldExist, field)
	return rf
}
