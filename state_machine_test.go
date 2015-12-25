package hackberry

import (
    "testing"
)

// customized State
type myState struct{
	id string
}

// customized Event
type myEvent struct{
	name string
}

func (s *myState) ID() string{
	return s.id
}

func (e *myEvent) Name() string{
	return e.name
}

func verify(t *testing.T, fun string, output, expected any){
	if output != expected {
        t.Errorf("%s: output %v != %v", fun, output, expected)
    }	
}

func TestStateMachine(t *testing.T) {
	var s1, s2, s3 State = &myState{"s1"}, &myState{"s2"}, &myState{"s3"}
	var e1, e2, e3 Event = &myEvent{"e1"}, &myEvent{"e2"}, &myEvent{"e3"}
	
	states := []*State{&s2, &s3}
	
	sm := NewStateMachine(nil, nil)
	sm.AddState(&s1).AddStates(states)
	
	sm.AddTransition(Transition{"s1", "s2", "e1", ""}).
		  AddTransition(Transition{"s2", "s3", "e2", ""}).
		  AddTransition(Transition{"s3", "s1", "e3", ""})
		
	sm.SetInitialStateID("s1");
	sm.Start();
	sm.SendEvent(&e1)
	verify(t, "TestStateMachine", (*sm.GetCurrentState()).ID(), "s2")
	sm.SendEvent(&e2)
	verify(t, "TestStateMachine", (*sm.GetCurrentState()).ID(), "s3")
	sm.SendEvent(&e3)
	verify(t, "TestStateMachine", (*sm.GetCurrentState()).ID(), "s1")
	sm.SendEvent(&e1)
	verify(t, "TestStateMachine", (*sm.GetCurrentState()).ID(), "s2")
}

