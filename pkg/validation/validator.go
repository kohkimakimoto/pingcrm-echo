package validation

import "fmt"

type Validator struct {
	Errors *ValidationErrors
}

func NewValidator() *Validator {
	return &Validator{
		Errors: NewValidationErrors(),
	}
}

func (v *Validator) HasErrors() bool {
	return v.Errors.Len() > 0
}

func (v *Validator) SetError(key string, value interface{}, message string) {
	v.Errors.Set(key, value, message)
}

func (v *Validator) Required(key string, value interface{}, msgAndArgs ...interface{}) bool {
	var result bool
	switch converted := value.(type) {
	case string:
		if converted == "" {
			result = false
		} else {
			result = true
		}
	default:
		if converted == nil {
			result = false
		} else {
			result = true
		}
	}

	if !result {
		v.Errors.Set(key, value, Message(fmt.Sprintf("The %s is required", key), msgAndArgs...))
		return false
	}
	return true
}
