package hackberry

import (
    "testing"
)

func TestEvaluator(t *testing.T) {
	evaluator := NewDefaultConditionEvaluator()
	sm := NewStateMachine(evaluator, nil)
	sm.AddStates(states)
	sm.SetInitialStateID("s1")
	sm.AddTransition(Transition{"s1", "s2", "e1", "x=0"})
	sm.AddTransition(Transition{"s1", "s3", "e1", "x=1"})
	sm.AddTransition(Transition{"s2", "s3", "e2", "x<=1"})
	sm.AddTransition(Transition{"s2", "s1", "e2", "x=2"})
	sm.AddTransition(Transition{"s2", "s4", "e4", "x=false"})
	sm.AddTransition(Transition{"s3", "s1", "e3", "x=2"})
	sm.AddTransition(Transition{"s3", "s2", "e3", "x>=3"})
	sm.AddTransition(Transition{"s3", "s4", "e4", "y=abc"})
	sm.AddTransition(Transition{"s4", "s3", "e4", "x=true"})
	sm.Start()
	
	sm.SendEvent(e1);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s1")
	sm.GetContext().SetAttribute("x", "0");
	sm.SendEvent(e1);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s2")
	sm.SendEvent(e2);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s3")
	sm.GetContext().SetAttribute("x", 3.0);
	sm.SendEvent(e3);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s2")
	sm.GetContext().SetAttribute("x", nil);
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s2")
	sm.GetContext().SetAttribute("x", false);
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s4")
	sm.GetContext().SetAttribute("x", true);
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s3")
	sm.Stop();
	
	sm.Start();
	sm.GetContext().SetAttribute("x", int8(1));
	sm.SendEvent(e1);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s3")
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s3")
	sm.GetContext().SetAttribute("y", "abcd");
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s3")
	sm.GetContext().SetAttribute("y", "abc");
	sm.SendEvent(e4);
	verify(t, "TestEvaluator", (*sm.GetCurrentState()).ID(), "s4")
}

