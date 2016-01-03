// Package hackberry provides one state machine and its related concepts:
//	State: state of state machine;
//	Event: event that drives state machine changed from one state to another;
//	Transition: one transition defines a changement of state machine, include source state, 
//				target state, event and a optional condition that must be satisfied;
//	Action: actions that will be executed when entering a state or exiting a state;
//	ConditionEvaluator: evaluate the conditions in transition;
//	ActionDispatcher: call action executor when entering or exiting state.
// 
// StateMachine can be set completely using its methods manully, and can also be set with config file.
//
// This package also has default implementation for those interfaces, include DefaultState, DefaultEvent,
// defaultConditionEvaluator, defaultActionDispatcher and configurerImpl.
//
package hackberry

import (
    "time"
    "sync"
)

// Any is an empty interface, can represent anything.
type Any interface{}

// State is an interface that abstracts state machine's states. User should  
// implement it, or using DefaultState. 
type State interface{
	// ID return the id of state, each state should has one unique id.
	// State's id is used in Transition.
    ID() string
}

// Event is an interface that abstracts events that drive state machine
// transforming. User should implement it, or using DefaultEvent. 
type Event interface{
	// Name return the name of event. Event's name is used in Transition.
    Name() string
}

// ConditionEvaluator judges the condition in transition is satisfied or not. 
// User can implement this, or using NewDefaultConditionEvaluator to get the
//  default evaluator.
type ConditionEvaluator interface{
    // IsSatisfied judges if condition is statisfied or not. Before state 
    // machine transforming, it calls this method to decide to transform or not.
    IsSatisfied(condition string, context *Context) bool
}

// ActionDispatcher calls corresponding method when entering or exit state.
// User can implement this, or using NewDefaultActionDispatcher to get the
// default dispatcher.
type ActionDispatcher interface{
    // Dispatch dispatch the action to corresponding method. State machine will 
    // call this method when entering or exiting state if the state has entry
    // actions or exit actions.
    Dispatch(action Action, context *Context)
}

// Configurer loads configuration into state machine.
// It can be implemented by user, or using configurerImpl.
// configurerImpl can parse xml file and json file.
type Configurer interface{
    configure(sm *StateMachine)
}

// Transition defines a state transformation.
type Transition struct{
    // SourceID is the id of the state that transforming begin with.
    SourceID string
    
    // TargetID is the id of the state that transforming begin with.
    TargetID string
    
    // EventName is the name of the event that drives state machine to transform.
    EventName string
    
    // Condition restricts the transformation. Only when the condition is
    // satisfied, the transformation will happen.
    Condition string
}

// Action defines a action when entering or exiting a state.
type Action struct{
	// Name indicates the name of the method should be called.
	// The name format should be fit ActionDispatcher. When using the default 
	// ActionDispatcher, the name should be like "aaa.bbb", aaa indicats a
	// object and bbb indicates the object's method.
    Name string
    
    // Parameters should be delivered to the method when calling it. Before 
    // delivering, each parameter convert to the type that the method need.
    Parameters []Any
}    

// The status of state machine
const (
	// The state machine is not running. It is before calling Start() or after
	// calling Stop().
    STATUS_STOPPED = iota
    
    // The state machine is running. It is after calling Start() and before
    // calling Stop().
    STATUS_RUNNING
)

// StateMachine defines a state machine. There are some step to use StateMachine
// like the following:
//	1. implement State interface and Event interface if needed, or use the default;
//	2. create ActionDispatcher and ConditionEvaluator if needed;
//	3. using NewStateMachine to create a state machine instance;
//	4. add all states to the state machine instance if not using DefaultState and Configurer;
//	5. configure state machine by Configurer or by methods;
//	6. start the state machine;
//	7. send event to the state machine;
//	8. stop the state machine if needed.
type StateMachine struct{
    // state machine status, receive event only when being running status
    runStatus int
    
    // state machine's initial state's id
    initialStateID string
    
    // state machine's current state
    currentState State
    
    // the previous state of state machine
    previousState State
    
    // the next state
    // the next state is nil normally.
    // when transfering state, before convert to target state, the next state
    // is the target state
    nextState State
    
    // the event triggered state machine just now
    event Event

    // state machine's context
    context Context
    
    // all states of this state machine
    states map[string]State
    
    // all transitions of this state machine. Each state has a transition list.
    transitions map[string][]Transition
    
    // all entry actions of this state machine. Each state has a entry action list.
    entryActions map[string][]Action
    
    // all exit actions of this state machine. Each state has a exit action list.
    exitActions map[string][]Action
    
    // all timeouts of state that are greater than zero.
    timeouts map[string]int
    
    // condition evaluator
    conditionEvaluator ConditionEvaluator
    
    // action dispatcher
    actionDispatcher ActionDispatcher
    
    // timeout event. If some states have timeout greater than zero, this 
    // attribute should be set.
    timeoutEvent Event
    
    // thd id of default state when timeout happened.
    defaultTimeoutStateID string
    
    // the channel to cancel timeout
    timeoutChannel chan int
    
    // transform locker
    locker sync.Mutex
}

