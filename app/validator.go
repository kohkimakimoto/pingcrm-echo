package app

import "github.com/kohkimakimoto/pingcrm-echo/pkg/validation"

type Validator struct {
	*validation.Validator
}

func NewValidator() *Validator {
	return &Validator{
		Validator: validation.NewValidator(),
	}
}

func (v *Validator) ErrorMessageMap() map[string]string {
	m := make(map[string]string)
	for err := v.Errors.First(); err != nil; err = err.Next() {
		m[err.Key] = err.Message
	}
	return m
}
