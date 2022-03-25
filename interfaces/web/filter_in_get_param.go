package web

import (
	"errors"
	"net/url"
	"strings"

	"github.com/fBloc/bloc-server/value_object"
)

// FilterInGetPath
type FilterInGetPath int

const (
	FilterInGetPathUnknown FilterInGetPath = iota
	FilterInGetPathEq
	FilterInGetPathLt
	FilterInGetPathLte
	FilterInGetPathGt
	FilterInGetPathGte
	FilterInGetPathNotEq
	FilterInGetPathIn
	FilterInGetPathNotIn
	FilterInGetPathContains
	FilterInGetPathNotContains
	FilterInGetPathLimit
	FilterInGetPathOffset
	FilterInGetPathSort
	FilterInGetPathFilterFields
	FilterInGetPathFilterOutFields
	maxFilterInGetPath
)
const (
	DefaultFilterInGetPath = FilterInGetPathUnknown
)

// WebReqSuffix http前端get请求时候的后缀
func (r FilterInGetPath) WebReqSuffix() string {
	switch r {
	case FilterInGetPathEq:
		return ""
	case FilterInGetPathLt:
		return "__lt"
	case FilterInGetPathLte:
		return "__lte"
	case FilterInGetPathGt:
		return "__gt"
	case FilterInGetPathGte:
		return "__gte"
	case FilterInGetPathNotEq:
		return "__not"
	case FilterInGetPathIn:
		return "__in"
	case FilterInGetPathNotIn:
		return "__notin"
	case FilterInGetPathContains:
		return "__contains"
	case FilterInGetPathNotContains:
		return "__notcontains"
	case FilterInGetPathSort:
		return "sort"
	case FilterInGetPathLimit:
		return "limit"
	case FilterInGetPathOffset:
		return "offset"
	case FilterInGetPathFilterFields:
		return "fields_only"
	case FilterInGetPathFilterOutFields:
		return "fields_without"
	default:
		return "not valid"
	}
}

func (r FilterInGetPath) Value() int {
	return int(r)
}

func (r FilterInGetPath) IsValid() bool {
	if r > 0 && r < maxFilterInGetPath {
		return true
	}
	return false
}

func (r FilterInGetPath) String() string {
	return r.WebReqSuffix()
}

func (r FilterInGetPath) StringByValue(value int) string {
	res := FilterInGetPath(value)
	return res.String()
}

func (r FilterInGetPath) ItemsAmount() int {
	return int(maxFilterInGetPath) - 1
}

func ParseReqQueryToGroupedFilters(queryMap url.Values) (map[FilterInGetPath][]string, error) {
	allSuffixMapFilterInGetPath := make(map[string]FilterInGetPath, int(maxFilterInGetPath)-1)
	for i := 1; i < int(maxFilterInGetPath); i++ {
		suffix := FilterInGetPath(i).WebReqSuffix()
		if suffix != "" {
			allSuffixMapFilterInGetPath[suffix] = FilterInGetPath(i)
		}
	}

	ret := make(map[FilterInGetPath][]string, maxFilterInGetPath-1)
	var theFilterInGetPath FilterInGetPath
	for reqKey := range queryMap {
		val := queryMap.Get(reqKey)
		theFilterInGetPath = FilterInGetPathEq

		switch reqKey {
		case FilterInGetPathLimit.String():
			theFilterInGetPath = FilterInGetPathLimit
		case FilterInGetPathOffset.String():
			theFilterInGetPath = FilterInGetPathOffset
		case FilterInGetPathSort.String():
			if val != "asc" && val != "desc" {
				return nil, errors.New("sort field's value must be asc/desc")
			}
			theFilterInGetPath = FilterInGetPathSort
		case FilterInGetPathFilterFields.String():
			theFilterInGetPath = FilterInGetPathFilterFields
		case FilterInGetPathFilterOutFields.String():
			theFilterInGetPath = FilterInGetPathFilterOutFields
		default:
			for suffix, filterInGetPath := range allSuffixMapFilterInGetPath {
				if strings.HasSuffix(reqKey, suffix) {
					reqKey = strings.ReplaceAll(reqKey, suffix, "")
					theFilterInGetPath = filterInGetPath
					break
				}
			}
		}

		ret[theFilterInGetPath] = append(ret[theFilterInGetPath], reqKey)
	}
	return ret, nil
}

func (f FilterInGetPath) AddToRepositoryFilter(
	filter *value_object.RepositoryFilter, key string, val interface{},
) {
	switch f {
	case FilterInGetPathEq:
		filter.AddEqual(key, val)
	case FilterInGetPathLt:
		filter.AddLt(key, val)
	case FilterInGetPathLte:
		filter.AddLte(key, val)
	case FilterInGetPathGt:
		filter.AddGt(key, val)
	case FilterInGetPathGte:
		filter.AddGte(key, val)
	case FilterInGetPathNotEq:
		filter.AddNotEqual(key, val)
	case FilterInGetPathIn:
		filter.AddIn(key, val)
	case FilterInGetPathNotIn:
		filter.AddNotIn(key, val)
	case FilterInGetPathContains:
		filter.AddContains(key, val)
	case FilterInGetPathNotContains:
		filter.AddNotContains(key, val)
	}
}
