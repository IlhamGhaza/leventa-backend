package validator

import (
	"testing"
)

func TestValidateRequired(t *testing.T) {
	v := New()

	type Input struct {
		Name string `validate:"required"`
	}

	errs := v.Validate(Input{Name: ""})
	if errs == nil {
		t.Fatal("expected validation errors for empty required field")
	}
	if _, ok := errs["name"]; !ok {
		t.Fatal("expected error for 'name' field")
	}

	errs = v.Validate(Input{Name: "John"})
	if errs != nil {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestValidateEmail(t *testing.T) {
	v := New()

	type Input struct {
		Email string `validate:"required,email"`
	}

	errs := v.Validate(Input{Email: "not-an-email"})
	if errs == nil {
		t.Fatal("expected validation error for invalid email")
	}

	errs = v.Validate(Input{Email: "valid@email.com"})
	if errs != nil {
		t.Fatalf("expected no errors for valid email, got %v", errs)
	}
}

func TestValidateMinMax(t *testing.T) {
	v := New()

	type Input struct {
		Password string `validate:"required,min=6,max=100"`
	}

	errs := v.Validate(Input{Password: "abc"})
	if errs == nil {
		t.Fatal("expected validation error for short password")
	}

	errs = v.Validate(Input{Password: "validpassword"})
	if errs != nil {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateOneof(t *testing.T) {
	v := New()

	type Input struct {
		Role string `validate:"required,oneof=customer sales staff admin"`
	}

	errs := v.Validate(Input{Role: "superuser"})
	if errs == nil {
		t.Fatal("expected validation error for invalid role")
	}

	errs = v.Validate(Input{Role: "admin"})
	if errs != nil {
		t.Fatalf("expected no errors for valid role, got %v", errs)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"PropertyID", "property_i_d"},
		{"email", "email"},
	}

	for _, tt := range tests {
		result := toSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("toSnakeCase(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
