package ipt

import (
	"strconv"

	"github.com/fBloc/bloc-server/pkg/value_type"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IptSlice []*Ipt

func (iS *IptSlice) iptIndexValid(iptIndex int) error {
	if len(*iS) < iptIndex-1 {
		return errors.New("iptIndex out of range")
	}
	return nil
}

func (iS *IptSlice) GetIntValue(iptIndex int, componentIndex int) (int, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return 0, err
	}
	return (*iS)[iptIndex].GetIntValue(componentIndex)
}

func (iS *IptSlice) GetIntSliceValue(iptIndex int, componentIndex int) ([]int, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []int{0}, err
	}
	return (*iS)[iptIndex].GetIntSliceValue(componentIndex)
}

func (iS *IptSlice) GetFloat64Value(iptIndex int, componentIndex int) (float64, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return 0, err
	}
	return (*iS)[iptIndex].GetFloat64Value(componentIndex)
}

func (iS *IptSlice) GetFloat64SliceValue(iptIndex int, componentIndex int) ([]float64, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []float64{}, err
	}
	return (*iS)[iptIndex].GetFloat64SliceValue(componentIndex)
}

func (iS *IptSlice) GetStringValue(iptIndex int, componentIndex int) (string, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return "", err
	}
	return (*iS)[iptIndex].GetStringValue(componentIndex)
}

func (iS *IptSlice) GetStringSliceValue(iptIndex int, componentIndex int) ([]string, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []string{}, err
	}
	return (*iS)[iptIndex].GetStringSliceValue(componentIndex)
}

func (iS *IptSlice) GetBoolValue(iptIndex int, componentIndex int) (bool, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return false, err
	}
	return (*iS)[iptIndex].GetBoolValue(componentIndex)
}

func (iS *IptSlice) GetBoolSliceValue(iptIndex int, componentIndex int) ([]bool, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []bool{}, err
	}
	return (*iS)[iptIndex].GetBoolSliceValue(componentIndex)
}

func (iS *IptSlice) GetJsonStrMapValue(iptIndex int, componentIndex int) (map[string]interface{}, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return (*iS)[iptIndex].GetJsonStrMapValue(componentIndex)
}

type Ipt struct {
	Key        string          `json:"key"`
	Display    string          `json:"display"`
	Must       bool            `json:"must"`
	Components []*IptComponent `json:"components"`
}

func (ipt *Ipt) String() string {
	resp := ipt.Key + ipt.Display + strconv.FormatBool(ipt.Must)
	for _, component := range ipt.Components {
		resp += component.String()
	}
	return resp
}

func (ipt *Ipt) Config() map[string]interface{} {
	config := make(map[string]interface{}, 4)
	config["key"] = ipt.Key
	config["display"] = ipt.Display
	config["must"] = ipt.Must

	fieldsConfigs := make([]map[string]interface{}, 0, len(ipt.Components))
	for _, field := range ipt.Components {
		fieldsConfigs = append(fieldsConfigs, field.Config())
	}
	config["components"] = fieldsConfigs
	return config
}

func (ipt *Ipt) GetIntValue(componentIndex int) (int, error) {
	if len(ipt.Components) < componentIndex-1 {
		return 0, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.IntValueType && component.ValueType != value_type.FloatValueType {
		return 0, errors.Errorf("valueType should be int/float but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return 0, errors.New("value should be a intSlice")
	}
	return cast.ToInt(component.Value), nil
}

func (ipt *Ipt) GetIntSliceValue(componentIndex int) (resp []int, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []int{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.IntValueType {
		return []int{}, errors.Errorf("valueType should be int but get: %s", component.ValueType)
	}
	if pa, ok := component.Value.(primitive.A); ok {
		valueMSI := []interface{}(pa)
		resp, err = cast.ToIntSliceE(valueMSI)
	} else {
		resp, err = cast.ToIntSliceE(component.Value)
	}
	return
}

func (ipt *Ipt) GetFloat64Value(componentIndex int) (float64, error) {
	if len(ipt.Components) < componentIndex-1 {
		return 0, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.FloatValueType {
		return 0, errors.Errorf("valueType should be float but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return 0, errors.New("value should be a intSlice")
	}
	return cast.ToFloat64(component.Value), nil
}

func (ipt *Ipt) GetFloat64SliceValue(componentIndex int) ([]float64, error) {
	if len(ipt.Components) < componentIndex-1 {
		return []float64{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.FloatValueType {
		return []float64{}, errors.Errorf("valueType should be float but get: %s", component.ValueType)
	}
	interfaceSlice := cast.ToSlice(component.Value)
	resp := make([]float64, 0, len(interfaceSlice))
	for _, i := range interfaceSlice {
		resp = append(resp, cast.ToFloat64(i))
	}
	return resp, nil
}

func (ipt *Ipt) GetStringValue(componentIndex int) (string, error) {
	if len(ipt.Components) < componentIndex-1 {
		return "", errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.StringValueType {
		return "", errors.Errorf("valueType should be string but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return "", errors.New("value should be a stringSlice")
	}
	return cast.ToString(component.Value), nil
}

func (ipt *Ipt) GetStringSliceValue(componentIndex int) (resp []string, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []string{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.StringValueType {
		return []string{}, errors.Errorf("valueType should be string but get: %s", component.ValueType)
	}
	if !component.AllowMulti {
		return []string{cast.ToString(component.Value)}, nil
	}
	if pa, ok := component.Value.(primitive.A); ok {
		valueMSI := []interface{}(pa)
		resp, err = cast.ToStringSliceE(valueMSI)
	} else {
		resp, err = cast.ToStringSliceE(component.Value)
	}
	return
}

func (ipt *Ipt) GetBoolValue(componentIndex int) (bool, error) {
	if len(ipt.Components) < componentIndex-1 {
		return false, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.BoolValueType {
		return false, errors.Errorf("valueType should be bool but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return false, errors.New("value should be a stringSlice")
	}
	return cast.ToBool(component.Value), nil
}

func (ipt *Ipt) GetBoolSliceValue(componentIndex int) (resp []bool, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []bool{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.StringValueType {
		return []bool{}, errors.Errorf("valueType should be bool but get: %s", component.ValueType)
	}
	if pa, ok := component.Value.(primitive.A); ok {
		valueMSI := []interface{}(pa)
		resp, err = cast.ToBoolSliceE(valueMSI)
	} else {
		resp, err = cast.ToBoolSliceE(component.Value)
	}
	return
}

func (ipt *Ipt) GetJsonStrMapValue(componentIndex int) (map[string]interface{}, error) {
	if len(ipt.Components) < componentIndex-1 {
		return map[string]interface{}{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != value_type.JsonValueType {
		return map[string]interface{}{}, errors.Errorf("valueType should be json but get: %s", component.ValueType)
	}
	if !component.AllowMulti {
		return map[string]interface{}{}, nil
	}
	return cast.ToStringMap(component.Value), nil
}
