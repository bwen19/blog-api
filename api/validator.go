package api

import "github.com/go-playground/validator/v10"

var validRole validator.Func = func(fl validator.FieldLevel) bool {
	if role, ok := fl.Field().Interface().(string); ok {
		switch role {
		case "ghost", "user", "author", "admin":
			return true
		}
	}
	return false
}

var validStatus validator.Func = func(fl validator.FieldLevel) bool {
	if status, ok := fl.Field().Interface().(string); ok {
		switch status {
		case "draft", "review", "published":
			return true
		}
	}
	return false
}
