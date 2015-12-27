package hackberry

import (
	"time"
	"sync"
)

type any interface{}

// state machine's state
type State interface{
	ID() string
}

// state machine's event
type Event interface{
	Name() string
}

// condition evaluator
type ConditionEvaluator interface{
	// state machine call this method
	IsSatisfied(condition string, context Context) bool
}

// action executor
type ActionExecutor interface{
	// state machine call this method
	execute(name string, parameters []any, context Context)
}

// state machine's context
type Context struct{
	stateMachine *StateMachine
	
	attributes map[any]any
}

// transition
type Transition struct{
	// source state id
	sourceID string
	
	// target state id
	targetID string
	
	// event name
	eventName string
	
	// condition to transfer
	condition string
}

// action
type Action struct{
	name string
	parameters []any
}	

// status of state machine
const (
	STATUS_STOPPED = iota
	STATUS_RUNNING
)

type StateMachine struct{
	// status，receive event only when running status
	runStatus int
	
	// initial state
	initialStateID string
	
	// current state
	currentState *State
	
	// the previous state
	previousState *State
	
	// the next state
	// the next state is nil normally.
	// when transfering state, before convert to target state, the next state
	// is the target state
	nextState *State
	
	// the event trigger state machine
	event *Event

	// context
	context Context
	
	// all states
	states map[string]*State
	
	// all transitions. Each state has a transition list
	transitions map[string][]Transition
	
	// all entry actions. Each state has a entry action list
	entryActions map[string][]Action
	
	// all exit actions. Each state has a exit action list
	exitActions map[string][]Action
	
	// all states' timeout 
	timeouts map[string]int
	
	// condition evaluator, should been implemented by application
	conditionEvaluator ConditionEvaluator
	
	// action executor, should been implemented by application
	actionExecutor ActionExecutor
	
	// timeout event
	timeoutEvent Event
	
	// default state when timeout
	defaultTimeoutStateID string
	
	// the channel to cancel timeout
	timeoutChannel chan int
	
	// transform locker
	locker sync.Mutex
}

// create a state machine
func NewStateMachine(conditionEvaluator ConditionEvaluator, actionExecutor ActionExecutor) *StateMachine {
	sm := StateMachine{}
	
	sm.context = Context{&sm, make(map[any]any)}
	sm.states = make(map[string]*State)
	sm.transitions = make(map[string][]Transition)
	sm.entryActions = make(map[string][]Action)
	sm.exitActions = make(map[string][]Action)
	sm.timeouts = make(map[string]int)

	sm.conditionEvaluator = conditionEvaluator
	sm.actionExecutor = actionExecutor
	
	return &sm;
}

// get context of state machine
func (sm *StateMachine) GetContext() *Context{
	return &sm.context
}

// add state to state machine
func (sm *StateMachine) AddState(s State) *StateMachine{
	sm.states[s.ID()] = &s
	return sm
}

// add some states to state machine
func (sm *StateMachine) AddStates(ss []State) *StateMachine{
	for i := 0; i < len(ss); i++{
		sm.states[ss[i].ID()] = &ss[i]
	}
	return sm
}

// add transition to state machine. 
// If transition has condition, there should be condition evaluator first.
func (sm *StateMachine) AddTransition(t Transition) *StateMachine{
	if t.condition != "" && sm.conditionEvaluator == nil {
		panic(&ConfigError{"Has no condition evaluator."})
	}

	l := append(sm.transitions[t.sourceID], t)
	sm.transitions[t.sourceID] = l
	
	return sm;
}

// add entry action to state machine. There should be action executor first.
func (sm *StateMachine) AddOnEntry(stateID string, a Action) *StateMachine{
	if sm.actionExecutor == nil {
		panic(&ConfigError{"Has no action executor."})
	}

	l := append(sm.entryActions[stateID], a)
	sm.entryActions[stateID] = l;
	
	return sm;
}

// add exit action to state machine. There should be action executor first.
func (sm *StateMachine) AddOnExit(stateID string, a Action) *StateMachine{
	if sm.actionExecutor == nil {
		panic(&ConfigError{"Has no action executor."})
	}

	l := append(sm.exitActions[stateID], a)
	sm.exitActions[stateID] = l;
	
	return sm;
}

// add timeout to state machine. The timeout event should be set first.
// seconds should be greater than zero.
func (sm *StateMachine) AddTimeout(stateID string, seconds int) *StateMachine{
	if sm.timeoutEvent == nil {
		panic(&ConfigError{"Has no timeout event."})
	}
	
	if seconds > 0 {
		sm.timeouts[stateID] = seconds
	}
	
	return sm;
}
	
