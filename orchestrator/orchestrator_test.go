package orchestrator

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalcHandlerSuccessCase(t *testing.T) {
	expected := `{"id":"1"}`
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(`{"expression":"2+2"}`))
	w := httptest.NewRecorder()
	calculateExpression(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if string(data) != expected {
		t.Errorf("Expected %s but got %v", expected, string(data))
	}
}
func TestCalcHandlerInvalidExpressionCase(t *testing.T) {
	expected := `{"id":"2"}`
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(`{"expression":"2+2("}`))
	w := httptest.NewRecorder()
	calculateExpression(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if string(data) != expected {
		t.Errorf("Expected %s but got %v", expected, string(data))
	}
}
func TestCalcHandlerDivisionByZeroCase(t *testing.T) {
	expected := `{"id":"3"}`
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(`{"expression":"2/0"}`))
	w := httptest.NewRecorder()
	calculateExpression(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if string(data) != expected {
		t.Errorf("Expected %s but got %v", expected, string(data))
	}
}
