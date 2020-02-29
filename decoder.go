package qry

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// DecodeInfo TODO
type DecodeInfo struct {
	Level  DecodeLevel
	Input  string
	Target reflect.Value
}

func (di DecodeInfo) String() string {
	if !di.Target.IsValid() {
		return "no info"
	}

	return fmt.Sprintf(
		"[%s] %q => %s (%s)",
		di.Level,
		di.Input,
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
	// Public levels
	LevelInvalid DecodeLevel = iota
	LevelQuery
	LevelField
	LevelKey
	LevelValueList
	LevelValue

	// TODO: This internal level stuff makes acting on DecodeInfo annoying.
	// Maybe try a second type like:
	// type DecodeSetLevel|DecodePublicLevel|... DecodeLevel

	// Internal levels
	levelRoot
	levelKeyChain
)

var decodeLevelNames = map[DecodeLevel]string{
	LevelInvalid:   "invalid",
	LevelQuery:     "query",
	LevelField:     "field",
	LevelKey:       "key",
	LevelValueList: "value list",
	LevelValue:     "value",

	levelRoot:     "root",
	levelKeyChain: "key chain",
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

// Useful for map and interface elements, which are not addressable and thus not settable
func ensureSettable(val reflect.Value) reflect.Value {
	res := reflect.New(val.Type()).Elem()
	res.Set(val)
	return res
}

// Decoder TODO
type Decoder struct {
	baseModes         levelModes
	ignoreInvalidKeys bool
	logTrace          Trace
	separators        ConfigSeparate

	converter    *converter
	structParser *structParser
	unmarshaler  *unmarshaler
}

// NewDecoder TODO: friendly.go
func NewDecoder(opts ...Option) *Decoder { return NewConfig().NewDecoder(opts...) }

// Unescape TODO: friendly.go
func (d *Decoder) Unescape(s string) (string, error) { return d.converter.Unescape(s) }

// DecodeQuery TODO: friendly.go
func (d *Decoder) DecodeQuery(query string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelQuery, query, v, traces...)
}

// DecodeField TODO: friendly.go
func (d *Decoder) DecodeField(field string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelField, field, v, traces...)
}

// DecodeKey TODO: friendly.go
func (d *Decoder) DecodeKey(key string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelKey, key, v, traces...)
}

// DecodeValueList TODO: friendly.go
func (d *Decoder) DecodeValueList(valueList string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelValueList, valueList, v, traces...)
}

// DecodeValue TODO: friendly.go
func (d *Decoder) DecodeValue(value string, v interface{}, traces ...Trace) error {
	return d.Decode(LevelValue, value, v, traces...)
}

