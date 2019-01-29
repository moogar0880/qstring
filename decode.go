package qstring

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshaler defines the interface for performing custom unmarshaling of
// query strings into struct values
type Unmarshaler interface {
	UnmarshalQuery([]string) error
}

// Unmarshal unmarshals the provided url.Values (query string) into the
// interface provided
func Unmarshal(data url.Values, v interface{}) error {
	var d decoder
	d.init(data)
	return d.unmarshal(v)
}

type decoder struct {
	data url.Values
}

func (d *decoder) init(data url.Values) *decoder {
	d.data = data
	return d
}

func (d *decoder) unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return NewInvalidUnmarshalError(reflect.TypeOf(v))
	}
	return d.value(rv)
}

func (d *decoder) elemIsUnmarshaler(val reflect.Value) (reflect.Value, bool) {
	if _, ok := val.Interface().(Unmarshaler); ok {
		return val, ok
	}

	if val.CanAddr() {
		if _, ok := val.Addr().Interface().(Unmarshaler); ok {
			return val.Addr(), ok
		}
	}
	return val, false
}

func (d *decoder) value(val reflect.Value) (err error) {
	var elem reflect.Value
	if val.CanAddr() {
		elem = val.Addr().Elem()
	} else {
		elem = val.Elem()
	}
	typ := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		elemField := elem.Field(i)
		typField := typ.Field(i)

		// pull out the qstring struct tag
		qstring, _ := parseTag(typField.Tag.Get(tag))
		if qstring == "" {
			// if this field doesn't have an explicit tag name, then base it
			// off of the name of the field
			qstring = strings.ToLower(typField.Name)
		}

		// determine if this is an un-settable field or was explicitly set to be
		// ignored
		if !elemField.CanSet() || qstring == "-" {
			continue
		}

		// only do work if the current fields query string parameter was provided
		if query, ok := d.data[qstring]; ok {
			// check up front to see if we can send off unmarshaling logic
			var isUnmarshaler bool
			elemField, isUnmarshaler = d.elemIsUnmarshaler(elemField)
			if isUnmarshaler {
				return elemField.Interface().(Unmarshaler).UnmarshalQuery(query)
			}

			// otherwise, process the field or slice normally
			switch k := typField.Type.Kind(); k {
			case reflect.Slice:
				err = d.coerceSlice(query, k, elemField)
			default:
				err = d.coerce(d.data.Get(qstring), k, elemField)
			}
		} else if typField.Type.Kind() == reflect.Struct {
			if elemField.CanAddr() {
				err = d.value(elemField.Addr())
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// coerce converts the provided query parameter slice into the proper type for
// the target field. this coerced value is then assigned to the current field
func (d *decoder) coerce(query string, target reflect.Kind, field reflect.Value) error {
	var err error
	var c interface{}

	switch target {
	case reflect.String:
		field.SetString(query)
	case reflect.Bool:
		c, err = strconv.ParseBool(query)
		if err == nil {
			field.SetBool(c.(bool))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		c, err = strconv.ParseInt(query, 10, 64)
		if err == nil {
			field.SetInt(c.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		c, err = strconv.ParseUint(query, 10, 64)
		if err == nil {
			field.SetUint(c.(uint64))
		}
	case reflect.Float32, reflect.Float64:
		c, err = strconv.ParseFloat(query, 64)
		if err == nil {
			field.SetFloat(c.(float64))
		}
	case reflect.Struct:
		return d.value(field)
	}
	return err
}

// coerceSlice creates a new slice of the appropriate type for the target field
// and coerces each of the query parameter values into the destination type.
// Should any of the provided query parameters fail to be coerced, an error is
// returned and the entire slice will not be applied
func (d *decoder) coerceSlice(query []string, target reflect.Kind, field reflect.Value) error {
	var err error
	sliceType := field.Type().Elem()
	coerceKind := sliceType.Kind()
	sl := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 0)
	// Create a pointer to a slice value and set it to the slice
	slice := reflect.New(sl.Type())
	slice.Elem().Set(sl)
	for _, q := range query {
		val := reflect.New(sliceType).Elem()
		if err = d.coerce(q, coerceKind, val); err != nil {
			return err
		}
		slice.Elem().Set(reflect.Append(slice.Elem(), val))
	}
	field.Set(slice.Elem())
	return nil
}
