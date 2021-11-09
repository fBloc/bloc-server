package ipt

import (
	"fmt"
	"strconv"

	"github.com/fBloc/bloc-backend-go/pkg/value_type"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type SelectOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// IptComponent 每个入参由一/多个IptComponent构成
type IptComponent struct {
	ValueType       value_type.ValueType         `json:"value_type"`
	FormControlType value_object.FormControlType `json:"formcontrol_type"`
	Hint            string                       `json:"hint"`
	DefaultValue    interface{}                  `json:"default_value"`
	AllowMulti      bool                         `json:"allow_multi"`
	SelectOptions   []SelectOption               `json:"select_options"` // only exist when FormControlType is selection
	Value           interface{}                  `json:"-"`
}

func (ipt *IptComponent) String() string {
	resp := string(ipt.ValueType) + string(ipt.FormControlType) + strconv.FormatBool(ipt.AllowMulti)
	for _, option := range ipt.SelectOptions {
		resp = resp + fmt.Sprintf("%v", option.Value)
	}
	return resp
}

func (ipt *IptComponent) Config() map[string]interface{} {
	config := make(map[string]interface{}, 6)
	config["value_type"] = ipt.ValueType
	config["formControl_type"] = ipt.FormControlType
	config["hint"] = ipt.Hint
	config["value"] = ipt.DefaultValue
	config["allow_multi"] = ipt.AllowMulti
	config["options"] = ipt.SelectOptions
	return config
}
