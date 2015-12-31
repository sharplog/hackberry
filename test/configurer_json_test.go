package test

import (
    "testing"
    "os"
//    "fmt"
    
    . ".."
)

var jsondir string = os.Getenv("GOPATH") + "/src/github.com/sharplog/hackberry/test/"

func TestConfigJSON(t *testing.T) {
//	exp := "An error occurred on opening file:"
//	defer verifyPanic(t, "TestConfigJSON", (*ConfigError)(nil), exp)
	
	file := dir + "stateMachine.json"
	dispatcher := NewDefaultActionDispatcher()
	evaluator := NewDefaultConditionEvaluator()
	sm := NewStateMachine(evaluator, dispatcher)
	cfg := NewConfigurerJSON(file)
	sm.LoadConfig(cfg)
}
