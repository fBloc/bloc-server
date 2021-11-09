// @Title  opt
// @Description 此为bloc的每个输出字段的描述。一个输出 = key + value type + description
package opt

import (
	"strconv"

	"github.com/fBloc/bloc-backend-go/pkg/value_type"
)

type Opt struct {
	Key         string               `json:"key"`
	Description string               `json:"description"`
	ValueType   value_type.ValueType `json:"value_type"`
	IsArray     bool                 `json:"is_array"`
}

func (opt *Opt) String() string {
	return opt.Key + opt.Description + string(opt.ValueType) + strconv.FormatBool(opt.IsArray)
}

func (opt *Opt) Config() map[string]interface{} {
	config := make(map[string]interface{}, 4)
	config["key"] = opt.Key
	config["value_type"] = opt.ValueType
	config["description"] = opt.Description
	config["is_array"] = opt.IsArray
	return config
}
