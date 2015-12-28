package hackberry

// default State
type DefaultState struct{
	id string
}

func (s *DefaultState) ID() string{
	return s.id
}

// default Event
type DefaultEvent struct{
	name string
}

func (e *DefaultEvent) Name() string{
	return e.name
}
