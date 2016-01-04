package hackberry

// DefaultState gives a default implementation of state interface.
type DefaultState struct{
    id string
}

// ID implements the State interface method.
func (s *DefaultState) ID() string{
    return s.id
}

// DefaultEvent gives a default implementation of event interface.
type DefaultEvent struct{
    name string
}

// Name implements the Event interface method.
func (e *DefaultEvent) Name() string{
    return e.name
}
