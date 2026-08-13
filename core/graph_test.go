package dge

import (
	"errors"
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, want any, got any, title string) {
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s: unexpected result. want %#v, got %#v", title, want, got)
	}
}

func assertShouldNotErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertShouldErr(t *testing.T, err, wantErr error) {
	if err == nil {
		t.Fatalf("expected error: %v", err)
	}

	if !errors.Is(err, wantErr) {
		t.Fatalf("want error: %v, got: %v", wantErr, err)
	}
}

var mockErr = errors.New("mock err")

type mockProcess struct {
	*vertexState
	errOnProcess    bool
	errOnPreprocess bool
}

func (m *mockProcess) preProcess(*GraphContext) error {
	if m.errOnPreprocess {
		return mockErr
	}
	return nil
}

func (m *mockProcess) process(*GraphContext) error {
	if m.errOnProcess {
		return mockErr
	}
	return nil
}

func (m *mockProcess) postProcess() {
}

type mockNonErrTask struct {
	wantErr bool
}

func (*mockNonErrTask) RunTask(_ *GraphContext) error {
	return nil
}

func TestBuilder_ConnectTwoVertexFlow(t *testing.T) {
	g := NewGraph("test graph builder")
	a := NewVertexState("A", &mockNonErrTask{})
	b := NewVertexState("B", &mockNonErrTask{})

	e, err := NewEdge(OnSuccess, a, b)
	assertShouldNotErr(t, err)

	expected := &Graph{
		vertices: map[vertex]struct{}{
			a: {},
			b: {},
		},
		edge: []*Edge{e},
		vertexState: vertexState{
			id: "test graph builder",
		},
	}

	if err := g.Connect(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expected, g) {
		t.Errorf("unexpected result, want %v, got %v", expected, g)
	}
}

