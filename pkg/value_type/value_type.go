package value_type

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/cast"
)

// ValueType 此控件输入值的类型
type ValueType string

const (
	IntValueType    ValueType = "int"
	FloatValueType  ValueType = "float"
	StringValueType ValueType = "string"
	BoolValueType   ValueType = "bool"
	JsonValueType   ValueType = "json"
)

func CheckValueTypeValueValid(valueType ValueType, value interface{}) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	if valueType == IntValueType {
		return validIntValueType(value)
	} else if valueType == FloatValueType {
		return validFloatValueType(value)
	} else if valueType == StringValueType {
		return validStringValueType(value)
	} else if valueType == BoolValueType {
		return validBoolValueType(value)
	} else if valueType == JsonValueType {
		return validJsonValueType(value)
	}
	return false
}

func validIntValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		_, err = cast.ToIntSliceE(val)
	case reflect.String:
		err = errors.New("need int get string")
	case reflect.Bool:
		err = errors.New("need int get bool")
	case reflect.Map:
		err = errors.New("need int get map")
	default:
		_, err = cast.ToIntE(val)
	}
	return err == nil
}

func validFloatValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(val)
		for i := 0; i < s.Len(); i++ {
			_, tmpE := cast.ToFloat64E(s.Index(i).Interface())
			if tmpE != nil {
				err = tmpE
				break
			}
		}
	case reflect.String:
		err = errors.New("need float get string")
	case reflect.Bool:
		err = errors.New("need float get bool")
	case reflect.Map:
		err = errors.New("need float get map")
	default:
		_, err = cast.ToFloat64E(val)
	}
	return err == nil
}

func validStringValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		_, err = cast.ToStringSliceE(val)
	case reflect.Float64:
		err = errors.New("need string get float64")
	case reflect.Int:
		err = errors.New("need string get int")
	case reflect.Bool:
		err = errors.New("need string get bool")
	case reflect.Map:
		err = errors.New("need string get map")
	default:
		_, err = cast.ToStringE(val)
	}
	return err == nil
}

func validBoolValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		_, err = cast.ToBoolSliceE(val)
	case reflect.Float64:
		err = errors.New("need bool get float64")
	case reflect.Int:
		err = errors.New("need bool get int")
	case reflect.Map:
		err = errors.New("need bool get map")
	case reflect.String:
		err = errors.New("need bool get string")
	default:
		_, err = cast.ToBoolE(val)
	}
	return err == nil
}

func validJsonValueType(val interface{}) bool {
	valStr, ok := val.(string)
	if !ok {
		return false
	}
	var js map[string]interface{}
	return json.Unmarshal([]byte(valStr), &js) == nil
}
