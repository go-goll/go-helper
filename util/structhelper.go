package util

import (
	"reflect"
	"strconv"
)

// map转结构体,key指结构体的tag名，目前仅支持内置类型转换
func HashToStruct(key string, data map[string]string, target interface{}) (err error) {
	_value := reflect.ValueOf(target).Elem()
	_type := reflect.TypeOf(target).Elem()

	for i := 0; i < _value.NumField(); i++ {
		// 获取tag
		tag := _type.Field(i).Tag.Get(key)
		// 获取结构体类型
		structType := _value.Field(i).Type().Kind().String()
		// 参数不存在跳过
		if dataVal, ok := data[tag]; ok {
			switch structType {
			case "string":
				_value.Field(i).Set(reflect.ValueOf(dataVal))
			case "uint8":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int8(tmpVal)))
			case "int32":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int32(tmpVal)))
			case "uint32":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int64(tmpVal)))
			case "int":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int(tmpVal)))
			case "int64":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int64(tmpVal)))
			case "uint64":
				tmpVal, err := strconv.Atoi(dataVal)
				if err != nil {
					return err
				}
				_value.Field(i).Set(reflect.ValueOf(int64(tmpVal)))
			default:

			}
		}
	}
	return
}
