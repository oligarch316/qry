package qry

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// DecodeInfo TODO
type DecodeInfo struct {
	Level  DecodeLevel
	Input  string
	Target reflect.Value
}

func (di DecodeInfo) String() string {
	return fmt.Sprintf(
		"[%s] %s => %s (%s)",
		di.Level,
		strconv.Quote(di.Input),
		di.Target.Type(),
		di.Target.Kind(),
	)
}

// DecodeError TODO
type DecodeError struct {
	DecodeInfo
	err error
}

// Unwrap TODO
func (de DecodeError) Unwrap() error { return de.err }

func (de DecodeError) Error() string {
	return fmt.Sprintf("%s: %s", de.DecodeInfo, de.err)
}

// DecodeLevel TODO
type DecodeLevel int

// Level TODO
const (
	LevelRoot DecodeLevel = iota
	LevelQuery
	LevelField
	LevelKey
	LevelValueList
	LevelValue
)

var decodeLevelNames = map[DecodeLevel]string{
	LevelRoot:      "root",
	LevelQuery:     "query",
	LevelField:     "field",
	LevelKey:       "key",
	LevelValueList: "value list",
	LevelValue:     "value",
}

func (dl DecodeLevel) String() string { return decodeLevelNames[dl] }

func (dl DecodeLevel) newDefault() reflect.Value {
	switch dl {
	case LevelQuery:
		item := make(map[string][]string)
		return reflect.ValueOf(&item).Elem()
	case LevelField:
		var item struct {
			Key    string
			Values []string
		}
		return reflect.ValueOf(&item).Elem()
	case LevelValueList:
		item := make([]string, 0)
		return reflect.ValueOf(&item).Elem()
	}

	var item string
	return reflect.ValueOf(&item).Elem()
}

func (dl DecodeLevel) newInfo(input string, target reflect.Value) DecodeInfo {
	return DecodeInfo{Level: dl, Input: input, Target: target}
}

func (dl DecodeLevel) wrapError(err error, input string, target reflect.Value) DecodeError {
	return DecodeError{err: err, DecodeInfo: dl.newInfo(input, target)}
}

func (dl DecodeLevel) newError(msg string, input string, target reflect.Value) DecodeError {
	return dl.wrapError(errors.New(msg), input, target)
}

func (dl DecodeLevel) newInternalError(msg string, input string, target reflect.Value) DecodeError {
	return dl.wrapError(fmt.Errorf("internal: %w", errors.New(msg)), input, target)
}

// Decoder TODO
type Decoder struct {
	separators SeparatorConfig
	baseModes  levelModes
	logTrace   Trace

	converter   *converter
	unmarshaler *unmarshaler
}

// NewDecoder TODO
func NewDecoder(opts ...Option) *Decoder { return NewConfig().NewDecoder(opts...) }

// Unescape TODO
func (d *Decoder) Unescape(s string) (string, error) { return d.converter.Unescape(s) }

// DecodeQuery TODO
func (d *Decoder) DecodeQuery(query string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelQuery, query, v, traces...)
}

// DecodeField TODO
func (d *Decoder) DecodeField(field string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelField, field, v, traces...)
}

// DecodeKey TODO
func (d *Decoder) DecodeKey(key string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelKey, key, v, traces...)
}

// DecodeValueList TODO
func (d *Decoder) DecodeValueList(valueList string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelValueList, valueList, v, traces...)
}

// DecodeValue TODO
func (d *Decoder) DecodeValue(value string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelValue, value, v, traces...)
}

// Decode TODO
func (d *Decoder) Decode(level DecodeLevel, input string, v interface{}, traces ...Trace) error {
	val := reflect.ValueOf(v)

	switch {
	case val.Kind() != reflect.Ptr:
		return LevelRoot.newError("non-pointer target", input, val)
	case val.IsNil():
		return LevelRoot.newError("nil pointer target", input, val)
	}

	if d.logTrace != nil {
		traces = append(traces, d.logTrace)
	}

	state := &decodeState{
		modes: d.baseModes,
		trace: mergeTraces(traces),
	}

	return d.decode(level, input, val.Elem(), state)
}

type decodeHandler func(DecodeLevel, string, reflect.Value, *decodeState) (bool, error)

func (d *Decoder) decode(level DecodeLevel, raw string, val reflect.Value, state *decodeState) error {
	state.trace.Mark(level, raw, val)

	if !val.CanSet() {
		return level.newError("non-settable target", raw, val)
	}

	for _, handler := range []decodeHandler{
		d.handleIndirects,
		d.handleLiterals,
		d.handleContainers,
	} {
		if complete, err := handler(level, raw, val, state); complete {
			return err
		}
	}

	return level.newError("unsupported target type", raw, val)
}

