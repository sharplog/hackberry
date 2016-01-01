package test

import (
    "testing"
    "os"
    "fmt"
    
    . ".."
)

var dir string = os.Getenv("GOPATH") + "/src/github.com/sharplog/hackberry/test/"

func TestConfigFileNotExist(t *testing.T) {
	exp := "An error occurred on opening file:"
	defer verifyPanic(t, "TestConfigFileNotExist", (*ConfigError)(nil), exp)
	
	file := dir + "none.xml"
	sm := NewStateMachine(nil, nil)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
}

// config file has syntax error
func TestConfigFileParseError(t *testing.T) {
	exp := "Fail to parse config file:"
	defer verifyPanic(t, "TestConfigFileParseError", (*ConfigError)(nil), exp)
	
	file := dir + "stateMachine_parseError.xml"
	sm := NewStateMachine(nil, nil)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
}

//// root element is not scxml
//func TestConfigFileRootError(t *testing.T) {
//	exp := "Config file's root element name is not scxml."
//	defer verifyPanic(t, "TestConfigFileRootError", (*ConfigError)(nil), exp)
//	
//	file := dir + "stateMachine_rootError.xml"
//	sm := NewStateMachine(nil, nil)
//	cfg := NewConfigurerXML(file)
//	sm.LoadConfig(cfg)
//}

// don't use default State and state machine has no states.
func TestConfigFileNoState(t *testing.T) {
	exp := "Has no state [s1]."
	defer verifyPanic(t, "TestConfigFileNoState", (*ConfigError)(nil), exp)
	
	file := dir + "stateMachine_noState.xml"
	sm := NewStateMachine(nil, nil)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
}

// use DefaultState as state type
func TestConfigFileDefaultState(t *testing.T) {
	file := dir + "stateMachine_defaultState.xml"
	sm := NewStateMachine(nil, nil)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
}

// use myState as state type
func TestConfigFileUseCustomizedState(t *testing.T) {
	file := dir + "stateMachine.xml"
	
	dispatcher := NewDefaultActionDispatcher()
	evaluator := NewDefaultConditionEvaluator()
	sm := NewStateMachine(evaluator, dispatcher)
	sm.AddStates(states)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
}

type myExecutor1 struct{
	result string
}

func (e *myExecutor1)M1(){
	e.result += "M1|"
}

func (e *myExecutor1)M2(a string, b int, c bool, d float64){
	e.result += fmt.Sprintf("M2|%s|%d|%t|%f|", a, b, c, d)
}

type myExecutor2 struct{
	result string
}

func (e *myExecutor2)M1(){
	e.result += "M1|"
}

// test action parameter and condition
func TestConfigFileParaCondition(t *testing.T) {
	file := dir + "stateMachine.xml"
	
	a1 := &myExecutor1{}
	a2 := &myExecutor2{}
	dispatcher := NewDefaultActionDispatcher()
	dispatcher.AddActionExecutor("a1", a1)
	dispatcher.AddActionExecutor("a2", a2)
	evaluator := NewDefaultConditionEvaluator()
	sm := NewStateMachine(evaluator, dispatcher)
	sm.AddStates(states)
	cfg := NewConfigurerXML(file)
	sm.LoadConfig(cfg)
	
	a1.result = ""
	a2.result = ""
	sm.Start();
	sm.SendEvent(e1)
	exp1 := "M1|M2|abc|123|true|456.789000|"
	exp2 := "M1|"
	verify(t, "TestConfigFileParaCondition 1", a1.result, exp1)
	verify(t, "TestConfigFileParaCondition 2", a2.result, exp2)
	
	sm.GetContext().SetAttribute("x", "0")
	sm.SendEvent(e2)
	verify(t, "TestConfigFileParaCondition 3",(*sm.GetCurrentState()).ID(), "s1")
}