// send event to state machine, trigger state transform.
func (sm *StateMachine) SendEvent(event Event){
	sm.locker.Lock()
	defer sm.locker.Unlock()
	
	if !sm.IsRunning() { return }
	
	if target := sm.getTarget(&event); target != nil {
		sm.transitState(&event, target);
	}
}

// get target state by event. lock before call this method
func (sm *StateMachine) getTarget(event *Event) *State{
	trans := sm.transitions[(*sm.currentState).ID()]
	for _, t := range trans{
		if (*event).Name() != t.eventName { continue }
		
		// has condition, but not satisfy
		if "" != t.condition && !sm.conditionEvaluator.IsSatisfied(t.condition, sm.context) {
			continue
		}	
		
		return sm.states[t.targetID]
	}

	// default timeout transition
	if sm.timeoutEvent != nil && sm.timeoutEvent.Name() == (*event).Name() {
		return sm.states[sm.defaultTimeoutStateID]
	}
	return nil
}

// transform state machine to new state
// lock before call this method
func (sm *StateMachine) transitState(event *Event, target *State) {
	sm.cancelTimeout();
	sm.event = event;
	sm.nextState = target;
	
	if sm.currentState != nil {
		// exit actions
		actions := sm.exitActions[(*sm.currentState).ID()]
		for _, a := range actions {
			sm.actionExecutor.execute(a.name, a.parameters, sm.context)
		}
	}
	
	// transform
	sm.previousState = sm.currentState;
	sm.currentState = sm.nextState;
	sm.nextState = nil;
	
	if sm.currentState != nil {
		// entry actions
		actions := sm.entryActions[(*sm.currentState).ID()]
		for _, a := range actions {
			sm.actionExecutor.execute(a.name, a.parameters, sm.context)
		}
		
		// begin to count time for timeout after all entry actions
		sm.createTimeout(sm.currentState);
	}
}

// create timeout
func (sm *StateMachine) createTimeout(state *State) {
	seconds := sm.timeouts[(*state).ID()]
	
	if seconds <= 0 { return }
	
	sm.timeoutChannel = make(chan int)
	go func(){
    	timeout := time.After(time.Duration(seconds) * time.Second)
    	select{
    		case <-sm.timeoutChannel:
    			return
    		case <-timeout:
    			sm.timeoutChannel = nil
    			sm.SendEvent(sm.timeoutEvent)
    	}
	}()
}

// cancel timeout
func (sm *StateMachine) cancelTimeout() {
	if sm.timeoutChannel != nil {
		sm.timeoutChannel <- 0
		sm.timeoutChannel = nil
	}
}

// set state machine's initial state
func (sm *StateMachine) SetInitialStateID(stateID string) *StateMachine{
	sm.initialStateID = stateID
	return sm
}

// load configuration. Should add all states to state machine before call this, if State is not default Type.
func (sm *StateMachine) LoadConfig(configurer *Configurer){
	configurer.configure(sm);
}

// start state machine, transform its state to initial state
func (sm *StateMachine) Start(){
	sm.locker.Lock()
	defer sm.locker.Unlock()
	
	sm.transitState(nil, sm.states[sm.initialStateID]);
	sm.runStatus = STATUS_RUNNING;
}

// stop state machine, it will not receive event 
func (sm *StateMachine) Stop(){
	sm.locker.Lock()
	defer sm.locker.Unlock()
	
	// exit from the last state
	sm.transitState(nil, nil);
	sm.runStatus = STATUS_STOPPED;
}

func (sm *StateMachine) SetTimeoutEvent(event Event) *StateMachine{
	sm.timeoutEvent = event
	return sm
}

// should add timeout state to state machine first
func (sm *StateMachine) SetDefaultTimeoutStateID(stateID string) *StateMachine{
	sm.defaultTimeoutStateID = stateID
	return sm
}

// get state machine's current state.
func (sm *StateMachine) GetCurrentState() *State{
	return sm.currentState;
}

// get state machine's previous state.
func (sm *StateMachine) GetPreviousState() *State{
	return sm.previousState;
}

// get state machine's next state. It's nil normally.
// Only when transforming, the next state is the new target start 
func (sm *StateMachine) GetNextState() *State{
	return sm.nextState;
}

// the event recieved by state machine now
func (sm *StateMachine) GetEvent() *Event{
	return sm.event;
}

// get state machine's all state
func (sm *StateMachine) getStates() []*State{
	states := make([]*State, len(sm.states))
	
	i := 0
	for _, v := range sm.states {
		states[i] = v
		i++
	}
	return states
}

// get state by id
func (sm *StateMachine) getState(id string) *State{
	return sm.states[id]
}

// state machine is running or not
func (sm *StateMachine) IsRunning() bool {
	return sm.runStatus == STATUS_RUNNING
}

// get timeout of one state
func (sm *StateMachine) GetTimeout(state *State) int {
	return sm.timeouts[(*state).ID()]
}
