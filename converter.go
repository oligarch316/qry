package qry

import (
	"reflect"
	"strconv"
)

// ConfigConvert TODO
type ConfigConvert struct {
	IntegerBase int
	Unescape    func(string) (string, error)
}

type convertSetter func(string, reflect.Value) error

type converter struct {
	ConfigConvert
	kindMap map[reflect.Kind]convertSetter
}

func newConverter(cfg ConfigConvert) *converter {
	res := &converter{ConfigConvert: cfg}

	res.kindMap = map[reflect.Kind]convertSetter{
		reflect.String: res.setString,

		reflect.Bool: res.setBool,

		reflect.Int:   res.intSetter(0),
		reflect.Int8:  res.intSetter(8),
		reflect.Int16: res.intSetter(16),
		reflect.Int32: res.intSetter(32),
		reflect.Int64: res.intSetter(64),

		reflect.Uint:   res.uintSetter(0),
		reflect.Uint8:  res.uintSetter(8),
		reflect.Uint16: res.uintSetter(16),
		reflect.Uint32: res.uintSetter(32),
		reflect.Uint64: res.uintSetter(64),

		reflect.Float32: res.floatSetter(32),
		reflect.Float64: res.floatSetter(64),

		// NOTE: bit sizes displayed here are for underlying floats (real half of the complex)
		reflect.Complex64:  res.complexSetter(32),
		reflect.Complex128: res.complexSetter(64),
	}

	return res
}

func (c *converter) handle(level DecodeLevel, raw string, val reflect.Value) (bool, error) {
	setter, ok := c.kindMap[val.Kind()]
	if !ok {
		return false, nil
	}

	str, err := c.Unescape(raw)
	if err != nil {
		return true, level.wrapError(err, raw, val)
	}

	if err = setter(str, val); err != nil {
		return true, level.wrapError(err, raw, val)
	}

	return true, nil
}

func (c *converter) setString(str string, val reflect.Value) error {
	val.SetString(str)
	return nil
}

func (c *converter) setBool(str string, val reflect.Value) error {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}
	val.SetBool(b)
	return nil
}

func (c *converter) intSetter(bitSize int) convertSetter {
	return func(str string, val reflect.Value) error {
		i, err := strconv.ParseInt(str, c.IntegerBase, bitSize)
		if err != nil {
			return err
		}
		val.SetInt(i)
		return nil
	}
}

func (c *converter) uintSetter(bitSize int) convertSetter {
	return func(str string, val reflect.Value) error {
		ui, err := strconv.ParseUint(str, c.IntegerBase, bitSize)
		if err != nil {
			return err
		}
		val.SetUint(ui)
		return nil
	}
}

func (c *converter) floatSetter(bitSize int) convertSetter {
	return func(str string, val reflect.Value) error {
		f, err := strconv.ParseFloat(str, bitSize)
		if err != nil {
			return err
		}
		val.SetFloat(f)
		return nil
	}
}

func (c *converter) complexSetter(floatBitSize int) convertSetter {
	return func(str string, val reflect.Value) error {
		f, err := strconv.ParseFloat(str, floatBitSize)
		if err != nil {
			return err
		}
		val.SetComplex(complex(f, 0))
		return nil
	}
}