// NewStateMachine create a state machine instance.
func NewStateMachine(ce ConditionEvaluator, ad ActionDispatcher) *StateMachine{
    sm := StateMachine{}
    
    sm.context = Context{&sm, make(map[Any]Any)}
    sm.states = make(map[string]State)
    sm.transitions = make(map[string][]Transition)
    sm.entryActions = make(map[string][]Action)
    sm.exitActions = make(map[string][]Action)
    sm.timeouts = make(map[string]int)

    sm.conditionEvaluator = ce
    sm.actionDispatcher = ad
    
    return &sm;
}

// GetContext returns the pointer of the context of state machine.
func (sm *StateMachine) GetContext() *Context{
    return &sm.context
}

// AddState adds one state to state machine.
func (sm *StateMachine) AddState(s State) *StateMachine{
    sm.states[s.ID()] = s
    return sm
}

// AddStates adds some states to state machine.
func (sm *StateMachine) AddStates(ss []State) *StateMachine{
    for i := 0; i < len(ss); i++{
        sm.states[ss[i].ID()] = ss[i]
    }
    return sm
}

// AddTransition adds one transition to state machine. If the transition has
// condition, the state machine must has condition evaluator first.
func (sm *StateMachine) AddTransition(t Transition) *StateMachine{
    if t.Condition != "" && sm.conditionEvaluator == nil {
        panic(&ConfigError{"Has no condition evaluator."})
    }

    l := append(sm.transitions[t.SourceID], t)
    sm.transitions[t.SourceID] = l
    
    return sm;
}

// AddOnEntry adds one entry action to state machine. The state machine must
// has action dispatcher first.
func (sm *StateMachine) AddOnEntry(stateID string, a Action) *StateMachine{
    if sm.actionDispatcher == nil {
        panic(&ConfigError{"Has no action dispatcher."})
    }

    l := append(sm.entryActions[stateID], a)
    sm.entryActions[stateID] = l;
    
    return sm;
}

// AddOnExit adds one exit action to state machine. The state machine must
// has action dispatcher first.
func (sm *StateMachine) AddOnExit(stateID string, a Action) *StateMachine{
    if sm.actionDispatcher == nil {
        panic(&ConfigError{"Has no action dispatcher."})
    }

    l := append(sm.exitActions[stateID], a)
    sm.exitActions[stateID] = l;
    
    return sm;
}

// AddTimeout adds a state's timeout to state machine. The state machine must
// has the timeout event set first. Seconds should be greater than zero.
func (sm *StateMachine) AddTimeout(stateID string, seconds int) *StateMachine{
    if sm.timeoutEvent == nil {
        panic(&ConfigError{"Has no timeout event."})
    }
    
    if seconds > 0 {
        sm.timeouts[stateID] = seconds
    }
    
    return sm;
}
    
// SendEvent sends the event to state machine, trigger state transform.
func (sm *StateMachine) SendEvent(event Event){
    sm.locker.Lock()
    defer sm.locker.Unlock()
    
    if !sm.IsRunning() { return }
    
    if target := sm.getTarget(event); target != nil {
        sm.transitState(event, target);
    }
}

// getTarget returns target state by event. Should ock before call this method.
func (sm *StateMachine) getTarget(event Event) State{
    trans := sm.transitions[sm.currentState.ID()]
    for _, t := range trans{
        if event.Name() != t.EventName { continue }
        
        // has condition, but not satisfy
        if "" != t.Condition && !sm.conditionEvaluator.IsSatisfied(t.Condition, &sm.context) {
            continue
        }    
        
        return sm.states[t.TargetID]
    }

    // default timeout transition
    if sm.timeoutEvent != nil && sm.timeoutEvent.Name() == event.Name() {
        return sm.states[sm.defaultTimeoutStateID]
    }
    return nil
}

