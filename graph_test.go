package dge

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func strPtr(s string) *string {
	return &s
}

type MockTask struct {
	task func(gCtx *GraphContext) error
}

func NewMockTask(task func(gCtx *GraphContext) error) *MockTask {
	return &MockTask{
		task: task,
	}
}

func (m *MockTask) Execute(gCtx *GraphContext) error {
	return m.task(gCtx)
}

func newVertex(id string, task func(gCtx *GraphContext) error) *BasicVertex {
	mockTask := NewMockTask(task)
	return &BasicVertex{id: id, task: mockTask, state: Pending}
}

func newEdge(from, to *BasicVertex) *Edge {
	return &Edge{from: from, to: to}
}

func newRuntimeState(state map[string]*BasicVertex) *RuntimeState {
	return &RuntimeState{state: state}
}

func newExpression(tk []token) expression {
	return expression{tokens: tk}
}

var taskFunc = func(err error) *MockTask {
	return &MockTask{
		task: func(gCtx *GraphContext) error {
			return err
		},
	}
}

type (
	tVertex struct {
		id   string
		task Task
	}

	relationship struct {
		from,
		to string
		lOp    tokenType
		pConst state
		tk     []token
	}

	tGraph struct {
		tVertex      []tVertex
		relationship []relationship
		log          GraphLogger
	}
)

func TestGraphWorkflow(t *testing.T) {
	testCase := []struct {
		title  string
		tGraph tGraph
		preRuntimeState,
		postRuntimeState *RuntimeState
		preVertex,
		postVertex []*BasicVertex
	}{
		{
			title: "TestSingleParentDeps_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpAnd,
						pConst: Success,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestSingleParentDeps_FailCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpAnd,
						pConst: Success,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Skipped,
					pendingEdge: 0,
					failEdge:    1,
				},
			},
		}, {
			title: "TestSingleParentDeps_FailCase_with_OR_logical_true",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpOr,
						pConst: Success,
						tk: []token{
							{
								id:    "isValidDate",
								eType: ExpVariable,
							}, {
								id:    "currentDate",
								eType: ExpVariable,
							}, {
								id:    "",
								eType: ExpEqual,
							},
						},
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"isValidDate": time.Now().Format("02-01-2006"),
					"currentDate": time.Now().Format("02-01-2006"),
				},
				vState: make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"isValidDate": time.Now().Format("02-01-2006"),
					"currentDate": time.Now().Format("02-01-2006"),
				},
				vState: make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestSingleParentDeps_FailCase_with_OR_logical_false",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpOr,
						pConst: Success,
						tk: []token{
							{
								id:    "isValidDate",
								eType: ExpVariable,
							}, {
								id:    "currentDate",
								eType: ExpVariable,
							}, {
								id:    "",
								eType: ExpEqual,
							},
						},
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"isValidDate": "02-06-2026",
					"currentDate": time.Now().Format("01-02-2006"),
				},
				vState: make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"isValidDate": "02-06-2026",
					"currentDate": time.Now().Format("01-02-2006"),
				},
				vState: make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Skipped,
					pendingEdge: 0,
					failEdge:    1,
				},
			},
		}, {
			title: "TestMultipleParentDeps_SingleParentStatus_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(nil),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vB",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 3,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vC",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestMultipleParentDeps_MultipleParentStatus_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Fail,
						tk:     nil,
					}, {
						from:   "vB",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Fail,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 3,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vC",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestMultipleParentDeps_MultipleParentStatus_Expression_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     []token{{"var1", ExpVariable}, {"var2", ExpVariable}, {"", ExpEqual}},
					}, {
						from:   "vB",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vD",
						lOp:    ExpOr,
						pConst: Fail,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"var1": "1",
					"var2": "1",
				},
				vState: make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state: make(map[string]*BasicVertex),
				variable: map[string]string{
					"var1": "1",
					"var2": "1",
				},
				vState: make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 3,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vC",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestMultipleRoot_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(nil),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					}, {
						id:   "vE",
						task: taskFunc(nil),
					}, {
						id:   "vF",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vB",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vD",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		}, {
			title: "TestMultipleRoot",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					}, {
						id:   "vE",
						task: taskFunc(nil),
					}, {
						id:   "vF",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Fail,
						tk:     nil,
					}, {
						from:   "vB",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vD",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Skipped,
					pendingEdge: 0,
					failEdge:    1,
				},
			},
		}, {
			title: "TestMultipleRoot_Expression",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(errors.New("mock error")),
					}, {
						id:   "vD",
						task: taskFunc(nil),
					}, {
						id:   "vE",
						task: taskFunc(nil),
					}, {
						id:   "vF",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vB",
						to:     "vE",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vC",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					}, {
						from:   "vD",
						to:     "vF",
						lOp:    ExpAnd,
						pConst: Success,
						tk:     nil,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Pending,
					pendingEdge: 2,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vC",
					state:       Fail,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vD",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vE",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				}, &BasicVertex{
					id:          "vF",
					state:       Skipped,
					pendingEdge: 0,
					failEdge:    1,
				},
			},
		}, {
			title: "TestDepth",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					}, {
						id:   "vC",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpAnd,
						pConst: Success,
					}, {
						from:   "vB",
						to:     "vC",
						lOp:    ExpAnd,
						pConst: Success,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*BasicVertex),
				variable: make(map[string]string),
				vState:   make(map[*BasicVertex]state),
			},
			preVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Pending,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vC",
					state:       Pending,
					pendingEdge: 1,
					failEdge:    0,
				},
			},
			postVertex: []*BasicVertex{
				&BasicVertex{
					id:          "vA",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vB",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
				&BasicVertex{
					id:          "vC",
					state:       Success,
					pendingEdge: 0,
					failEdge:    0,
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.title, func(t *testing.T) {
			g := NewGraph(tc.title)
			gCtx := NewGraphContext(NewLogger(os.Stdout), tc.preRuntimeState, nil)

			for _, v := range tc.tGraph.tVertex {
				g.Add(v.id, v.task)
			}

			for _, r := range tc.tGraph.relationship {
				fV := g.GetVertex(r.from)
				tV := g.GetVertex(r.to)

				if fV == nil || tV == nil {
					t.Fatalf("unexpected while setup test case, cannot find vertex with id %s or %s: from %v, to %v", r.from, r.to, fV, tV)
				}

				g.Connect(fV, tV, r.pConst, r.lOp, r.tk)
			}

			//assert pre test
			if !cmp.Equal(gCtx.rState, tc.preRuntimeState, cmp.AllowUnexported(RuntimeState{}, BasicVertex{}, Edge{}, expression{}, token{})) {
				t.Errorf("unexpected pre-state test result.\n want\t%+v,\n got\t%+v", tc.preRuntimeState, gCtx.rState)
			}

			if !cmp.Equal(g.vertex, tc.preVertex, cmp.AllowUnexported(BasicVertex{}), cmpopts.IgnoreFields(BasicVertex{}, "in", "out", "task")) {
				t.Errorf("unexpected pre-vertex result.\n want\t%+v,\n got\t%+v", tc.preVertex, g.vertex)
			}

			g.RunWithContext(gCtx)

			//assert post test
			if !cmp.Equal(gCtx.rState, tc.postRuntimeState, cmp.AllowUnexported(RuntimeState{}, BasicVertex{}, Edge{}, expression{}, token{})) {
				t.Errorf("unexpected post-state test result.\n want\t%+v,\n got\t%+v", tc.postRuntimeState, gCtx.rState)
			}

			if !cmp.Equal(g.vertex, tc.postVertex, cmp.AllowUnexported(BasicVertex{}), cmpopts.IgnoreFields(BasicVertex{}, "in", "out", "task")) {
				t.Errorf("unexpected post-vertex result.\n want\t%+v,\n got\t%+v", tc.postVertex, g.vertex)
			}
		})
	}

}

