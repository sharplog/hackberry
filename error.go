package hackberry

import (

)

// ParseError is created when it's failure to parse a value from a string.
type ParseError struct{
    Message string
}

// ConfigError is created when there is error to configure state machine.
type ConfigError struct{
    Message string
}

// ActionError is created when can't dispatch a action normally.
type ActionError struct{
    Message string
}

// ConditionError is created when can't evaluate a transition condition.
type ConditionError struct{
    Message string
}