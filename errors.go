package qstring

import "reflect"

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func NewInvalidUnmarshalError(t reflect.Type) error {
	return &InvalidUnmarshalError{Type: t}
}

func (e InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "qstring: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "qstring: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "qstring: Unmarshal(nil " + e.Type.String() + ")"
}

// An InvalidMarshalError describes an invalid argument passed to Marshal or
// MarshalValue. (The argument to Marshal must be a non-nil pointer.)
type InvalidMarshalError struct {
	Type reflect.Type
}

func NewInvalidMarshalError(t reflect.Type) error {
	return &InvalidMarshalError{Type: t}
}

func (e InvalidMarshalError) Error() string {
	if e.Type == nil {
		return "qstring: MarshalString(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "qstring: MarshalString(non-pointer " + e.Type.String() + ")"
	}
	return "qstring: MarshalString(nil " + e.Type.String() + ")"
}
