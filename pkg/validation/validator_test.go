package validation

import "testing"

func TestValidator_Required(t *testing.T) {
	v := NewValidator()
	v.Required("name", "")
	if !v.HasErrors() {
		t.Error("Expected validation errors")
	}
}
