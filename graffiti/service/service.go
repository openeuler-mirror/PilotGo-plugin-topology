package service

// Type describes the service type
type Type string

func (t Type) String() string {
	return string(t)
}
