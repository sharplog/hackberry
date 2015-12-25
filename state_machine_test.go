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

func verifyStateNil(t *testing.T, fun string, output *State){
	if output != nil {
        t.Errorf("%s: output %v != %v", fun, output, nil)
    }	
}

func verifyEventNil(t *testing.T, fun string, output *Event){
	if output != nil {
        t.Errorf("%s: output %v != %v", fun, output, nil)
    }	
}

var s1, s2, s3 State = &myState{"s1"}, &myState{"s2"}, &myState{"s3"}
var e1, e2, e3 Event = &myEvent{"e1"}, &myEvent{"e2"}, &myEvent{"e3"}
var states []*State = []*State{&s1, &s2, &s3}

func TestStateMachine(t *testing.T) {
	
	states2 := []*State{&s2, &s3}
	
	sm := NewStateMachine(nil, nil)
	sm.AddState(&s1).AddStates(states2)
	
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

func TestStart(t *testing.T){
	sm := NewStateMachine(nil, nil)
	
	sm.AddStates(states)
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.SetInitialStateID("s1");
	
	// don't receive event before starting
	sm.SendEvent(&e1)
	verifyStateNil(t, "TestStart", sm.GetCurrentState())
	verifyEventNil(t, "TestStart", sm.GetEvent(),)
	
	sm.Start();
	
	// after staring, the state is initial state and the event is nil
	verify(t, "TestStart", sm.GetCurrentState(), &s1)
	verifyEventNil(t, "TestStart", sm.GetEvent())
	
	// receive event
	sm.SendEvent(&e1);
	verify(t, "TestStart", sm.GetCurrentState(), &s2);
	verify(t, "TestStart", sm.GetEvent(), &e1)
}

/*
public void testStop() throws ConfigException, IllegalActionException, IllegalConditionException{
	StateMachine<State, Event> sm = new StateMachine<State, Event>();
	sm.addStates(State.values());
	sm.addTransition(new Transition<State>(State.S1, State.S2, "E1", null));
	sm.addTransition(new Transition<State>(State.S1, State.S2, "E2", null));
	
	sm.setInitialState(State.S1);
	sm.start();
	sm.sendEvent(Event.E1);
	
	sm.stop();
	
	// 停止后退出最后状态，并且event是null
	assertNull(sm.getCurrentState());
	assertNull(sm.getEvent());
	
	// 停止后不接受事件
	sm.sendEvent(Event.E2);
	assertNull(sm.getCurrentState());
	assertNull(sm.getEvent());
}
*/
