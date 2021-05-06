package util

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type IValidator interface {
	Validate() error
}

func StructFromValue(tmpl interface{}, value interface{}) (interface{}, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("cannot encode value to JSON: %v", err)
	}
	return StructFromJSON(tmpl, jsonValue)
} //StructFromValue()

func StructFromJSONReader(tmpl interface{}, reader io.Reader) (interface{}, error) {
	structPtrValue, _, err := newStruct(tmpl)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(reader).Decode(structPtrValue.Interface()); err != nil {
		return nil, fmt.Errorf("cannot parse JSON into %T: %v", tmpl, err)
	}
	if validator, ok := structPtrValue.Interface().(IValidator); ok {
		if err := validator.Validate(); err != nil {
			return nil, fmt.Errorf("invalid: %v", err)
		}
	}
	return tmplStruct(tmpl, structPtrValue)
}

func StructFromJSON(tmpl interface{}, jsonData []byte) (interface{}, error) {
	structPtrValue, _, err := newStruct(tmpl)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, structPtrValue.Interface()); err != nil {
		return nil, fmt.Errorf("cannot parse JSON into %T: %v", tmpl, err)
	}
	if validator, ok := structPtrValue.Interface().(IValidator); ok {
		if err := validator.Validate(); err != nil {
			return nil, fmt.Errorf("invalid: %v", err)
		}
	}
	return tmplStruct(tmpl, structPtrValue)
} //StructFromJSON()

//in URL, values are always []string even when only one string or when int etc
//in other objects you may have float storing 1.0 and struct expects int
//this function tries to parse the values into the desired struct field types
//it also match map names to field name or json tag names in the struct
//but it does not parse nested data (yet)
func StructFromMap(tmpl interface{}, obj map[string]interface{}) (interface{}, error) {
	structPtrValue, structType, err := newStruct(tmpl)
	if err != nil {
		return nil, err
	}
	for paramName, paramValue := range obj {
		structField, ok := structType.FieldByNameFunc(func(fieldName string) bool {
			if paramName == fieldName {
				return true
			}
			structTypeField, _ := structType.FieldByName(fieldName)
			jsonTags := strings.SplitN(structTypeField.Tag.Get("json"), ",", 2)
			if len(jsonTags) > 0 && jsonTags[0] == paramName {
				return true
			}
			return false
		})
		if !ok {
			return nil, fmt.Errorf("unknown struct field %s.%s does not exist", structType.Name(), paramName)
		}

		//parse the value into the field
		fieldValue := structPtrValue.Elem().FieldByIndex(structField.Index)
		switch fieldValue.Type().Kind() {
		case reflect.String:
			fieldValue.Set(reflect.ValueOf(fmt.Sprintf("%v", paramValue)))
		case reflect.Int:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(int(intValue)))
		case reflect.Int8:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(int8(intValue)))
		case reflect.Uint8:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(uint8(intValue)))
		case reflect.Int16:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(int16(intValue)))
		case reflect.Uint16:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(uint16(intValue)))
		case reflect.Int32:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(int32(intValue)))
		case reflect.Uint32:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(uint32(intValue)))
		case reflect.Int64:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(int64(intValue)))
		case reflect.Uint64:
			intValue, err := strconv.ParseInt(fmt.Sprintf("%v", paramValue), 10, 8)
			if err != nil {
				return nil, fmt.Errorf("cannot parse int from %s=(%T)%v for %v", paramName, paramValue, paramValue, fieldValue.Type())
			}
			fieldValue.Set(reflect.ValueOf(uint64(intValue)))
		default:
			return nil, fmt.Errorf("cannot store %s=(%T)%v in %v", paramName, paramValue, paramValue, fieldValue.Type())
		}
	}
	return tmplStruct(tmpl, structPtrValue)
}

//params:
//	tmpl is your template struct or ptr to struct, whichever type you want returned
//	the data in tmpl will be copied before parsing, so can define default values
//	if tmpl implements IValidator, the parsed data will be validated
//return:
//	structPtrValue is value of &yourStruct
//	structType is type of struct, always with kind=struct
//	err only if failed
func newStruct(tmpl interface{}) (structPtrValue reflect.Value, structType reflect.Type, err error) {
	if tmpl == nil {
		return reflect.ValueOf(nil), reflect.TypeOf(nil), fmt.Errorf("tmpl is nil")
	}

	//allocate a new struct with the same type as tmpl
	tmplType := reflect.TypeOf(tmpl)
	structType = tmplType
	switch tmplType.Kind() {
	case reflect.Ptr:
		if tmplType.Elem().Kind() != reflect.Struct {
			return reflect.ValueOf(nil), reflect.TypeOf(nil), fmt.Errorf("%T is not a struct", tmpl)
		}
		structType = tmplType.Elem()
		structPtrValue = reflect.New(structType)
	case reflect.Struct:
		structPtrValue = reflect.New(structType)
	default:
		return reflect.ValueOf(nil), reflect.TypeOf(nil), fmt.Errorf("%T is not a struct", tmpl)
	}

	//copy value of tmpl to the new struct (as default field values) before we start parsing
	structPtrValue.Elem().Set(reflect.ValueOf(tmpl))
	return structPtrValue, structType, nil
} //newStruct()

//use this at the end of a function that started with newStruct() to return either
//the struct or &struct as requested by the type of tmpl
func tmplStruct(tmpl interface{}, structPtrValue reflect.Value) (interface{}, error) {
	if reflect.TypeOf(tmpl).Kind() == reflect.Ptr {
		return structPtrValue.Interface(), nil
	}
	return structPtrValue.Elem().Interface(), nil
}
