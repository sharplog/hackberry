package hackberry

import (
	"os"
	"bufio"
	"encoding/json"
)

// Configurer to parse xml file
type ConfigurerJSON struct{
	file string
}

func NewConfigurerJSON(file string) *ConfigurerJSON{
	return &ConfigurerJSON{file}
}

// state machine for unmarshal json
type jStateMachine struct{
	Defaultstate bool
	Initialstate string
	Timeoutstate string
	States []jState
}

// state for unmarshal json
type jState struct{
	Id string
	Timeout float64
	Onentry []jAction
	Onexit []jAction
	Transitions []jTransition
}

// action for unmarshal json
type jAction struct{
	Name string
	Paras []Any
}

// transition for unmarshal json
type jTransition struct{
	Event string
	Cond string
	Target string
}

// load configuration to state machine
func (c *ConfigurerJSON)configure(sm *StateMachine) {
	if sm == nil {
		panic(&ConfigError{"State machine is nil!"})
	}
	
	input, err := os.Open(c.file)
    if err != nil {
    	panic(&ConfigError{"An error occurred on opening file: " + c.file})
    }
    defer input.Close()

	inputReader := bufio.NewReader(input)
	p := json.NewDecoder(inputReader)
	
	var jsm jStateMachine
	
	// default value is true
	jsm.Defaultstate = true
	if err := p.Decode(&jsm); err != nil{
		panic(&ConfigError{"Fail to parse config file: " + err.Error()})
	}
	
	for _, s := range jsm.States {
		c.parseState(s, sm, jsm.Defaultstate)
	}
	
    if jsm.Initialstate != "" && sm.getState(jsm.Initialstate) == nil {
    	panic(&ConfigError{"Has no initial state [" + jsm.Initialstate + "]."})
    }
    if jsm.Timeoutstate != "" && sm.getState(jsm.Timeoutstate) == nil {
    	panic(&ConfigError{"Has no timeout state [" + jsm.Timeoutstate + "]."})
    }
    
    sm.SetInitialStateID(jsm.Initialstate)
    sm.SetDefaultTimeoutStateID(jsm.Timeoutstate)
}

func (c *ConfigurerJSON)parseState(s jState, sm *StateMachine, useDefaultState bool){
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
	
	for _, ja := range s.Onentry{
		sm.AddOnEntry(s.Id, c.parseAction(ja))
	}
	
	for _, ja := range s.Onexit{
		sm.AddOnExit(s.Id, c.parseAction(ja))
	}
		
	for _, jt := range s.Transitions{
		sm.AddTransition(c.parseTransition(s.Id, jt))
	}

}

func (c *ConfigurerJSON)parseAction(ja jAction)(a Action){
	a.Name = ja.Name
	a.Parameters = ja.Paras
	return
}

func (c *ConfigurerJSON)parseTransition(stateId string, jt jTransition)(t Transition){
	t.SourceID = stateId
	t.TargetID = jt.Target
	t.EventName = jt.Event
	t.Condition = jt.Cond
	return
}
