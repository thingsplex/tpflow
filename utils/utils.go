package utils

import (
	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	"strings"
)

func GenerateId(len int) string {
	return uniuri.NewLen(len)
}

func ConfigValueToNumber(valueType string,value interface{})(float64,error){
	if valueType == "int" {
		switch val := value.(type) {
		case int64 :
			return float64(val),nil
		case float64:
			return val,nil
		default:
			return 0, errors.New("Can't convert interface{} to int64")

		}
	}else
	if valueType == "float" {
		floatVal,ok := value.(float64)
		if ok {
			return floatVal , nil
		}else {
			return 0, errors.New("Can't convert interface{} to float64")
		}
	}
	return 0,errors.New("Not numeric value type")
}



func match(route []string, topic []string) bool {
	if len(route) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if route[0] == "#" {
			return true
		}
		return false
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return match(route[1:], topic[1:])
	}

	return false
}

func RouteIncludesTopic(route, topic string) bool {
	return match(strings.Split(route, "/"), strings.Split(topic, "/"))
}





