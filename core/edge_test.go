package dge

import "testing"

func TestNotifyNoChildEdge(t *testing.T) {
	mock := &mockProcess{
		vertexState: newVertexState("testNotify"),
	}

	e, err := NewEdge(OnSuccess, mock, nil)
	assertShouldNotErr(t, err)

	mock.finish()
	v := e.EvaluateConstrain(nil)

	assertEqual(t, nil, v, "")
}

func TestNotifyEdge(t *testing.T) {
	mock1 := &mockProcess{
		vertexState: newVertexState("TestNotifyEdge1"),
	}
	mock2 := &mockProcess{
		vertexState: newVertexState("TestNotifyEdge2"),
	}

	e, err := NewEdge(OnSuccess, mock1, mock2)
	assertShouldNotErr(t, err)

	mock1.setExecutionStatus(Success)
	mock1.finish()

	v := e.EvaluateConstrain(nil)

	assertEqual(t, mock2, v, "")
}