// Decode TODO
func (d *Decoder) Decode(level DecodeLevel, input string, v interface{}, traces ...Trace) error {
	val := reflect.ValueOf(v)

	switch {
	case val.Kind() != reflect.Ptr:
		return levelRoot.newError("non-pointer target", input, val)
	case val.IsNil():
		return levelRoot.newError("nil pointer target", input, val)
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

func (d *Decoder) decode(level DecodeLevel, raw string, val reflect.Value, state *decodeState) error {
	state.trace.Mark(level, raw, val)

	if !val.CanSet() {
		return level.newInternalError("non-settable target", raw, val)
	}

	if complete, err := d.handleIndirects(level, raw, val, state); complete {
		return err
	}

	if complete, err := d.handleLiterals(level, raw, val, state); complete {
		return err
	}

	if complete, err := d.handleContainers(level, raw, val, state); complete {
		return err
	}

	return level.newError("unsupported target type", raw, val)
}

func (d *Decoder) handleIndirects(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	shouldReplace := state.modes[level].ReplaceIndirect || val.IsZero()

	switch val.Kind() {
	case reflect.Ptr:
		if shouldReplace {
			val.Set(reflect.New(val.Type().Elem()))
		}

		return true, d.decode(level, raw, val.Elem(), state.child())

	case reflect.Interface:
		var elem reflect.Value

		if shouldReplace {
			elem = level.newDefault()
		} else {
			elem = ensureSettable(val.Elem())
		}

		if err := d.decode(level, raw, elem, state.child()); err != nil {
			return true, err
		}

		val.Set(elem)
		return true, nil
	}

	return false, nil
}

func (d *Decoder) handleLiterals(level DecodeLevel, raw string, val reflect.Value, state *decodeState) (bool, error) {
	// Check for unmarshalers
	if complete, err := d.unmarshaler.handle(level, raw, val); complete {
		return true, err
	}

	// TODO: Given the CanSet() check/heuristic inherant in decode(...), is there
	// any actual need for this CanAddr() check? (settable ==impies=> addressable, no?)
	if val.CanAddr() {
		if complete, err := d.unmarshaler.handle(level, raw, val.Addr()); complete {
			return true, err
		}
	}

	// Disregard literal kinds unless allowed
	if !state.modes[level].AllowLiteral {
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
		return true, level.wrapError(err, raw, val)
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
			return true, level.newError("insufficient destination array length", raw, val)
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
	shouldReplace := state.modes[level].ReplaceContainer || val.IsZero()

	switch val.Kind() {
	case reflect.Slice:
		var (
			childLevel DecodeLevel
			rawItems   []string
		)

		switch level {
		case LevelQuery:
			childLevel, rawItems = LevelField, d.separators.Fields(raw)
		case LevelValueList:
			childLevel, rawItems = LevelValue, d.separators.Values(raw)
		default:
			// Only query and value list levels support slices
			return false, nil
		}

		var (
			elemType = val.Type().Elem()
			newSlice reflect.Value
		)

		if shouldReplace {
			newSlice = reflect.MakeSlice(reflect.SliceOf(elemType), 0, val.Cap())
		} else {
			newSlice = reflect.MakeSlice(reflect.SliceOf(elemType), val.Len(), val.Cap())
			reflect.Copy(newSlice, val)
		}

		for _, rawItem := range rawItems {
			newElem := reflect.New(elemType).Elem()
			if err := d.decode(childLevel, rawItem, newElem, state.child()); err != nil {
				return true, err
			}
			newSlice = reflect.Append(newSlice, newElem)
		}

		val.Set(newSlice)
		return true, nil

	case reflect.Array:
		var (
			childLevel DecodeLevel
			rawItems   []string
		)

		switch level {
		case LevelQuery:
			childLevel, rawItems = LevelField, d.separators.Fields(raw)
		case LevelValueList:
			childLevel, rawItems = LevelValue, d.separators.Values(raw)
		default:
			// Only query and value list levels support arrays
			return false, nil
		}

		// TODO: While there's certainly no reasonable way to do anything but
		// replace the entire array (how to choose what index to begin writing
		// from otherwise?), would it be more correct to return an error if the
		// set mode for this level is "Update"? Or are errors for array decoding
		// when using defaults ("Update" being default) malicious?

		if val.Len() < len(rawItems) {
			return true, level.newError("insufficient destination array length", raw, val)
		}

		// TODO: Double check but, like struct, shouldn't need to create a new Array
		// here if val.IsZero() is true??? Maybe??? Are arrays still reference types like slices???
		newArray := reflect.New(val.Type()).Elem()

		for i, rawItem := range rawItems {
			if err := d.decode(childLevel, rawItem, newArray.Index(i), state.child()); err != nil {
				return true, err
			}
		}

		val.Set(newArray)
		return true, nil

	case reflect.Map:
		var rawFields []string

		switch level {
		case LevelQuery:
			rawFields = d.separators.Fields(raw)
		case LevelField:
			rawFields = []string{raw}
		default:
			// Only query and field levels support maps
			return false, nil
		}

		var dstMap reflect.Value
		if shouldReplace {
			dstMap = reflect.MakeMap(val.Type())
		} else {
			dstMap = val
		}

		for _, rawField := range rawFields {
			var (
				rawKey, rawValueList = d.separators.KeyVals(rawField)
				rawKeyChain          = d.separators.KeyChain(rawKey)
			)

			if err := d.decodeKeyChain(rawKeyChain, rawValueList, dstMap, state.child()); err != nil {
				return true, err
			}
		}

		if shouldReplace {
			val.Set(dstMap)
		}

		return true, nil

	case reflect.Struct:
		if level != LevelQuery && level != LevelField {
			// Only query and field levels support structs
			return false, nil
		}

		var dstStruct reflect.Value

		if shouldReplace {
			// TODO: We don't really need to create a new struct if val.IsZero() is true here
			dstStruct = reflect.New(val.Type()).Elem()
		} else {
			dstStruct = val
		}

		switch level {
		case LevelQuery:
			rawFields := d.separators.Fields(raw)
			for _, rawField := range rawFields {
				var (
					rawKey, rawValueList = d.separators.KeyVals(rawField)
					rawKeyChain          = d.separators.KeyChain(rawKey)
				)

				if err := d.decodeKeyChain(rawKeyChain, rawValueList, dstStruct, state.child()); err != nil {
					return true, err
				}
			}

		case LevelField:
			items, parseErr := d.structParser.parse(dstStruct)
			if parseErr != nil {
				return true, level.wrapError(parseErr, raw, val)
			}

			rawKey, rawValueList := d.separators.KeyVals(raw)

			// TODO: magic => constant
			if item, ok := items["key"]; ok {
				childState := state.childWithSetMode(LevelKey, item.setOpts)
				if err := d.decode(LevelKey, rawKey, item.val, childState); err != nil {
					return true, err
				}
			}

			// TODO: magic => constant
			if item, ok := items["values"]; ok {
				childState := state.childWithSetMode(LevelValueList, item.setOpts)
				if err := d.decode(LevelValueList, rawValueList, item.val, childState); err != nil {
					return true, err
				}
			}
		}

		if shouldReplace {
			val.Set(dstStruct)
		}

		return true, nil
	}

	return false, nil
}

func (d *Decoder) decodeKeyChain(rawChain []string, raw string, val reflect.Value, state *decodeState) error {
	// Shuttle work off to decode() once key chain is exhausted
	if len(rawChain) < 1 {
		return d.decode(LevelValueList, raw, val, state)
	}

	// TODO: We're breaking the nice structured nature of DecodeInfo with this
	// custom string formatting of "input"
	//
	// Also doing unneeded format work when trace is a no-op. Need to reconsider
	// just checking for nil rather than a noop function for "no trace"
	inputStr := fmt.Sprintf("%s | %s", strings.Join(rawChain, ", "), raw)

	state.trace.Mark(levelKeyChain, inputStr, val)

	kind := val.Kind()

	if kind == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}

		return d.decodeKeyChain(rawChain, raw, val.Elem(), state.child())
	}

	rawKey, remainingChain := rawChain[0], rawChain[1:]

	if kind == reflect.Map {
		if val.IsZero() {
			val.Set(reflect.MakeMap(val.Type()))
		}

		var (
			keyType  = val.Type().Key()
			elemType = val.Type().Elem()
			newKey   = reflect.New(keyType).Elem()
		)

		if err := d.decode(LevelKey, rawKey, newKey, state.child()); err != nil {
			return err
		}

		elem := val.MapIndex(newKey)
		if !elem.IsValid() {
			// Map does not contain newKey
			elem = reflect.New(elemType).Elem()
		} else {
			// NOTE/TODO?
			// This is potentially a heavy cost, but optimizing might be very complex
			// (namely breaking the very useful "always .CanSet()" heuristic)
			//
			// Elaboration example: long chain of non-zero maps terminating in a
			// pointer to the value list target wouldn't need all the ensureSettable
			// calls. The final pointer makes the value list item settable.
			elem = ensureSettable(elem)
		}

		if err := d.decodeKeyChain(remainingChain, raw, elem, state.child()); err != nil {
			return err
		}

		val.SetMapIndex(newKey, elem)
		return nil
	}

	if kind == reflect.Struct {
		unescapedKey, unescapeErr := d.converter.Unescape(rawKey)
		if unescapeErr != nil {
			return levelKeyChain.wrapError(unescapeErr, inputStr, val)
		}

		// TODO:
		// Struct info caching was already a priority, but the fact that this
		// parse() call now occurs within the "for field in fields..." loop
		// rather than outside it exacerbates the necessity.
		items, parseErr := d.structParser.parse(val)
		if parseErr != nil {
			return levelKeyChain.wrapError(parseErr, inputStr, val)
		}

		item, exists := items[unescapedKey]
		if !exists {
			if d.ignoreInvalidKeys {
				return nil
			}

			return levelKeyChain.newError("unknown key", inputStr, val)
		}

		childState := state.childWithSetMode(LevelValueList, item.setOpts)
		return d.decodeKeyChain(remainingChain, raw, item.val, childState)
	}

	if d.ignoreInvalidKeys {
		return nil
	}

	return levelKeyChain.newError("non-indexable key chain target", inputStr, val)
}
