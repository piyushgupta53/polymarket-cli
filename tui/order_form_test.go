package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/piyushgupta/polymarket-cli/internal/api/clob"
)

func TestOrderFormInit(t *testing.T) {
	m := NewOrderFormModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	if m.state != stateForm {
		t.Errorf("state = %d, want stateForm", m.state)
	}
}

func TestOrderFormValidatePrice(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"0.50", true},
		{"0.01", true},
		{"0.99", true},
		{"0.00", false},
		{"1.00", false},
		{"-0.50", false},
		{"abc", false},
	}
	for _, tc := range tests {
		err := validatePrice(tc.input)
		if tc.valid && err != nil {
			t.Errorf("validatePrice(%q) = %v, want nil", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("validatePrice(%q) = nil, want error", tc.input)
		}
	}
}

func TestOrderFormValidateSize(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"10", true},
		{"0.5", true},
		{"0", false},
		{"-1", false},
		{"abc", false},
	}
	for _, tc := range tests {
		err := validateSize(tc.input)
		if tc.valid && err != nil {
			t.Errorf("validateSize(%q) = %v, want nil", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("validateSize(%q) = nil, want error", tc.input)
		}
	}
}

func TestOrderFormSubmitResult(t *testing.T) {
	m := NewOrderFormModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.state = stateSubmitting

	resp := &clob.OrderResponse{Success: true, OrderID: "order-123", Status: "LIVE"}
	updated, _ := m.Update(orderSubmittedMsg{resp: resp})
	m = updated.(*OrderFormModel)

	if m.state != stateResult {
		t.Errorf("state = %d, want stateResult", m.state)
	}
	if m.result == nil {
		t.Fatal("expected result to be set")
	}
	if m.result.OrderID != "order-123" {
		t.Errorf("OrderID = %q, want %q", m.result.OrderID, "order-123")
	}
}

func TestOrderFormSubmitError(t *testing.T) {
	m := NewOrderFormModel("tok1", "Test?", nil, 80, 40)
	m.Init()
	m.state = stateSubmitting

	updated, _ := m.Update(orderSubmitErrorMsg{err: fmt.Errorf("auth required")})
	m = updated.(*OrderFormModel)

	if m.state != stateResult {
		t.Errorf("state = %d, want stateResult", m.state)
	}
	if m.err == nil {
		t.Fatal("expected err to be set")
	}
}

func TestOrderFormResize(t *testing.T) {
	m := NewOrderFormModel("tok1", "Test?", nil, 80, 40)
	m.Init()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = updated.(*OrderFormModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
}
