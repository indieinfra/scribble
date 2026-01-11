package config

import (
	"path"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

func ValidateAbsPath(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	return s != "" && path.IsAbs(s)
}

func ValidateLocalpath(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	return s != "" && filepath.IsLocal(s)
}
