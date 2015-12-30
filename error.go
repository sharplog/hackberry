package hackberry

import (

)

type ParseError struct{
	Message string
}

type ConfigError struct{
	Message string
}

type IllegalActionError struct{
	Message string
}

type IllegalConditionError struct{
	Message string
}