func TestLoopGraphWorkflow(t *testing.T) {
	loopGraph := &TabularLoop{
		id:               "TabularLoop1",
		tabularStorageId: "1",
		vState:           Pending,
		maxLoop:          3,
	}

	g := NewGraph("TabularLoop1")

	vA := g.Add("A", taskFunc(nil))
	vB := g.Add("B", taskFunc(nil))

	g.Connect(vA, vB, Success, ExpOr, nil)
	storage := NewStorage()
	storage.SetTabular("1", Tabular{
		columns: []Column{
			{
				name: "A",
				data: []Variable{
					{
						code: 0,
						raw:  []byte("1"),
					},
					{
						code: 0,
						raw:  []byte("2"),
					}, {
						code: 0,
						raw:  []byte("1"),
					},
					{
						code: 0,
						raw:  []byte("2"),
					},
				},
			}, {
				name: "B",
				data: []Variable{
					{
						code: 0,
						raw:  []byte("1"),
					},
					{
						code: 0,
						raw:  []byte("2"),
					}, {
						code: 0,
						raw:  []byte("1"),
					},
					{
						code: 0,
						raw:  []byte("2"),
					},
				},
			},
		},
	})

	loopGraph.graph = g

	gCtx := NewGraphContext(NewLogger(os.Stdout), NewRuntimeState(make(map[string]string)), storage)
	loopGraph.ExecuteTask(gCtx)
	t.Logf("%v", loopGraph)
}
