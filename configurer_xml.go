package hackberry

import (
	"os"
	"bufio"
	"encoding/xml"
	"io"
)
//<!-- 
//     initialstate defines state machine initial state.
//     defaultstate defines use DefaultState as state type or not. It is true default.
//     If using DefaultState, it need not init state machine's state.
//     timeoutstate defines default target state when timeout event happenning.
//	   if using timeout event, should set state machine's timeoutEvent property.
// -->
// 
// <scxml initialstate="s1" timeoutstate="s1" stringstate="true">
//        <state id="s1">
//                <!-- actions when entering state -->
//                <onentry>
//                        <!-- action has parameters -->
//                        <action name="ao1.m1">
//                                <para value="abc" />
//                                <para value="123" />
//                                <para value="456.789" />
//                                <para>true</para>
//                                <para>v1</para>
//                        </action>
//                        
//                        <!-- action has no parameter -->
//                        <action name="a2" />
//                </onentry>
//                
//                <!-- actions when exiting state -->
//                <onexit>
//                        <action name="a3" />
//                </onexit>
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
 
// Configurer to parse xml file
type ConfigurerXML struct{
	file string
}

func NewConfigurerXML(file string) *ConfigurerXML{
	return &ConfigurerXML{file}
}

// load configuration to state machine
func (c *ConfigurerXML)configure(sm *StateMachine) {
	if sm == nil {
		panic(&ConfigError{"State machine is nil!"})
	}
	
	input, err := os.Open(c.file)
    if err != nil {
    	panic(&ConfigError{"An error occurred on opening file: " + c.file})
    }
    defer input.Close()

	inputReader := bufio.NewReader(input)
	p := xml.NewDecoder(inputReader)
	
	var useDefaultState, scxmlRoot, parseEnd bool = true, false, false
	var initialStateID, timeoutStateID string
	var t xml.Token
    for t, err = p.Token(); err == nil && !parseEnd; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "scxml" {
        			scxmlRoot = true
					if v := c.getAttr(token, "defaultstate"); v != ""{
						useDefaultState = (v == "true")
					}	 
					initialStateID = c.getAttr(token, "initialstate")
					timeoutStateID = c.getAttr(token, "timeoutstate")
				}else {
        			if !scxmlRoot {
						panic(&ConfigError{"Config file's root element name is not scxml."})
					}
	        		if token.Name.Local == "state" {
	        			c.parseState(p, token, sm, useDefaultState)
	        		}
	        	}
			case xml.EndElement:
				if token.Name.Local == "scxml" {
					parseEnd = true
				}	
        }
    }
    
    if err != nil && err != io.EOF {
    	panic(&ConfigError{"Fail to parse config file: " + err.Error()})
    }
    if !parseEnd {
    	panic(&ConfigError{"Fail to parse config file: scxml element has no end. "})
    }
    if initialStateID != "" && sm.getState(initialStateID) == nil {
    	panic(&ConfigError{"Has no initial state [" + initialStateID + "]."})
    }
    if timeoutStateID != "" && sm.getState(timeoutStateID) == nil {
    	panic(&ConfigError{"Has no timeout state [" + timeoutStateID + "]."})
    }
    
    sm.SetInitialStateID(initialStateID)
    sm.SetDefaultTimeoutStateID(timeoutStateID)
}

// parse state   
func (c *ConfigurerXML)parseState(p *xml.Decoder, e xml.StartElement, sm *StateMachine, useDefaultState bool) (err error){
	stateID := c.getAttr(e, "id")
	
	state := sm.getState(stateID)
	if state == nil && useDefaultState {
		sm.AddState(&DefaultState{stateID})
		state = sm.getState(stateID)
	}
	if state == nil {
		panic(&ConfigError{"Has no state [" + stateID + "]."})
	}
	
	if timeout := c.getAttr(e, "timeout"); timeout != "" {
		sm.AddTimeout(stateID, int(parseInt(timeout)))
	}
	
	var parseEnd bool
	var t xml.Token
    for t, err = p.Token(); err == nil && !parseEnd; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		switch token.Name.Local {
        			case "onentry":
        				c.parseOnEntryAction(p, stateID, sm)
        			case "onexit":
        				c.parseOnExitAction(p, stateID, sm)
        			case "transition":	
        				c.parseTransition(token, stateID, sm)
        		}
        	case xml.EndElement:
        		if token.Name.Local == "state" {
        			parseEnd = true	
        		}
        }		
	}
    return
}

func (c *ConfigurerXML)parseOnEntryAction(p *xml.Decoder, stateID string, sm *StateMachine) (err error){
	var parseEnd bool
	var t xml.Token
	var a Action
    for t, err = p.Token(); err == nil && !parseEnd; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "action" {
        			a, err = c.parseAction(p, token)
        			if err == nil { sm.AddOnEntry(stateID, a) }
        		}	
        	case xml.EndElement:
        		parseEnd = true	
        }		
	}
    return
}

func (c *ConfigurerXML)parseOnExitAction(p *xml.Decoder, stateID string, sm *StateMachine) (err error){
	var parseEnd bool
	var t xml.Token
	var a Action
    for t, err = p.Token(); err == nil && !parseEnd; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "action" {
        			a, err = c.parseAction(p, token)
        			if err == nil { sm.AddOnEntry(stateID, a) }
        		}	
        	case xml.EndElement:
        		parseEnd = true	
        }		
	}
    return
}

func (c *ConfigurerXML)parseTransition(e xml.StartElement, stateID string, sm *StateMachine){
	t := Transition{SourceID: stateID}
	t.EventName = c.getAttr(e, "event")
	t.Condition = c.getAttr(e, "cond")
	t.TargetID = c.getAttr(e, "target")
	
	sm.AddTransition(t)
}

func (c *ConfigurerXML)parseAction(p *xml.Decoder, e xml.StartElement) (a Action, err error){
	a.Name = c.getAttr(e, "name")
	a.Parameters = make([]Any, 0)
	
	var pValue string
	var inPara, parseEnd bool
	var t xml.Token
    for t, err = p.Token(); err == nil && !parseEnd; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "para" {
        			inPara = true
					pValue = c.getAttr(token, "value")
        		}
        	case xml.CharData:
        		if inPara {
        			pValue = string([]byte(token))
        		}
        	case xml.EndElement:
        		if token.Name.Local == "para" {
        			a.Parameters = append(a.Parameters, pValue)
        			inPara = false
        		}else if token.Name.Local == "action" {
        			parseEnd = true
        		}
        }		
	}
	
	return
}

func (c *ConfigurerXML)getAttr(e xml.StartElement, name string) string{
	for _, attr := range e.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}	
	}
	return ""
}
