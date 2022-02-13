package app

import "testing"

func TestUser_VerifyPassword(t *testing.T) {
	u := &User{}
	u.SetPlainPassword("secret")
	if !u.VerifyPassword("secret") {
		t.Errorf("password should be verified")
	}
}
