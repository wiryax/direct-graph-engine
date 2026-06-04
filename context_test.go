package dge

import (
	"testing"
)

func TestContextGetVariable(t *testing.T) {
	rState := &RuntimeState{
		variable: map[string]string{
			"a": "a",
		},
	}

	gCtx := NewGraphContext(NewLogger(nil), rState, nil)
	r, err := gCtx.GetVariable("a")
	if err != nil {
		t.Fatalf("unexpected error, error should be nil\n")
	}

	if r != "a" {
		t.Errorf("unexpected result, want %s, got %s", "a", r)
	}
}

func TestContextSetVariable(t *testing.T) {
	rState := &RuntimeState{
		variable: map[string]string{
			"a": "a",
		},
	}

	gCtx := NewGraphContext(NewLogger(nil), rState, nil)
	gCtx.SetVariable("a", "B")

	r, err := gCtx.GetVariable("a")
	if err != nil {
		t.Fatalf("unexpected error, error should be nil\n")
	}

	if r != "B" {
		t.Errorf("unexpected result, want %s, got %s", "B", r)
	}
}
