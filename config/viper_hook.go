package config

import (
	"reflect"
	"strconv"

	"go.uber.org/zap"
)

func zapLevelHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if to != reflect.TypeOf(zap.AtomicLevel{}) {
		return data, nil
	}
	s, ok := data.(string)
	if !ok {
		return data, nil
	}
	var lvl zap.AtomicLevel
	if err := lvl.UnmarshalText([]byte(s)); err != nil {
		return nil, err
	}
	return lvl, nil
}

func stringToIntHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Int {
		return data, nil
	}
	return strconv.Atoi(data.(string))
}
