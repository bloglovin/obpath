package main

import (
	"fmt"
	"reflect"
)

// FloatCast converts all non complex numbers to float64s
func FloatCast(n interface{}) (error, float64) {
	k := reflect.TypeOf(n).Kind()

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil, float64(reflect.ValueOf(n).Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil, float64(reflect.ValueOf(n).Uint())
	case reflect.Float32, reflect.Float64:
		return nil, reflect.ValueOf(n).Float()
	}

	return fmt.Errorf("wrong kind of value: %v", k.String()), 0.0
}
