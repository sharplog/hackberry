package hackberry

import (

)

// 状态，应用可以实现本接口
type State interface{
	String() string
}

type Event interface{
	String() string
}

type StateMachine struct{
	
}
