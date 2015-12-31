package test

import (
    "testing"
    "fmt"
    . ".."
)

type actionExecutor struct {
	result string
}

func (ae *actionExecutor)M1(){
	ae.result += "M1|";
}

func (ae *actionExecutor)M2(p1 int16, p2 int64, p3 uint, p4 float32, p5 string, p6 bool){
	ae.result += fmt.Sprintf("M2|%d|%d|%d|%f|%s|%v|", p1, p2, p3, p4, p5, p6)
}

// test invoke method
func TestMethodInvoke(t *testing.T){
	ae := actionExecutor{}
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &ae)
	
	sm := NewStateMachine(nil, dispatcher)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnEntry("s1", Action{"ao1.M1", nil})
	
	exp := "M1|";
	ae.result = "";
	sm.Start()
	verify(t, "TestMethodInvoke", ae.result, exp)
}

// test parameter
func TestMethodParameter(t *testing.T){
	ae := actionExecutor{}
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &ae)
	sm := NewStateMachine(nil, dispatcher)
	sm.AddStates(states)
	sm.SetInitialStateID("s1").AddTransition(Transition{"s1", "s2", "e1", ""})
	
	l := make([]Any, 6)
	l[0] = int16(1)
	l[1] = int64(2)
	l[2] = uint(3)
	l[3] = float32(4)
	l[4] = "str"
	l[5] = true
	sm.AddOnEntry("s1", Action{"ao1.M2", l})
	sm.AddOnExit("s1", Action{"ao1.M2", l})
	
	exp := "M2|1|2|3|4.000000|str|true|";
	ae.result = "";
	sm.Start()
	verify(t, "TestMethodParameter 1", ae.result, exp)
	
	exp = "M2|1|2|3|4.000000|str|true|";
	ae.result = "";
	sm.SendEvent(e1)
	verify(t, "TestMethodParameter 2", ae.result, exp)
}

func TestHasNoActionExecutor(t *testing.T){
	expected := "Has no action executor for [ao2]."
	defer verifyPanic(t, "TestHasNoActionExecutor", (*IllegalActionError)(nil), expected)
	
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &actionExecutor{})
	sm := NewStateMachine(nil, dispatcher)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnEntry("s1", Action{"ao2.M1", nil})
	
	sm.Start()
}

func TestHasNoActionMethod(t *testing.T){
	expected := "Has no method [ao1.mm]."
	defer verifyPanic(t, "TestHasNoActionMethod", (*IllegalActionError)(nil), expected)
	
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &actionExecutor{})
	sm := NewStateMachine(nil, dispatcher)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", ""})
	
	// no method
	sm.AddOnEntry("s1", Action{"ao1.mm", nil})
	
	sm.Start()
}
