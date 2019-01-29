package qstring

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Marshaler defines the interface for performing custom marshaling of struct
// values into query strings.
type Marshaler interface {
	MarshalQuery() (url.Values, error)
}

// Marshal marshals the provided struct into a url.Values collection.
func Marshal(v interface{}) (url.Values, error) {
	var e encoder
	e.init(v)
	return e.marshal()
}

// MarshalString marshals the provided struct into a raw query string and
// returns a conditional error.
func MarshalString(v interface{}) (string, error) {
	values, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return values.Encode(), nil
}

type encoder struct {
	data interface{}
}

func (e *encoder) init(v interface{}) *encoder {
	e.data = v
	return e
}

func (e *encoder) marshal() (url.Values, error) {
	rv := reflect.ValueOf(e.data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, NewInvalidMarshalError(reflect.TypeOf(e.data))
	}

	switch val := e.data.(type) {
	case Marshaler:
		return val.MarshalQuery()
	default:
		return e.value(rv)
	}
}

func (e *encoder) value(val reflect.Value) (url.Values, error) {
	elem := val.Elem()
	typ := elem.Type()

	var err error
	var output = make(url.Values)
	for i := 0; i < elem.NumField(); i++ {
		// pull out the qstring struct tag
		elemField := elem.Field(i)
		typField := typ.Field(i)
		qstring, omit := parseTag(typField.Tag.Get(tag))
		if qstring == "" {
			qstring = strings.ToLower(typField.Name)
		}

		// determine if this is an unsettable field or was explicitly set to be
		// ignored
		if !elemField.CanSet() || qstring == "-" || (omit && isEmptyValue(elemField)) {
			continue
		}

		// only do work if the current fields query string parameter was provided
		switch k := typField.Type.Kind(); k {
		default:
			output.Set(qstring, marshalValue(elemField, k))
		case reflect.Slice:
			output[qstring] = marshalSlice(elemField)
		case reflect.Ptr:
			if err := marshalStruct(output, qstring, reflect.Indirect(elemField), k); err != nil {
				return nil, err
			}
		case reflect.Struct:
			if err := marshalStruct(output, qstring, elemField, k); err != nil {
				return nil, err
			}
		}
	}
	return output, err
}

func marshalSlice(field reflect.Value) []string {
	var out []string
	for i := 0; i < field.Len(); i++ {
		out = append(out, marshalValue(field.Index(i), field.Index(i).Kind()))
	}
	return out
}

func marshalValue(field reflect.Value, source reflect.Kind) string {
	switch source {
	case reflect.String:
		return field.String()
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'G', -1, 64)
	case reflect.Struct:
		// switch field.Interface().(type) {
		// case time.Time:
		// 	return field.Interface().(time.Time).Format(time.RFC3339)
		// case ComparativeTime:
		// 	return field.Interface().(ComparativeTime).String()
		// }
	}
	return ""
}

func marshalStruct(output url.Values, qstring string, field reflect.Value, source reflect.Kind) error {
	var err error
	switch field.Interface().(type) {
	// case time.Time, ComparativeTime:
	// 	output.Set(qstring, marshalValue(field, source))
	default:
		var values url.Values
		if field.CanAddr() {
			values, err = Marshal(field.Addr().Interface())
		}

		if err != nil {
			return err
		}
		for key, list := range values {
			output[key] = list
		}
	}
	return nil
}

// isEmptyValue returns true if the provided reflect.Value
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// switch t := v.Interface().(type) {
		// case time.Time:
		// 	return t.IsZero()
		// case ComparativeTime:
		// 	return t.Time.IsZero()
		// }
	}
	return false
}