// transitStatet transforms state machine to new state. Should lock before
// call this method
func (sm *StateMachine) transitState(event Event, target State) {
    sm.cancelTimeout();
    sm.event = event;
    sm.nextState = target;
    
    if sm.currentState != nil {
        // exit actions
        actions := sm.exitActions[sm.currentState.ID()]
        for _, a := range actions {
            sm.actionDispatcher.Dispatch(a, &sm.context)
        }
    }
    
    // transform
    sm.previousState = sm.currentState;
    sm.currentState = sm.nextState;
    sm.nextState = nil;
    
    if sm.currentState != nil {
        // entry actions
        actions := sm.entryActions[sm.currentState.ID()]
        for _, a := range actions {
            sm.actionDispatcher.Dispatch(a, &sm.context)
        }
        
        // begin to count time for timeout after all entry actions
        sm.createTimeout(sm.currentState);
    }
}

// createTimeout creates timeout when enter this state.
func (sm *StateMachine) createTimeout(state State) {
    seconds := sm.timeouts[state.ID()]
    
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

// cancelTimeout cancels the current timeout. 
func (sm *StateMachine) cancelTimeout() {
    if sm.timeoutChannel != nil {
        sm.timeoutChannel <- 0
        sm.timeoutChannel = nil
    }
}

// SetInitialStateID sets the state machine's initial state's id.
func (sm *StateMachine) SetInitialStateID(stateID string) *StateMachine{
    sm.initialStateID = stateID
    return sm
}

// LoadConfig loads state machine configuration using configurer from config
// file. Before call this method, all states should be added to state machine
// if not using DefaultState.
func (sm *StateMachine) LoadConfig(configurer Configurer){
    configurer.configure(sm);
}

// Start starts the state machine, transform its state to initial state and 
// begin to receive event.
func (sm *StateMachine) Start(){
    sm.locker.Lock()
    defer sm.locker.Unlock()
    
    sm.transitState(nil, sm.states[sm.initialStateID]);
    sm.runStatus = STATUS_RUNNING;
}

// Stop stops the state machine, it exit its current state, and will not 
// receive event any more.
func (sm *StateMachine) Stop(){
    sm.locker.Lock()
    defer sm.locker.Unlock()
    
    // exit from the last state
    sm.transitState(nil, nil);
    sm.runStatus = STATUS_STOPPED;
}

// SetTimeoutEvent set a timeout event to the state machine. When timeout 
// happened, the event will be send to state machine.
func (sm *StateMachine) SetTimeoutEvent(event Event) *StateMachine{
    sm.timeoutEvent = event
    return sm
}

// SetDefaultTimeoutStateID sets the id of default timeout state. When timeout
// happened, the state machine trans to this state if there has no corresponding
// transition.
func (sm *StateMachine) SetDefaultTimeoutStateID(stateID string) *StateMachine{
    sm.defaultTimeoutStateID = stateID
    return sm
}

// GetCurrentState return state machine's current state.
func (sm *StateMachine) GetCurrentState() State{
    return sm.currentState;
}

// GetPreviousState return state machine's previous state.
func (sm *StateMachine) GetPreviousState() State{
    return sm.previousState;
}

// GetNextState return state machine's next state. It's nil normally.
// Only when transforming, the next state is the new target state. 
func (sm *StateMachine) GetNextState() State{
    return sm.nextState;
}

// GetEvent return the event recieved by state machine now.
func (sm *StateMachine) GetEvent() Event{
    return sm.event;
}

// getStates return all states of the state machine. 
func (sm *StateMachine) getStates() []State{
    states := make([]State, len(sm.states))
    
    i := 0
    for _, v := range sm.states {
        states[i] = v
        i++
    }
    return states
}

// getState return a state by id.
func (sm *StateMachine) getState(id string) State{
    return sm.states[id]
}

// IsRunning return if the state machine is running or not.
func (sm *StateMachine) IsRunning() bool {
    return sm.runStatus == STATUS_RUNNING
}

// GetTimeout return the timeout seconds of one state.
func (sm *StateMachine) GetTimeout(state State) int {
    return sm.timeouts[state.ID()]
}


// Context is the state machine's context. Application can set some attributes in it,
// and get state machine instance from it.
type Context struct{
    stateMachine *StateMachine
    
    attributes map[Any]Any
}

// GetStateMachine return the state machine instance.
func (c *Context) GetStateMachine() *StateMachine{
    return c.stateMachine
}

// GetAttributes return all attributes in the context.
func (c *Context) GetAttributes() map[Any]Any{
    return c.attributes
}

// GetAttribute return the attribute value by key get attribute from context.
func (c *Context) GetAttribute(key Any) Any{
    return c.attributes[key]
}

// SetAttribute set attribute into the context.
func (c *Context) SetAttribute(key, value Any) {
    c.attributes[key] = value
}
