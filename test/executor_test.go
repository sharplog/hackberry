package test

import (
    "testing"
    "fmt"
    . ".."
)

var executorTestResult string = ""

type actionExecutor struct {
}

func (*actionExecutor)M1(){
	executorTestResult += "M1|";
}

func (*actionExecutor)M2(p1 int16, p2 int64, p3 uint, p4 float32, p5 string, p6 bool){
	executorTestResult += fmt.Sprintf("M2|%d|%d|%d|%f|%s|%v|", p1, p2, p3, p4, p5, p6)
}

// test invoke method
func TestMethodInvoke(t *testing.T){
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &actionExecutor{})
	
	sm := NewStateMachine(nil, dispatcher)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddOnEntry("s1", Action{"ao1.M1", nil})
	
	exp := "M1|";
	executorTestResult = "";
	sm.Start()
	verify(t, "TestMethodInvoke", executorTestResult, exp)
}

// test parameter
func TestMethodParameter(t *testing.T){
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("ao1", &actionExecutor{})
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
	executorTestResult = "";
	sm.Start()
	verify(t, "TestMethodParameter 1", executorTestResult, exp)
	
	exp = "M2|1|2|3|4.000000|str|true|";
	executorTestResult = "";
	sm.SendEvent(e1)
	verify(t, "TestMethodParameter 2", executorTestResult, exp)
}

func TestHasNoActionExecutor(t *testing.T){
	expected := "Has no action executor for [ao2]."
	defer func (){
		if e := recover(); e != nil {
	        a, b :=e.(*IllegalActionError) 
	        if !b || a.Message != expected {
	        	t.Errorf("Has no expected error!%s")
	        }
	    }
	}()
	
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
	defer func (){
		if e := recover(); e != nil {
	        a, b :=e.(*IllegalActionError) 
	        if !b || a.Message != expected {
	        	t.Errorf("Has no expected error!%s")
	        }
	    }
	}()
	
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
