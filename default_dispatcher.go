package hackberry

import (
	"strings"
	"reflect"
)

// a simple action executor, to dispatch actions
type defaultActionDispatcher struct{
	executors map[string]Any
}

func NewDefaultActionDispatcher() *defaultActionDispatcher{
	return &defaultActionDispatcher{make(map[string]Any)}
}

// add a action executor to action dispatcher
func (ad *defaultActionDispatcher)AddActionExecutor(name string, executor Any) *defaultActionDispatcher{
	ad.executors[name] = executor
	return ad
}

// dispath a action
func (ad *defaultActionDispatcher)Dispatch(a Action, context *Context){
	names := strings.Split(a.Name, `.`)
	if len(names) != 2 {
		panic(&IllegalActionError{"Action name format should be like objname.method, but [" + a.Name + "]."})
	}
	
	execName := names[0]
	methodName := names[1];
	executor := ad.executors[execName]
	if executor == nil {
		panic(&IllegalActionError{"Has no action executor for [" + execName + "]."})
	}
	
	mtv := reflect.ValueOf(executor).Elem()
	method := mtv.MethodByName(methodName)
	if method.IsNil() {
		panic(&IllegalActionError{"Has no method [" + a.Name + "]."})
	}
	
	params := make([]reflect.Value, len(a.Parameters))
	for i, p := range a.Parameters{
		params[i] = reflect.ValueOf(p)
	}
	method.Call(params)
}