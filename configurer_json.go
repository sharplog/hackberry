package hackberry

import (
	"os"
	"bufio"
	"encoding/json"
)

// Configurer to parse xml file
type ConfigurerImpl struct{
	csm stateMachine
}

// state machine for unmarshal json
type stateMachine struct{
	Defaultstate bool
	Initialstate string
	Timeoutstate string
	States []state
}

// state for unmarshal json
type state struct{
	Id string
	Timeout float64
	Onentry []action
	Onexit []action
	Transitions []transition
}

// action for unmarshal json
type action struct{
	Name string
	Paras []Any
}

// transition for unmarshal json
type transition struct{
	Event string
	Cond string
	Target string
}

func NewConfigurerJSON(file string) *ConfigurerImpl{
	c := &ConfigurerImpl{}
	c.csm.Defaultstate = true
	c.parseStateMachineFromJSON(file)

	return c
}

// load configuration to state machine
func (c *ConfigurerImpl)configure(sm *StateMachine) {
	if sm == nil {
		panic(&ConfigError{"State machine is nil!"})
	}
	
	csm := c.csm
	for _, s := range csm.States {
		c.parseState(s, sm, csm.Defaultstate)
	}
	
    if csm.Initialstate != "" && sm.getState(csm.Initialstate) == nil {
    	panic(&ConfigError{"Has no initial state [" + csm.Initialstate + "]."})
    }
    if csm.Timeoutstate != "" && sm.getState(csm.Timeoutstate) == nil {
    	panic(&ConfigError{"Has no timeout state [" + csm.Timeoutstate + "]."})
    }
    
    sm.SetInitialStateID(csm.Initialstate)
    sm.SetDefaultTimeoutStateID(csm.Timeoutstate)
}

func (c *ConfigurerImpl)parseStateMachineFromJSON(file string){
	input, err := os.Open(file)
    if err != nil {
    	panic(&ConfigError{"An error occurred on opening file: " + file})
    }
    defer input.Close()

	inputReader := bufio.NewReader(input)
	p := json.NewDecoder(inputReader)
	
	if err := p.Decode(&c.csm); err != nil{
		panic(&ConfigError{"Fail to parse config file: " + err.Error()})
	}
}

func (c *ConfigurerImpl)parseState(s state, sm *StateMachine, useDefaultState bool){
	state := sm.getState(s.Id)
	if state == nil && useDefaultState {
		sm.AddState(&DefaultState{s.Id})
		state = sm.getState(s.Id)
	}
	if state == nil {
		panic(&ConfigError{"Has no state [" + s.Id + "]."})
	}
	
	if s.Timeout > 0 {
		sm.AddTimeout(s.Id, int(s.Timeout))
	}
	
	for _, a := range s.Onentry{
		sm.AddOnEntry(s.Id, c.parseAction(a))
	}
	
	for _, a := range s.Onexit{
		sm.AddOnExit(s.Id, c.parseAction(a))
	}
		
	for _, t := range s.Transitions{
		sm.AddTransition(c.parseTransition(s.Id, t))
	}

}

func (c *ConfigurerImpl)parseAction(ja action)(a Action){
	a.Name = ja.Name
	a.Parameters = ja.Paras
	return
}

func (c *ConfigurerImpl)parseTransition(stateId string, tran transition)(t Transition){
	t.SourceID = stateId
	t.TargetID = tran.Target
	t.EventName = tran.Event
	t.Condition = tran.Cond
	return
}
