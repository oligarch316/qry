package qry

import (
	"encoding"
	"reflect"
)

// RawTextUnmarshaler TODO
type RawTextUnmarshaler interface{ UnmarshalRawText([]byte) error }

// RawString TODO
type RawString string

// UnmarshalRawText TODO
func (rs *RawString) UnmarshalRawText(text []byte) error {
	*rs = RawString(text)
	return nil
}

type unmarshaler struct {
	textUnmarshalerT, rawTextUnmarshalerT reflect.Type
	unescape                              func(string) (string, error)
}

func newUnmarshaler(unescape func(string) (string, error)) *unmarshaler {
	var (
		tu  encoding.TextUnmarshaler
		rtu RawTextUnmarshaler
	)

	return &unmarshaler{
		textUnmarshalerT:    reflect.TypeOf(&tu).Elem(),
		rawTextUnmarshalerT: reflect.TypeOf(&rtu).Elem(),
		unescape:            unescape,
	}
}

func (u *unmarshaler) check(t reflect.Type) bool {
	return t.Implements(u.textUnmarshalerT) || t.Implements(u.rawTextUnmarshalerT)
}

func (u *unmarshaler) handle(level DecodeLevel, raw string, val reflect.Value) (bool, error) {
	var err error

	switch t := val.Interface().(type) {
	case RawTextUnmarshaler:
		err = t.UnmarshalRawText([]byte(raw))
	case encoding.TextUnmarshaler:
		var unescaped string
		if unescaped, err = u.unescape(raw); err == nil {
			err = t.UnmarshalText([]byte(unescaped))
		}
	default:
		return false, nil
	}

	if err != nil {
		err = level.wrapError(err, raw, val)
	}

	return true, err
}
