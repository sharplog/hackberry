package hackberry

import (

)

type ParseError struct{
    Message string
}

type ConfigError struct{
    Message string
}

type ActionError struct{
    Message string
}

type ConditionError struct{
    Message string
}