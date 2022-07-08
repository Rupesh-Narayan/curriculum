package helper

import (
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// PrettyPrint to dump struct value
func PrettyPrint(i interface{}) {
	if i == nil {
		logger.Client.Info("nil")
	}
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		panic(err)
	}
	logger.Client.Info(string(s))
}

// GetIntIDFromIntefaceValue transform string or int or float of various type to int
func GetIntIDFromIntefaceValue(roleInterface interface{}) int {
	var id int
	dataType := reflect.TypeOf(roleInterface).Kind()

	switch dataType {
	case reflect.String:
		id, _ = strconv.Atoi(roleInterface.(string))
	case reflect.Int32:
		id = int(roleInterface.(int32))
	case reflect.Int64:
		id = int(roleInterface.(int64))
	case reflect.Float32:
		id = int(roleInterface.(float32))
	case reflect.Float64:
		id = int(roleInterface.(float64))
	default:
	}
	return id
}

// Contains data
func Contains(data []int, datum int) bool {
	for _, a := range data {
		if a == datum {
			return true
		}
	}
	return false
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
