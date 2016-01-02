package hackberry

import (
    "fmt"
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
        panic(&ActionError{"Action name format should be like objname.method, but [" + a.Name + "]."})
    }
    
    execName := names[0]
    methodName := names[1];
    executor := ad.executors[execName]
    if executor == nil {
        panic(&ActionError{"Has no action executor for [" + execName + "]."})
    }
    
    method := reflect.ValueOf(executor).MethodByName(methodName)
    if !method.IsValid() {
        panic(&ActionError{"Has no method [" + a.Name + "]."})
    }
    
    methodS, _ := reflect.TypeOf(executor).MethodByName(methodName)
    methodT := methodS.Type
    
    // NumIn take receiver as the first parameter
    if methodT.NumIn() - 1 != len(a.Parameters) {
        panic(&ActionError{"Parameter number is not correct for method [" + a.Name + "]."})
    }
    
    params := make([]reflect.Value, len(a.Parameters))
    for i, p := range a.Parameters{
        v := transValue(methodT.In(i + 1).Name(), p, a.Name)
        params[i] = reflect.ValueOf(v)
    }
    method.Call(params)
}

func transValue(name string, v Any, action string) Any{
    s := fmt.Sprintf("%v", v)
    switch name{
        case "bool":
            return parseBool(s)
        case "int8":
            return int8(parseInt(s))
        case "int16":
            return int16(parseInt(s))
        case "int32":
            return int32(parseInt(s))
        case "int64":
            return parseInt(s)
        case "int":
            return int(parseInt(s))
        case "uint8":
            return uint8(parseUint(s))
        case "uint16":
            return uint16(parseUint(s))
        case "uint32":
            return uint32(parseUint(s))
        case "uint64":
            return parseUint(s)
        case "uint":
            return uint(parseUint(s))
        case "float32":
            return float32(parseFloat(s))
        case "float64":
            return parseFloat(s)
        case "string":
            return s
        default:
            return v
    }
}