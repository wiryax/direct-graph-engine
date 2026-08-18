package dge

import (
	"errors"
)

var ErrCyclicDestination = errors.New("cannot connect edge with same destination as source")
var ErrInvalidRelation = errors.New("cannot connect edge with empty parent")

type evaluateStatus int

const (
	Success100 evaluateStatus = iota
	Fail100
	Compilation100
)

type operand int

const (
	OpAnd operand = iota
	OpOr
)

type constrain int

const (
	OnSuccess constrain = iota
	OnFail
	OnCompilation
)

type Edge struct {
	exp    expression
	op     operand
	pConst constrain
	parent vertex
	child  vertex
}

func NewEdge(constrain constrain, parent, child vertex) (*Edge, error) {
	if parent == nil {
		return nil, ErrInvalidRelation
	}

	if parent == child || (child != nil && parent.ID() == child.ID()) {
		return nil, ErrCyclicRelation
	}

	return &Edge{
		parent: parent,
		child:  child,
		pConst: constrain,
	}, nil
}

func (e *Edge) SetExpression(exp expression, op operand) {
	e.exp = exp
	e.op = op
}

func (e *Edge) Out() vertex {
	return e.child
}

func (e *Edge) In() vertex {
	return e.parent
}

func (e *Edge) wired(*GraphContext) error {
	if pv, ok := e.parent.(WriteOnlyBufferVertex); ok {
		if cv, ok := e.child.(ReadOnlyBufferVertex); ok {
			pv.SetSenderBuffer(cv.GetBuffer())
			return nil
		}
		return errors.New("failed wired vertex")
	}

	return nil
}

func (e *Edge) EvaluateConstrain(gCtx *GraphContext) evalResult {
	if e.child == nil {
		return Ready
	}
	if e.pConst == OnCompilation {
		e.child.notify(Compilation100)
		return Ready
	}

	select {
	case <-e.parent.done():
		// case <-gCtx.Done():
		// return Skip
	}

	var (
		status    evaluateStatus
		eval      = evaluate(gCtx, e.exp.tokens)
		constrain = (e.parent.State() == Success && e.pConst == OnSuccess) || (e.parent.State() == Fail && e.pConst == OnFail)
	)

	if e.op == OpAnd && constrain && eval == True {
		status = Success100

	} else if e.op == OpOr && (constrain || eval == True) {
		status = Success100
	} else {
		status = Fail100
	}

	e.child.notify(status)
	return e.child.validate()
}
