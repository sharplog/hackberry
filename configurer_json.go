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
	
	var root map[string]interface{}
	if err := p.Decode(&root); err != nil{
		panic(&ConfigError{"Fail to parse config file: " + err.Error()})
	}
	
	// deal with type assertion failure
	defer func(){
	  	if e := recover(); e != nil{
	  		panic(&ConfigError{"Fail to parse config file: " + err.Error()})
	  	}
  	}()
	
	useDefaultState := true
	if root["defaultstate"] != nil {
		useDefaultState = root["defaultstate"].(bool)
	}
	
	states, _ := root["states"].([]Any)
	c.parseState(states, sm, useDefaultState)
	
	initialStateID := c.transToStr(root["initialstate"])
	timeoutStateID := c.transToStr(root["timeoutstate"])
	
    if initialStateID != "" && sm.getState(initialStateID) == nil {
    	panic(&ConfigError{"Has no initial state [" + initialStateID + "]."})
    }
    if timeoutStateID != "" && sm.getState(timeoutStateID) == nil {
    	panic(&ConfigError{"Has no timeout state [" + timeoutStateID + "]."})
    }
    
    sm.SetInitialStateID(initialStateID)
    sm.SetDefaultTimeoutStateID(timeoutStateID)

}

func (c *ConfigurerJSON)parseState(states []Any, sm *StateMachine, useDefaultState bool){
	if states == nil { return }
	
	
}


func (c *ConfigurerJSON)transToStr(v Any) string{
	if v != nil {
		return v.(string)
	}
	return ""
}
