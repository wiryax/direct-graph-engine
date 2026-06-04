package dge

import (
	"fmt"
)

type tokenType int

const (
	ExpUnknown tokenType = iota
	ExpVariable
	ExpState
	ExpEmptyToken
	ExpAnd
	ExpOr
	ExpBoolFalse
	ExpBoolTrue
	ExpOnSuccess
	ExpOnFail
	ExpCompilation
	ExpOnPending
	ExpEqual
	ExpOpenParentheses
	ExpCloseParentheses
)

type expressionState int

const (
	False expressionState = iota
	True
	Unknown
)

type token struct {
	id    string
	eType tokenType
}

func ParseExpression(exp string) (*expression, error) {
	return nil, nil
}

func createNewTk(id string, t tokenType) *token {
	return &token{id, t}
}

type expression struct {
	tokens,
	stack []token
	lTokenScan token
	i, length  int
}

func newStackExpression(length int) *expression {
	if length <= 0 {
		length = 8
	}

	tokens := make([]token, 0, length)
	stack := make([]token, 0, length)

	return &expression{tokens, stack, token{}, 0, length}
}

func (e *expression) push(tk token) {
	e.tokens = append(e.tokens, tk)
}

func (e *expression) pop() token {
	tkLen := len(e.tokens)
	if tkLen == 0 || e.tokens == nil {
		return token{
			id:    "",
			eType: ExpEmptyToken,
		}
	}
	tk := e.tokens[tkLen-1]
	e.tokens = e.tokens[:tkLen-1]

	return tk
}

func tokenToExpState(tk tokenType) expressionState {
	switch tk {
	case ExpBoolTrue:
		return True
	case ExpBoolFalse:
		return False
	default:
		return Unknown
	}
}

func isComparableToken(tk tokenType) bool {
	return tk == ExpEqual
}

func isOperatorToken(tk tokenType) bool {
	return tk == ExpAnd || tk == ExpOr
}

func evaluate(gCtx *GraphContext, queue []token) expressionState {
	qLen := len(queue)

	if qLen == 0 {
		return True
	}

	stackExp := newStackExpression(qLen)

	for j := 0; j < qLen; j++ {
		tk := queue[j]

		switch {
		case isOperatorToken(tk.eType):
			t1 := stackExp.pop()
			t2 := stackExp.pop()

			if t1.eType != ExpBoolTrue && t1.eType != ExpBoolFalse && t2.eType != ExpBoolFalse && t2.eType != ExpBoolTrue {
				panic(fmt.Sprintf("cannot compare token with id %d and %d", t1.eType, t2.eType))
			}

			if tk.eType == ExpAnd {
				if t1.eType != t2.eType {
					tk.eType = ExpBoolFalse
				} else {
					tk.eType = ExpBoolTrue
				}
			} else {
				if t1.eType != ExpBoolTrue && t2.eType != ExpBoolTrue {
					tk.eType = ExpBoolFalse
				} else {
					tk.eType = ExpBoolTrue
				}
			}
		case isComparableToken(tk.eType):
			t1 := stackExp.pop()
			t2 := stackExp.pop()

			var (
				exp1,
				exp2 string
			)

			exp1, err := gCtx.GetVariable(t1.id)
			if err != nil {
				panic(err)
			}

			exp2, err = gCtx.GetVariable(t2.id)
			if err != nil {
				panic(err)
			}

			switch tk.eType {
			case ExpEqual:
				if exp1 != exp2 {
					tk.eType = ExpBoolFalse
				} else {
					tk.eType = ExpBoolTrue
				}
				break
			default:
				panic(fmt.Sprintf("unknown token type with id %d", tk.eType))
			}
		default:
			break
		}
		stackExp.push(tk)
	}

	if len(stackExp.tokens) != 1 {
		panic(fmt.Sprintf("error on tokens composition, remain length %d", len(stackExp.tokens)))
	}

	return tokenToExpState(stackExp.tokens[0].eType)
}

type RuntimeState struct {
	state    map[string]*BasicVertex
	variable map[string]string
	vState   map[*BasicVertex]state
}

func NewRuntimeState(variable map[string]string) *RuntimeState {
	return &RuntimeState{
		variable: variable,
	}
}

func (r *RuntimeState) GetVertexState(id string) tokenType {
	v := r.state[id]
	if v == nil {
		return ExpEmptyToken
	}

	switch v.state {
	case Success:
		return ExpOnSuccess
	case Pending:
		return ExpOnPending
	default:
		return ExpOnFail
	}
}

func (r *RuntimeState) GetVariable(key string) (string, error) {
	v, ok := r.variable[key]
	if !ok {
		return "", fmt.Errorf("cannot find variable with id %s", key)
	}

	return v, nil
}

func (r *RuntimeState) SetVariable(key string, val string) {
	r.variable[key] = val
}
