package helper

import (
	"bitbucket.org/noon-micro/curriculum/pkg/lib/logger"
	"encoding/json"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"reflect"
)

var validate *validator.Validate

func InitializeValidator() {
	validate = validator.New()
	_ = validate.RegisterValidation("contains-nil", validateNil)
}

func validateNil(fl validator.FieldLevel) bool {
	field, _, _ := fl.ExtractType(fl.Field())
	if field.Kind() != reflect.Slice {
		return true
	}
	for i := 0; i < field.Len(); i++ {
		if field.Index(i).IsNil() {
			return false
		}
	}
	return true
}

func Validate(i interface{}) error {
	if i == nil {
		logger.Client.Info("nil")
		return nil
	}
	err := validate.Struct(i)
	return err
}

func JsonDecoder(body io.Reader, i interface{}) error {
	if i == nil {
		logger.Client.Info("nil")
		return nil
	}
	// decode request body
	decoder := json.NewDecoder(body)
	return decoder.Decode(i)
}
