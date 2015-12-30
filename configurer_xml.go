package hackberry

import (
	"os"
	"bufio"
	"encoding/xml"
)
//<!-- 
//     initialstate指定状态机的初始状态。
//     stringstate指明是否使用String作为状态类型，如果是，就可以不初始化状态机的所有状态。当没有该状态时，直接把ID转换成状态。
//     timeoutstate指定默认的超时转移目标状态，发生timeoutEvent事件时，如果所有的transition都不能执行，就转移到这个状态。
//        如果状态类是自定义的类，则toString方法需要返回状态的ID。
// -->
// 
// <scxml initialstate="s1" timeoutstate="s1" stringstate="true">
//        <state id="s1">
//                <!-- 进入状态时执行的动作 -->
//                <onentry>
//                        <!-- 有参数的动作，支持6种基本类型的参数 -->
//                        <action name="ao1.m1">
//                                <para type="string" value="v1" />
//                                <para type="int" value="1" />
//                                <para type="float" value="2.3" />
//                                <para type="boolean">true</para>
//                                <!-- 缺省类型是string -->
//                                <para value="v1" />
//                        </action>
//                        
//                        <!-- 无参数的动作 -->
//                        <action name="a2" />
//                </onentry>
//                
//                <!-- 退出状态时执行的动作 -->
//                <onexit>
//                        <action name="a3" />
//                </onexit>
//                <transition event="e1" target="s2" />
//        </state>
//        <state id="s2">
//                <!-- 有条件的转移 -->
//                <transition event="e2" cond="x=1" target="s3" />
//                <transition event="e2" cond="x=0" target="s1" />
//        </state>
//        <!-- 可以为状态设置超时时间，超时后状态机自动触发超时事件，超时事件由应用通过setTimeoutEvent来设置 -->
//        <state id="s3" timeout="30">
//                <transition event="e3" target="s1" />
//                <!-- 需要将名称为timeout的事件设置为状态机的timeoutEvent -->
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
	
	var useDefaultState, scxmlRoot bool = true, false
	var initialStateID, timeoutStateID string
    for t, err := p.Token(); err == nil; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "state" {
        			if !scxmlRoot {
						panic(&ConfigError{"Config file's root element name is not scxml."})
					}
        			parseState(p, token, sm, useDefaultState)
        		}else if token.Name.Local == "scxml" {
        			scxmlRoot = true
					if v := getAttr(token, "defaultstate"); v != ""{
						useDefaultState = (v == "true")
					}	 
					initialStateID = getAttr(token, "initialstate")
					timeoutStateID = getAttr(token, "timeoutstate")
				}
			case xml.EndElement:
				if token.Name.Local == "scxml" { break }
        }
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
func parseState(p *xml.Decoder, e xml.StartElement, sm *StateMachine, useDefaultState bool){
	stateID := getAttr(e, "id")
	
	state := sm.getState(stateID)
	if state == nil && useDefaultState {
		sm.AddState(&DefaultState{stateID})
		state = sm.getState(stateID)
	}
	if state == nil {
		panic(&ConfigError{"Has no state [" + stateID + "]."})
	}
	
	if timeout := getAttr(e, "timeout"); timeout != "" {
		sm.AddTimeout(stateID, int(parseInt(timeout)))
	}
	
    for t, err := p.Token(); err == nil; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		switch token.Name.Local {
        			case "onentry":
        				parseOnEntryAction(p, stateID, sm)
        			case "onexit":
        				parseOnExitAction(p, stateID, sm)
        			case "transition":	
        				parseTransition(token, stateID, sm)
        		}
        	case xml.EndElement:
        		break	
        }		
	}
}

func parseOnEntryAction(p *xml.Decoder, stateID string, sm *StateMachine){
    for t, err := p.Token(); err == nil; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "action" {
        			sm.AddOnEntry(stateID, parseAction(p, token))
        		}	
        	case xml.EndElement:
        		break	
        }		
	}
}

func parseOnExitAction(p *xml.Decoder, stateID string, sm *StateMachine){
    for t, err := p.Token(); err == nil; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "action" {
        			sm.AddOnExit(stateID, parseAction(p, token))
        		}	
        	case xml.EndElement:
        		break	
        }		
	}
}

func parseTransition(e xml.StartElement, stateID string, sm *StateMachine){
	t := Transition{SourceID: stateID}
	t.EventName = getAttr(e, "event")
	t.Condition = getAttr(e, "cond")
	t.TargetID = getAttr(e, "target")
	
	sm.AddTransition(t)
}

func parseAction(p *xml.Decoder, e xml.StartElement) Action{
	a := Action{}
	a.Name = getAttr(e, "name")
	a.Parameters = make([]Any, 1)
	
	var pType, pValue string
    for t, err := p.Token(); err == nil; t, err = p.Token() {
        switch token := t.(type) {
        	case xml.StartElement:
        		if token.Name.Local == "para" {
 					pType = getAttr(token, "type")
					pValue = getAttr(token, "value")
					if pType == "" { pType = "string" }
					if pValue != "" { 
						a.Parameters = append(a.Parameters, getParaValue(pType, pValue))
						pType = ""
					}	
        		}
        	case xml.CharData:
        		if pType != "" {
        			pValue := string([]byte(token))
        			a.Parameters = append(a.Parameters, getParaValue(pType, pValue))
        			pType = ""
        		}	
        	case xml.EndElement:
        		if token.Name.Local == "action" { break	}
        }		
	}
	
	return a
}

func getParaValue(pType, pValue string) Any{
	switch pType {
		case "int":
			return parseInt(pValue)
		case "float":
			return parseFloat(pValue)
		case "string":
			return pValue
		case "bool":
			return parseBool(pValue)
	}
	panic(&ConfigError{"Not support parameter type [" + pType + "]."})
}

func getAttr(e xml.StartElement, name string) string{
	for _, attr := range e.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}	
	}
	return ""
}
