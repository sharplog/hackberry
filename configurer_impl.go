package hackberry

import (
	"os"
	"bufio"
	"encoding/json"
	"encoding/xml"
)
//<!-- xml config file format
//     initialstate defines state machine initial state.
//     defaultstate defines use DefaultState as state type or not. It is true default.
//     If using DefaultState, it need not init state machine's state.
//     timeoutstate defines default target state when timeout event happenning.
//	   if using timeout event, should set state machine's timeoutEvent property.
// -->
// 
// <scxml initialstate="s1" timeoutstate="s3" defaultstate="true">
//        <state id="s1" timeout="60">
//                <!-- actions when entering state, has parameters -->
//                <onentry name="ao1.m1">
//                                <para>abc</para>
//                                <para>123</para>
//                                <para>true</para>
//                                <para>456.789</para>
//                                <para>v1</para>
//				   </onentry>
//                 <!-- action has no parameter -->
//                 <onentry name="a2" />       
//                
//                <!-- actions when exiting state -->
//                <onexit name="a3" />
//                <transition event="e1" target="s2" />
//        </state>
//        <state id="s2">
//                <!-- with condition -->
//                <transition event="e2" cond="x=1" target="s3" />
//                <transition event="e2" cond="x=0" target="s1" />
//        </state>
//        <!-- set timeout for state. should set state machine's timeoutEvent first. -->
//        <state id="s3" timeout="30">
//                <transition event="e3" target="s1" />
//                <!-- timeoutEvent's name should be as the follow name, it is "timeout" here -->
//                <transition event="timeout" target="s2" />
//        </state>
// </scxml>

// JSON file format
//{"initialstate":"s1",
// "defaultstate":true,
// "timeoutstate":"s3"	
// "states":[
//   {"id":"s1",
//   	"timeout":60
//     "onexit":[
//         {"name":"a1.M1"}
//       ],
//     "transtions":[
//       {"event":"e1", "target":"s2"}
//     ]},
//   {"id":"s2",
//     "onentry":[
//         {"name":"a1.M2",
//          "paras":["abc", 123, true, 456.789]},
//         {"name":"a2.M1"}
//       ],
//     "transitions":[
//       {"event":"e2", "cond":"x=1", "target":"s3"},
//       {"event":"e2", "cond":"x=0", "target":"s1"}
//     ]},
//   {"id":"s3",
//     "transitions":[
//       {"event":"e3", "target":"s1"}
//     ]}
// ]
//}

// Configurer to parse xml file
type ConfigurerImpl struct{
	csm stateMachine
}

// state machine for unmarshal json
type stateMachine struct{
	Defaultstate bool	`xml:"defaultstate,attr"`
	Initialstate string	`xml:"initialstate,attr"`
	Timeoutstate string	`xml:"timeoutstate,attr"`
	States []state		`xml:"state"`
}

// state for unmarshal json
type state struct{
	Id string			`xml:"id,attr"`
	Timeout float64		`xml:"timeout,attr"`
	Onentry []action	`xml:"onentry"`
	Onexit []action		`xml:"onexit"`
	Transitions []transition	`xml:"transition"`
}

// action for unmarshal json
type action struct{
	Name string		`xml:"name,attr"`
	Paras []Any
	ParasXML []string	`xml:"para"`	// for xml	
}

// transition for unmarshal json
type transition struct{
	Event string	`xml:"event,attr"`
	Cond string		`xml:"cond,attr"`
	Target string	`xml:"target,attr"`
}

func NewConfigurerJSON(JSONfile string) *ConfigurerImpl{
	c := &ConfigurerImpl{}
	c.csm.Defaultstate = true
	c.parseStateMachineFromFile(JSONfile, "json")

	return c
}

func NewConfigurerXML(XMLfile string) *ConfigurerImpl{
	c := &ConfigurerImpl{}
	c.csm.Defaultstate = true
	c.parseStateMachineFromFile(XMLfile, "xml")

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

func (c *ConfigurerImpl)parseStateMachineFromFile(file, format string){
	input, err := os.Open(file)
    if err != nil {
    	panic(&ConfigError{"An error occurred on opening file: " + file})
    }
    defer input.Close()

	inputReader := bufio.NewReader(input)
	switch format {
		case "json":
			p := json.NewDecoder(inputReader)
			err = p.Decode(&c.csm)
		case "xml":
			p := xml.NewDecoder(inputReader)
			err = p.Decode(&c.csm)
	}
	
	if err != nil{
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

func (c *ConfigurerImpl)parseAction(ac action)(a Action){
	a.Name = ac.Name
	
	// action parsed from xml
	if len(ac.ParasXML) > 0 {
		a.Parameters = make([]Any, len(ac.ParasXML))
		for i, s := range ac.ParasXML{
			a.Parameters[i] = s
		}
	}else{
		a.Parameters = ac.Paras
	}	
	return
}

func (c *ConfigurerImpl)parseTransition(stateId string, tran transition)(t Transition){
	t.SourceID = stateId
	t.TargetID = tran.Target
	t.EventName = tran.Event
	t.Condition = tran.Cond
	return
}
