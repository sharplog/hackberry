package test

import (
    "testing"
    "reflect"
    . ".."
)

// action dispatcher
type testDispatcher struct{
	result string
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
	d.result += sb
}

// test onEntry action
func TestOnEntryAction(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddTransition(Transition{"s2", "s1", "e2", ""})
	sm.AddOnEntry("s2", Action{"a1", nil})
	sm.Start()
	
	// one action
	exp := "s2|e1|a1|"
	d.result = ""
	sm.SendEvent(e1)
	verify(t, "TestOnEntryAction 1", d.result, exp)
	
	// two actions
	sm.AddOnEntry("s1", Action{"a2", nil})
	sm.AddOnEntry("s1", Action{"a3", nil})
	exp = "s1|e2|a2|" + "s1|e2|a3|";
	d.result = "";
	sm.SendEvent(e2)
	verify(t, "TestOnEntryAction 2", d.result, exp)
}

// test onExit action
func TestOnExitAction(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddTransition(Transition{"s2", "s1", "e2", ""})
	sm.AddOnExit("s1", Action{"a3", nil})
	sm.Start()
	
	// one action
	exp := "s1|e1|a3|";
	d.result = "";
	sm.SendEvent(e1)
	verify(t, "TestOnExitAction 1", d.result, exp)
	
	// two actions
	sm.AddOnExit("s2", Action{"a4", nil})
	sm.AddOnExit("s2", Action{"a5", nil})
	exp = "s2|e2|a4|" + "s2|e2|a5|";
	d.result = "";
	sm.SendEvent(e2)
	verify(t, "TestOnExitAction 2", d.result, exp)
}

// test onEntry and onExit action
func TestAction(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	sm.AddOnEntry("s2", Action{"a1", nil})
	sm.AddOnEntry("s2", Action{"a2", nil})
	sm.AddOnExit("s1", Action{"a3", nil})
	sm.AddOnExit("s1", Action{"a4", nil})
	sm.Start()
	
	exp := "s1|e1|a3|" + "s1|e1|a4|" + "s2|e1|a1|" + "s2|e1|a2|";
	d.result = "";
	sm.SendEvent(e1)
	verify(t, "TestAction", d.result, exp)
}

// test onEntry action when startup
func TestOnEntryActionOnStart(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnEntry("s1", Action{"a1", nil})
	
	exp := "s1|nil|a1|";
	d.result = "";
	sm.Start()
	verify(t, "TestOnEntryActionOnStart", d.result, exp)
}

// test onExit action when stop
func TestOnExitActionOnStop(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnExit("s1", Action{"a1", nil})
	sm.Start()
	
	exp := "s1|nil|a1|";
	d.result = "";
	sm.Stop()
	verify(t, "TestOnExitActionOnStop", d.result, exp)
}

// test action parameters
func TestActionParameters(t *testing.T){
	d := &testDispatcher{}
	sm := NewStateMachine(nil, d)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	
	ps := make([]Any, 1)
	ps[0] = "v1"
	sm.AddOnEntry("s2", Action{"a1", ps})
	sm.AddOnExit("s1", Action{"a2", ps})
	sm.Start()
	
	exp := "s1|e1|a2|v1|" + "s2|e1|a1|v1|";
	d.result = "";
	sm.SendEvent(e1)
	verify(t, "TestActionParameters", d.result, exp)
}