func (d *Decoder) handleIndirects(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	replaceMode := state.modes[level].ReplaceIndirect

	switch val.Kind() {
	case reflect.Ptr:
		if replaceMode || val.IsZero() {
			elemType := val.Type().Elem()
			val.Set(reflect.New(elemType))
		}

		return true, d.decode(level, raw, val.Elem(), state.child())

	case reflect.Interface:
		if replaceMode || val.IsZero() {
			// Create new item of default type for level, process and set
			newItem := level.newDefault()
			if err := d.decode(level, raw, newItem, state.child()); err != nil {
				return true, err
			}

			val.Set(newItem)
			return true, nil
		}

		elem := val.Elem()

		// NOTE:
		// We descend into non-nil pointers here, but nothing more arduous.
		// It may be possible to descend into other container types (like maps)
		// but it's not yet clear that any real-world use justifies the work
		if elem.Kind() == reflect.Ptr && !elem.IsNil() {
			// Follow pointer and process it's element
			return true, d.decode(level, raw, elem.Elem(), state.child())
		}

		// Create a new item of the same type, process and set
		newItem := reflect.New(elem.Type()).Elem()
		if err := d.decode(level, raw, newItem, state.child()); err != nil {
			return true, err
		}

		val.Set(newItem)
		return true, nil
	}

	return false, nil
}

func (d *Decoder) handleLiterals(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	// Check for unmarshalers
	if complete, err := d.unmarshaler.handle(level, raw, val); complete {
		return true, err
	}

	if val.CanAddr() {
		if complete, err := d.unmarshaler.handle(level, raw, val.Addr()); complete {
			return true, err
		}
	}

	if !state.modes[level].AllowLiteral {
		// Disregard literal kinds unless allowed
		return false, nil
	}

	// Try direct conversion to basic types
	if complete, err := d.converter.handle(level, raw, val); complete {
		return true, err
	}

	// Try container types that should be treated as literals
	return d.handleFauxLiterals(level, raw, val, state)
}

func (d *Decoder) handleFauxLiterals(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	kind := val.Kind()

	// Here we're only interested in slices/arrays of ...
	if kind != reflect.Slice && kind != reflect.Array {
		return false, nil
	}

	var (
		elemType = val.Type().Elem()
		elemKind = elemType.Kind()
	)

	// ... bytes/runes (aliases of uint8/int32 respectively)
	if elemKind != reflect.Uint8 && elemKind != reflect.Int32 {
		return false, nil
	}

	// Don't stomp on user-defined unmarshaling functions
	if d.unmarshaler.check(elemType) || d.unmarshaler.check(reflect.PtrTo(elemType)) {
		// HEURISTIC:
		// This PtrTo check relies on the (correct) assumption that any container
		// handler for reflect.Slice or reflect.Array will create new (valid)
		// values of that slice/array's elements for processing
		return false, nil
	}

	str, err := d.converter.Unescape(raw)
	if err != nil {
		return true, err
	}

	var dstVal, srcVal reflect.Value

	switch elemKind {
	case reflect.Uint8:
		srcVal = reflect.ValueOf([]byte(str))
	case reflect.Int32:
		srcVal = reflect.ValueOf([]rune(str))
	default:
		return true, level.newInternalError("handleFauxLiterals element kind not byte or rune", raw, val)
	}

	switch kind {
	case reflect.Slice:
		dstVal = srcVal
	case reflect.Array:
		var (
			srcLen = srcVal.Len()
			dstLen = val.Len()
		)

		if srcLen > dstLen {
			return true, level.newError("insufficient target length", raw, val)
		}

		dstVal = reflect.New(val.Type()).Elem()
		if n := reflect.Copy(dstVal.Slice(0, srcLen), srcVal); n < srcLen {
			return true, level.newInternalError("handleFauxLiterals short copy", raw, val)
		}
	default:
		return true, level.newInternalError("handleFauxLiterals value kind not slice or array", raw, val)
	}

	val.Set(dstVal)
	return true, nil
}

func (d *Decoder) handleContainers(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	switch level {
	case LevelQuery:
		return d.handleQueryContainers(raw, val, state)
	case LevelField:
		return d.handleFieldContainers(raw, val, state)
	case LevelValueList:
		return d.handleValueListContainers(raw, val, state)
	}

	// LevelKey and LevelValue do not support container kinds
	return false, nil
}

func (d *Decoder) handleQueryContainers(raw string, val reflect.Value, state *decodeState) (bool, error) {
	replaceMode := state.modes[LevelQuery].ReplaceContainer

	switch val.Kind() {
	case reflect.Slice:
		elemType := val.Type().Elem()
		newSlice := reflect.MakeSlice(reflect.SliceOf(elemType), val.Len(), val.Cap())

		if !replaceMode {
			reflect.Copy(newSlice, val)
		}

		for _, rawField := range d.separators.Fields(raw) {
			newElem := reflect.New(elemType).Elem()
			if err := d.decode(LevelField, rawField, newElem, state.child()); err != nil {
				return true, err
			}
			newSlice = reflect.Append(newSlice, newElem)
		}

		val.Set(newSlice)
		return true, nil

		// case reflect.Array:
		// TODO

		// case reflect.Map:
		// TODO

		// case reflect.Struct:
		// TODO
	}

	return false, nil
}

func (d *Decoder) handleFieldContainers(raw string, val reflect.Value, state *decodeState) (bool, error) {
	// TODO: Slices, Arrays, Maps and Structs

	return false, nil
}

func (d *Decoder) handleValueListContainers(raw string, val reflect.Value, state *decodeState) (bool, error) {
	// TODO: Slices and Arrays

	return false, nil
}
