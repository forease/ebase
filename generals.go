package ebase

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type KeyValue struct {
	Key   interface{}
	Value interface{}
}

type Generals map[string]interface{}

func NewSettings() Generals {
	return Generals{}
}

func (set Generals) Get(name string, value interface{}) (err error) {
	tmp, ok := set[name]
	if !ok {
		return fmt.Errorf("keys not found.[%s]", name)
	}

	var v interface{}

	tmpType := getTypeOf(tmp)
	valueType := getTypeOf(value)

	//
	if tmpType == valueType {
		refValue := reflect.Indirect(reflect.ValueOf(value))
		refValue.Set(reflect.Indirect(reflect.ValueOf(tmp)))
		return
	}

	if tmpType == "string" {
		switch valueType {
		case "struct", "map": // for struct and map, using json.Unmarshal
			return json.Unmarshal([]byte(tmp.(string)), value)
		case "int":
			v, err = strconv.Atoi(tmp.(string))
		case "string":
			v = tmp
		case "bool":
			var a int
			a, err = strconv.Atoi(tmp.(string))
			if err != nil {
				return
			}
			if a == 0 {
				v = false
			} else {
				v = true
			}

		default:
			err = fmt.Errorf("unspport value type[%s]", valueType)
		}
	} else {
		err = fmt.Errorf("unspport source type[%s]", tmpType)
	}

	// 给传递进来的参数赋值
	if err == nil {
		refValue := reflect.Indirect(reflect.ValueOf(value))
		refValue.Set(reflect.Indirect(reflect.ValueOf(v)))
	}

	return
}

func getTypeOf(val interface{}) (typeName string) {

	tp := reflect.TypeOf(val)
	typeName = tp.Kind().String()

	if typeName == "ptr" {
		typeName = tp.Elem().Kind().String()
	}

	return
}
