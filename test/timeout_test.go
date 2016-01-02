package test

import (
    "testing"
    "time"
    
    . ".."
)

// test timeout event triggering
func TestTimeout(t *testing.T) {
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states).
	  SetInitialStateID("s1").
	  SetTimeoutEvent(timeoutEvent).
	  AddTimeout("s1", 1).
	  AddTransition(Transition{"s1", "s2", "timeoutEvt", ""})
	
	sm.Start()
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestTimeout", sm.GetCurrentState().ID(), "s2")
}

// test canceling timeout event. happend some other events before timeout
func TestTimeoutCancel(t *testing.T) {
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states).
	  SetInitialStateID("s1").
	  SetTimeoutEvent(timeoutEvent).
	  AddTimeout("s1", 1).
	  AddTransition(Transition{"s1", "s2", "timeoutEvt", ""}).
	  AddTransition(Transition{"s1", "s3", "e1", ""})
	
	sm.Start()
	// e1 changed state machine's state
	sm.SendEvent(e1);
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestTimeoutCancel", sm.GetCurrentState().ID(), "s3")
}	

// test not canceling timeout event. 
// happend some other events before timeout, but dose not change state
func TestTimeoutNotCancel(t *testing.T) {
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states).
	  SetInitialStateID("s1").
	  SetTimeoutEvent(timeoutEvent).
	  AddTimeout("s1", 1).
	  AddTransition(Transition{"s1", "s2", "timeoutEvt", ""}).
	  AddTransition(Transition{"s1", "s3", "e1", ""})
	
	sm.Start()
	// e2 dose not changed state machine's state
	sm.SendEvent(e2)
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestTimeoutNotCancel", sm.GetCurrentState().ID(), "s2")
}	

// test default timeout state
func TestDefaultTimeoutState(t *testing.T) {
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states).
	  SetInitialStateID("s1").
	  SetTimeoutEvent(timeoutEvent).
	  AddTimeout("s1", 1).
	  SetDefaultTimeoutStateID("s3")
	
	sm.Start()
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestDefaultTimeoutState", sm.GetCurrentState().ID(), "s3")
}	

// test if the state machine is normal or not after receiving timeout event
func TestTimeoutStateNormal(t *testing.T) {
	sm := NewStateMachine(nil, nil)
	sm.AddStates(states).
	  SetInitialStateID("s1").
	  SetTimeoutEvent(timeoutEvent).
	  AddTimeout("s1", 1).
	  SetDefaultTimeoutStateID("s3").
	  AddTransition(Transition{"s3", "s2", "e1", ""})
	
	sm.Start()
	// atfer timeout, the state should be s3
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestTimeoutStateNormal 1", sm.GetCurrentState().ID(), "s3")
	
	// after e1, the state should be s2
	sm.SendEvent(e1)
	time.Sleep(1200 * time.Millisecond)
	verify(t, "TestTimeoutStateNormal", sm.GetCurrentState().ID(), "s2")
}	
