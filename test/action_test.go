package test

import (
    "testing"
    "reflect"
    . ".."
)

var actionTestResult string = ""

// action dispatcher
type testDispatcher struct{
}	

func (d *testDispatcher) Dispatch(a Action, c *Context){
	state := c.GetStateMachine().GetCurrentState()
	event := c.GetStateMachine().GetEvent()
	
	sb := (*state).ID() + "|"
	if event != nil {
		sb += (*event).Name()
	}else{
		sb += "nil"
	}
	sb += "|" + a.Name + "|"
	
	for _, p := range a.Parameters {
		sb += reflect.ValueOf(p).String() + "|"
	}
	actionTestResult += sb
}

// test onEntry action
func TestOnEntryAction(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddTransition(Transition{"s2", "s1", "e2", ""})
	sm.AddOnEntry("s2", Action{"a1", nil})
	sm.Start()
	
	// one action
	exp := "s2|e1|a1|"
	actionTestResult = ""
	sm.SendEvent(e1)
	verify(t, "TestOnEntryAction 1", actionTestResult, exp)
	
	// two actions
	sm.AddOnEntry("s1", Action{"a2", nil})
	sm.AddOnEntry("s1", Action{"a3", nil})
	exp = "s1|e2|a2|" + "s1|e2|a3|";
	actionTestResult = "";
	sm.SendEvent(e2)
	verify(t, "TestOnEntryAction 2", actionTestResult, exp)
}

// test onExit action
func TestOnExitAction(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddTransition(Transition{"s2", "s1", "e2", ""})
	sm.AddOnExit("s1", Action{"a3", nil})
	sm.Start()
	
	// one action
	exp := "s1|e1|a3|";
	actionTestResult = "";
	sm.SendEvent(e1)
	verify(t, "TestOnExitAction 1", actionTestResult, exp)
	
	// two actions
	sm.AddOnExit("s2", Action{"a4", nil})
	sm.AddOnExit("s2", Action{"a5", nil})
	exp = "s2|e2|a4|" + "s2|e2|a5|";
	actionTestResult = "";
	sm.SendEvent(e2)
	verify(t, "TestOnExitAction 2", actionTestResult, exp)
}

// test onEntry and onExit action
func TestAction(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddOnEntry("s2", Action{"a1", nil})
	sm.AddOnEntry("s2", Action{"a2", nil})
	sm.AddOnExit("s1", Action{"a3", nil})
	sm.AddOnExit("s1", Action{"a4", nil})
	sm.Start()
	
	exp := "s1|e1|a3|" + "s1|e1|a4|" + "s2|e1|a1|" + "s2|e1|a2|";
	actionTestResult = "";
	sm.SendEvent(e1)
	verify(t, "TestAction", actionTestResult, exp)
}

// test onEntry action when startup
func TestOnEntryActionOnStart(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnEntry("s1", Action{"a1", nil})
	
	exp := "s1|nil|a1|";
	actionTestResult = "";
	sm.Start()
	verify(t, "TestOnEntryActionOnStart", actionTestResult, exp)
}

// test onExit action when stop
func TestOnExitActionOnStop(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnExit("s1", Action{"a1", nil})
	sm.Start()
	
	exp := "s1|nil|a1|";
	actionTestResult = "";
	sm.Stop()
	verify(t, "TestOnExitActionOnStop", actionTestResult, exp)
}

// test action parameters
func TestActionParameters(t *testing.T){
	sm := NewStateMachine(nil, &testDispatcher{})
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	
	ps := make([]Any, 1)
	ps[0] = "v1"
	sm.AddOnEntry("s2", Action{"a1", ps})
	sm.AddOnExit("s1", Action{"a2", ps})
	sm.Start()
	
	exp := "s1|e1|a2|v1|" + "s2|e1|a1|v1|";
	actionTestResult = "";
	sm.SendEvent(e1)
	verify(t, "TestActionParameters", actionTestResult, exp)
}
