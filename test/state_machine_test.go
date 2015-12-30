package test

import (
    "testing"
    "reflect"
    . ".."
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

func verify(t *testing.T, fun string, output, expected Any){
	if output != expected {
        t.Errorf("%s: output %v != %v", fun, output, expected)
    }	
}

func verifyNil(t *testing.T, fun string, output Any){
	v := reflect.ValueOf(output)
	if v.IsValid() && !v.IsNil() {
        t.Errorf("%s: output %v is not nil", fun, output)
    }	
}

func verifyPanic(t *testing.T, fun string, errType interface{}, expectedMsg string){
	msgFeild := "Message"
	
	e := recover()
	if e == nil {
		t.Errorf("%s: expected error [%T], but a nil!", fun, errType)
		return
	}

	if reflect.TypeOf(errType) != reflect.TypeOf(e) {
		t.Errorf("%s: expected error [%T], but [%T]!", fun, errType, e)
		return
	}

    errMsg := reflect.ValueOf(e).Elem().FieldByName(msgFeild)
    if errMsg.String() != expectedMsg {
    	t.Errorf("%s: expected error message [%s], but [%s]!", fun, expectedMsg, errMsg)
    	return
    }
}


var s1, s2, s3, s4, s5, s6 State = &myState{"s1"}, &myState{"s2"}, &myState{"s3"}, &myState{"s4"}, &myState{"s5"}, &myState{"s6"}
var e1, e2, e3, e4, e5, e6 Event = &myEvent{"e1"}, &myEvent{"e2"}, &myEvent{"e3"}, &myEvent{"e4"}, &myEvent{"e5"}, &myEvent{"e6"}
var states []State = []State{s1, s2, s3, s4, s5, s6}

func TestStateMachine(t *testing.T) {
	
	states2 := []State{s2, s3}
	
	sm := NewStateMachine(nil, nil)
	sm.AddState(s1).AddStates(states2)
	
	sm.AddTransition(Transition{"s1", "s2", "e1", ""}).
		  AddTransition(Transition{"s2", "s3", "e2", ""}).
		  AddTransition(Transition{"s3", "s1", "e3", ""})
		
	sm.SetInitialStateID("s1");
	sm.Start();
	sm.SendEvent(e1)
	verify(t, "TestStateMachine 1", (*sm.GetCurrentState()).ID(), "s2")
	sm.SendEvent(e2)
	verify(t, "TestStateMachine 2", (*sm.GetCurrentState()).ID(), "s3")
	sm.SendEvent(e3)
	verify(t, "TestStateMachine 3", (*sm.GetCurrentState()).ID(), "s1")
	sm.SendEvent(e1)
	verify(t, "TestStateMachine 4", (*sm.GetCurrentState()).ID(), "s2")
}

func TestStart(t *testing.T){
	sm := NewStateMachine(nil, nil)
	
	sm.AddStates(states[:])
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.SetInitialStateID("s1");
	
	// don't receive event before starting
	sm.SendEvent(e1)
	verifyNil(t, "TestStart 1", sm.GetCurrentState())
	verifyNil(t, "TestStart 2", sm.GetEvent())
	
	sm.Start();
	
	// after staring, the state is initial state and the event is nil
	verify(t, "TestStart 3", (*sm.GetCurrentState()).ID(), "s1")
	verifyNil(t, "TestStart 4", sm.GetEvent())
	
	// receive event
	sm.SendEvent(e1);
	verify(t, "TestStart 5", (*sm.GetCurrentState()).ID(), "s2");
	verify(t, "TestStart 6", (*sm.GetEvent()).Name(), "e1")
}

func TestStop(t *testing.T){
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states);
	sm.AddTransition(Transition{"s1", "s2", "e1", ""});
	sm.AddTransition(Transition{"s1", "s2", "e2", ""});
	
	sm.SetInitialStateID("s1");
	sm.Start();
	sm.SendEvent(e1);
	
	sm.Stop();
	
	// after stopped，state is nil and event is nil
	verifyNil(t, "TestStop 1", sm.GetCurrentState());
	verifyNil(t, "TestStop 2", sm.GetEvent());
	
	// after stopped, dose not receive event
	sm.SendEvent(e2);
	verifyNil(t, "TestStop 3", sm.GetCurrentState());
	verifyNil(t, "TestStop 4", sm.GetEvent());
}

func TestContext(t *testing.T){
	sm := NewStateMachine(nil, nil)
	cxt := sm.GetContext()
	cxt.SetAttribute(1, "abc")
	cxt.SetAttribute("abc", 1)
	
	verify(t, "TestContext 1", cxt.GetAttribute(1), "abc")
	verify(t, "TestContext 2", cxt.GetAttribute("abc"), 1)
	
	as := cxt.GetAttributes()
	verify(t, "TestContext 3",len(as), 2)
}
