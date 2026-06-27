package dge

import (
	"log/slog"
	"testing"
)

func TestAndOpExpression_PositiveCase(t *testing.T) {
	tokens := []token{
		token{"vA", ExpVariable},
		token{"vA2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"", ExpAnd},
	}

	rState := &RuntimeState{
		variable: map[string]string{
			"vA":  "true",
			"vA2": "true",
			"vB":  "false",
			"vB2": "false",
		},
	}

	gCtx := NewGraphWithLogContext(slog.Default(), rState, nil)

	result := evaluate(gCtx, tokens)
	if result != True {
		t.Errorf("unExpected result, want %v got %v", ExpBoolTrue, result)
	}
}

func TestAndOpExpression_NegativeCase(t *testing.T) {
	tokens := []token{
		token{"vA", ExpVariable},
		token{"vA2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"", ExpAnd},
	}

	rState := &RuntimeState{
		variable: map[string]string{
			"vA":  "true",
			"vA2": "true",
			"vB":  "false",
			"vB2": "true",
		},
	}

	gCtx := NewGraphWithLogContext(slog.Default(), rState, nil)

	result := evaluate(gCtx, tokens)
	if result != False {
		t.Errorf("unExpected result, want %v got %v", False, result)
	}
}

func TestOrOpExpression_PositiveCase(t *testing.T) {
	tokens := []token{
		token{"vA", ExpVariable},
		token{"vA2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"", ExpOr},
	}

	rState := &RuntimeState{
		variable: map[string]string{
			"vA":  "true",
			"vA2": "true",
			"vB":  "false",
			"vB2": "false",
		},
	}

	gCtx := NewGraphWithLogContext(slog.Default(), rState, nil)

	result := evaluate(gCtx, tokens)
	if result != True {
		t.Errorf("unExpected result, want %v got %v", True, result)
	}
}

func TestOrOpExpression_NegativeCase(t *testing.T) {
	tokens := []token{
		token{"vA", ExpVariable},
		token{"vA2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"", ExpOr},
		token{"", ExpOr},
	}

	rState := &RuntimeState{
		variable: map[string]string{
			"vA":  "true",
			"vA2": "true",
			"vB":  "false",
			"vB2": "false",
		},
	}

	gCtx := NewGraphWithLogContext(slog.Default(), rState, nil)

	result := evaluate(gCtx, tokens)
	if result != True {
		t.Errorf("unExpected result, want %v got %v", True, result)
	}
}

func TestComplexExpression_WithoutParentness(t *testing.T) {
	tokens := []token{
		token{"vA", ExpVariable},
		token{"vA2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"vB", ExpVariable},
		token{"vB2", ExpVariable},
		token{"", ExpEqual},
		token{"", ExpOr},
		token{"", ExpAnd},
	}

	rState := &RuntimeState{
		variable: map[string]string{
			"vA":  "true",
			"vA2": "true",
			"vB":  "false",
			"vB2": "false",
		},
	}

	gCtx := NewGraphWithLogContext(slog.Default(), rState, nil)

	result := evaluate(gCtx, tokens)
	if result != True {
		t.Errorf("unExpected result, want %v got %v", True, result)
	}
}
