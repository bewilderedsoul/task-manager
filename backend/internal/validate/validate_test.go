package validate

import "testing"

type sample struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Status   string `validate:"omitempty,oneof=todo done"`
}

func TestStructValid(t *testing.T) {
	if errs := Struct(sample{Email: "a@b.com", Password: "longenough"}); errs != nil {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestStructReportsEachInvalidField(t *testing.T) {
	errs := Struct(sample{Email: "not-an-email", Password: "short", Status: "bogus"})
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}
	for _, field := range []string{"Email", "Password", "Status"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("expected an error for field %q, got %v", field, errs)
		}
	}
}

func TestRequiredMessage(t *testing.T) {
	errs := Struct(sample{Password: "longenough"})
	if errs["Email"] != "is required" {
		t.Errorf("Email error = %q, want %q", errs["Email"], "is required")
	}
}