func TestBuilder_ConnectOneVertexFlow(t *testing.T) {
	g := NewGraph("test ConnectOneVertexFlow")
	v := NewVertexState("A", &mockNonErrTask{})

	e, err := NewEdge(OnSuccess, v, nil)
	assertShouldNotErr(t, err)

	expected := &Graph{
		vertexState: vertexState{
			id: "test ConnectOneVertexFlow",
		},
		edge: []*Edge{e},
		vertices: map[vertex]struct{}{
			v: {},
		},
	}

	if err := g.Connect(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(g, expected) {
		t.Errorf("unexpected result. want %v, got %v", expected, g)
	}
}

func TestBuilder_ConnectSameVertex(t *testing.T) {
	g := NewGraph("test ConnectSameVertex")
	v := NewVertexState("A", &mockNonErrTask{})

	expected := &Graph{
		vertexState: vertexState{
			id: "test ConnectSameVertex",
		},
		vertices: make(map[vertex]struct{}),
	}

	e, err := NewEdge(OnSuccess, v, v)
	assertShouldErr(t, err, ErrCyclicRelation)
	assertEqual(t, true, e == nil, "nil edge")
	assertEqual(t, expected, g, "")
}

func TestBuilder_SimpleCyclic(t *testing.T) {
	g := NewGraph("test SimpleCyclic")
	a := NewVertexState("A", &mockNonErrTask{})

	b := NewVertexState("B", &mockNonErrTask{})

	e1, err := NewEdge(OnSuccess, a, b)
	assertShouldNotErr(t, err)

	e2, err := NewEdge(OnSuccess, b, a)
	assertShouldNotErr(t, err)

	expected := &Graph{
		vertexState: vertexState{
			id: "test SimpleCyclic",
		},
		vertices: map[vertex]struct{}{
			a: {},
			b: {},
		},
		edge: []*Edge{e1},
	}

	assertShouldNotErr(t, g.Connect(e1))
	assertShouldErr(t, g.Connect(e2), ErrCyclicRelation)
	assertEqual(t, expected, g, "")
}

func TestBuilder_LongTailCyclic(t *testing.T) {
	g := NewGraph("test LongTailCyclic")

	a := NewVertexState("A", &mockNonErrTask{})
	b := NewVertexState("B", &mockNonErrTask{})
	c := NewVertexState("C", &mockNonErrTask{})
	d := NewVertexState("D", &mockNonErrTask{})
	e := NewVertexState("E", &mockNonErrTask{})

	e1, err := NewEdge(OnSuccess, a, b)
	assertShouldNotErr(t, err)

	e11, err := NewEdge(OnSuccess, a, c)
	assertShouldNotErr(t, err)

	e12, err := NewEdge(OnSuccess, c, d)
	assertShouldNotErr(t, err)

	e13, err := NewEdge(OnSuccess, c, e)
	assertShouldNotErr(t, err)

	e14, err := NewEdge(OnSuccess, e, a)
	assertShouldNotErr(t, err)

	assertShouldNotErr(t, g.Connect(e1))
	assertShouldNotErr(t, g.Connect(e11))
	assertShouldNotErr(t, g.Connect(e12))
	assertShouldNotErr(t, g.Connect(e13))
	assertShouldErr(t, g.Connect(e14), ErrCyclicRelation)

	expected := &Graph{
		vertexState: vertexState{
			id: "test LongTailCyclic",
		},
		vertices: map[vertex]struct{}{
			a: {},
			b: {},
			c: {},
			d: {},
			e: {},
		},
		edge: []*Edge{e1, e11, e12, e13},
	}

	assertEqual(t, expected, g, "")
}

func TestWorkflow_SingleNode(t *testing.T) {
	g := NewGraph("SingleNode")

	a := NewVertexState("A", &mockNonErrTask{})
	e, err := NewEdge(OnSuccess, a, nil)
	assertShouldNotErr(t, err)

	assertShouldNotErr(t, g.Connect(e))
	assertShouldNotErr(t, g.Run())

	assertEqual(t, Success, a.execStatus, "1")
	assertEqual(t, int64(0), a.pendingEdge.Load(), "2")
	assertEqual(t, int64(0), a.failEdge.Load(), "3")
}

func TestWorkFlow_Simple2Vertex(t *testing.T) {
	g := NewGraph("Simple2Vertex")
	a := NewVertexState("A", &mockNonErrTask{})
	b := NewVertexState("B", &mockNonErrTask{})
	c := NewVertexState("C", &mockNonErrTask{})
	e, err := NewEdge(OnSuccess, a, b)
	assertShouldNotErr(t, err)
	e1, err := NewEdge(OnSuccess, b, c)
	assertShouldNotErr(t, err)

	assertShouldNotErr(t, g.Connect(e))
	assertShouldNotErr(t, g.Connect(e1))
	assertShouldNotErr(t, g.Run())

	assertEqual(t, int64(0), a.failEdge.Load(), "")
	assertEqual(t, int64(0), a.pendingEdge.Load(), "")
	assertEqual(t, Success, a.execStatus, "")
	assertEqual(t, Success, b.execStatus, "")
	assertEqual(t, Success, b.execStatus, "")
	assertEqual(t, int64(0), b.failEdge.Load(), "")
	assertEqual(t, int64(0), b.pendingEdge.Load(), "")
}

func TestWorkflow_MultipleChild(t *testing.T) {
	g := NewGraph("MultipleChild")
	a := NewVertexState("A", &mockNonErrTask{})
	b := NewVertexState("B", &mockNonErrTask{})
	c := NewVertexState("C", &mockNonErrTask{})

	e, err := NewEdge(OnSuccess, a, b)
	assertShouldNotErr(t, err)

	e1, err := NewEdge(OnSuccess, a, c)
	assertShouldNotErr(t, err)

	assertShouldNotErr(t, g.Connect(e))
	assertShouldNotErr(t, g.Connect(e1))
	assertShouldNotErr(t, g.Run())

	assertEqual(t, Success, a.execStatus, "")
	assertEqual(t, int64(0), a.failEdge.Load(), "")
	assertEqual(t, int64(0), a.pendingEdge.Load(), "")
	assertEqual(t, Success, b.execStatus, "")
	assertEqual(t, int64(0), b.failEdge.Load(), "")
	assertEqual(t, int64(0), b.pendingEdge.Load(), "")
	assertEqual(t, Success, c.execStatus, "")
	assertEqual(t, int64(0), c.failEdge.Load(), "")
	assertEqual(t, int64(0), c.pendingEdge.Load(), "")
}
