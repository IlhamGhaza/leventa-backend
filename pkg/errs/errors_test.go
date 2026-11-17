package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := New(400, "bad request")
	if err.Code != 400 {
		t.Errorf("expected code 400, got %d", err.Code)
	}
	if err.Message != "bad request" {
		t.Errorf("expected message 'bad request', got %s", err.Message)
	}
}

func TestAppErrorImplementsError(t *testing.T) {
	var err error = New(500, "server error")
	if err.Error() != "server error" {
		t.Errorf("expected 'server error', got %s", err.Error())
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("Property")
	if err.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.Code)
	}
	if err.Message != "Property not found" {
		t.Errorf("expected 'Property not found', got %s", err.Message)
	}
}

func TestConflict(t *testing.T) {
	err := Conflict("Email already registered")
	if err.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", err.Code)
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("Not allowed")
	if err.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", err.Code)
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("Invalid input")
	if err.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.Code)
	}
}

func TestIsAppError(t *testing.T) {
	appErr := New(404, "not found")
	found, ok := IsAppError(appErr)
	if !ok {
		t.Fatal("expected IsAppError to return true")
	}
	if found.Code != 404 {
		t.Errorf("expected code 404, got %d", found.Code)
	}

	stdErr := errors.New("standard error")
	_, ok = IsAppError(stdErr)
	if ok {
		t.Fatal("expected IsAppError to return false for standard error")
	}
}

func TestPredefinedErrors(t *testing.T) {
	if ErrNotFound.Code != 404 {
		t.Error("ErrNotFound should be 404")
	}
	if ErrUnauthorized.Code != 401 {
		t.Error("ErrUnauthorized should be 401")
	}
	if ErrForbidden.Code != 403 {
		t.Error("ErrForbidden should be 403")
	}
	if ErrConflict.Code != 409 {
		t.Error("ErrConflict should be 409")
	}
	if ErrInternalServer.Code != 500 {
		t.Error("ErrInternalServer should be 500")
	}
}
