package validation

import "container/list"

// ValidationError represents a validation error against a specific field.
type ValidationError struct {
	Key     string
	Value   interface{}
	Message string
	element *list.Element
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Prev() *ValidationError {
	return elementToError(e.element.Prev())
}

func (e *ValidationError) Next() *ValidationError {
	return elementToError(e.element.Next())
}

type ValidationErrors struct {
	errors map[string]*ValidationError
	list   *list.List
}

func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		errors: make(map[string]*ValidationError),
		list:   list.New(),
	}
}

func (ve *ValidationErrors) Get(key string) *ValidationError {
	return ve.errors[key]
}

func (ve *ValidationErrors) Set(key string, value interface{}, message string) {
	if err, exists := ve.errors[key]; exists {
		err.Value = value
		err.Message = message
		return
	}

	err := &ValidationError{
		Key:     key,
		Value:   value,
		Message: message,
	}
	err.element = ve.list.PushBack(err)
	ve.errors[key] = err
}

func (ve *ValidationErrors) Len() int {
	return len(ve.errors)
}

func (ve *ValidationErrors) First() *ValidationError {
	return elementToError(ve.list.Front())
}

func (ve *ValidationErrors) Last() *ValidationError {
	return elementToError(ve.list.Back())
}

func elementToError(element *list.Element) *ValidationError {
	if element == nil {
		return nil
	}
	return element.Value.(*ValidationError)
}
