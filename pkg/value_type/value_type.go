package value_type

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		return ValidIntValueType(value)
	} else if valueType == FloatValueType {
		return ValidFloatValueType(value)
	} else if valueType == StringValueType {
		return ValidStringValueType(value)
	} else if valueType == BoolValueType {
		return ValidBoolValueType(value)
	} else if valueType == JsonValueType {
		return ValidJsonValueType(value)
	}
	return false
}

func ValidIntValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		if pa, ok := val.(primitive.A); ok {
			valueMSI := []interface{}(pa)
			_, err = cast.ToIntSliceE(valueMSI)
		} else {
			_, err = cast.ToIntSliceE(val)
		}
	default:
		_, err = cast.ToIntE(val)
	}
	return err == nil
}

func ValidFloatValueType(val interface{}) bool {
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
	default:
		_, err = cast.ToFloat64E(val)
	}
	return err == nil
}

func ValidStringValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		if pa, ok := val.(primitive.A); ok {
			valueMSI := []interface{}(pa)
			_, err = cast.ToStringSliceE(valueMSI)
		} else {
			_, err = cast.ToStringSliceE(val)
		}
	default:
		_, err = cast.ToStringE(val)
	}
	return err == nil
}

func ValidBoolValueType(val interface{}) bool {
	var err error
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		if pa, ok := val.(primitive.A); ok {
			valueMSI := []interface{}(pa)
			_, err = cast.ToBoolSliceE(valueMSI)
		} else {
			_, err = cast.ToBoolSliceE(val)
		}
	default:
		_, err = cast.ToBoolE(val)
	}
	return err == nil
}

func ValidJsonValueType(val interface{}) bool {
	valStr, ok := val.(string)
	if !ok {
		return false
	}
	var js map[string]interface{}
	return json.Unmarshal([]byte(valStr), &js) == nil
}
