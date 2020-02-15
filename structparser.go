package qry

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	sTagSep            = ","
	sTagOmit           = "-"
	sTagDirectiveEmbed = "embed"
)

// StructFieldError TODO
type StructFieldError struct {
	StructFieldInfo
	err error
}

// Unwrap TODO
func (sfe StructFieldError) Unwrap() error { return sfe.err }

func (sfe StructFieldError) Error() string {
	return fmt.Sprintf("%s: %s", sfe.StructFieldInfo, sfe.err)
}

// StructFieldInfo TODO
type StructFieldInfo struct {
	Name string
	Type reflect.Type

	anonymous  bool
	exported   bool
	tagName    string
	tagEmbed   bool
	tagSetOpts []SetOption
}

// DecodeName TODO
func (sfi StructFieldInfo) DecodeName() string {
	switch {
	case sfi.tagName != "":
		return sfi.tagName
	case sfi.Name == "":
		return ""
	}

	// utf8 is fine because: https://golang.org/ref/spec#Source_code_representation
	r, n := utf8.DecodeRuneInString(sfi.Name)
	return string(unicode.ToLower(r)) + sfi.Name[n:]
}

func (sfi StructFieldInfo) String() string {
	return fmt.Sprintf("[%s] %s (%s)", sfi.Name, sfi.Type, sfi.Type.Kind())
}

func (sfi StructFieldInfo) newError(msg string) StructFieldError {
	return StructFieldError{
		StructFieldInfo: sfi,
		err:             errors.New(msg),
	}
}

type structItem struct {
	setOpts []SetOption
	val     reflect.Value
}

type structParser string

func (sp structParser) loadFieldInfo(field reflect.StructField) (*StructFieldInfo, bool) {
	rawTag := field.Tag.Get(string(sp))
	if rawTag == sTagOmit {
		return nil, true
	}

	items := strings.Split(rawTag, sTagSep)
	res := &StructFieldInfo{
		Name:      field.Name,
		Type:      field.Type,
		anonymous: field.Anonymous,
		exported:  field.PkgPath == "",
		tagName:   items[0],
	}

	for _, item := range items[1:] {
		if item == sTagDirectiveEmbed {
			res.tagEmbed = true
			continue
		}
		res.tagSetOpts = append(res.tagSetOpts, SetOption(item))
	}

	return res, false
}

func (sp structParser) parse(val reflect.Value) (map[string]structItem, error) {
	var (
		workList = []reflect.Value{val}
		res      = make(map[string]structItem)
	)

	for len(workList) > 0 {
		// Pop next item (heuristic: guarenteed kind of reflect.Struct)
		workItem := workList[0]
		workList = workList[1:]

		var (
			sType   = workItem.Type()
			nFields = sType.NumField()
		)

		for i := 0; i < nFields; i++ {
			fieldInfo, omitted := sp.loadFieldInfo(sType.Field(i))
			if omitted {
				continue
			}

			// Explicit embed
			if fieldInfo.tagEmbed {
				switch {
				case fieldInfo.tagName != "":
					return nil, fieldInfo.newError("mutually exclusive non-empty name and embed directives")
				case fieldInfo.anonymous:
					return nil, fieldInfo.newError("unecessary embed directive on anonymous field")
				case !fieldInfo.exported:
					return nil, fieldInfo.newError("embed directive on unexported field")
				}

				switch fieldInfo.Type.Kind() {
				case reflect.Struct:
					workList = append(workList, workItem.Field(i))
					continue
				case reflect.Ptr:
					elemType := fieldInfo.Type.Elem()
					if elemType.Kind() != reflect.Struct {
						return nil, fieldInfo.newError("embed directive on pointer to invalid type")
					}

					ptrVal := workItem.Field(i)
					if ptrVal.IsNil() {
						ptrVal.Set(reflect.New(elemType))
					}

					workList = append(workList, ptrVal.Elem())
					continue
				}

				return nil, fieldInfo.newError("embed directive on invalid type")
			}

			// Possible implicit embed
			if fieldInfo.tagName == "" && fieldInfo.anonymous {
				// Safe to consider before the unmarshaler check given that a struct w/
				// anonymous unmarshaler field would already itself be an unmarshaler.
				// This is true for embedding to any depth.

				switch fieldInfo.Type.Kind() {
				case reflect.Struct:
					workList = append(workList, workItem.Field(i))
					continue
				case reflect.Ptr:
					if fieldInfo.exported && fieldInfo.Type.Elem().Kind() == reflect.Struct {
						ptrVal := workItem.Field(i)
						if ptrVal.IsNil() {
							ptrVal.Set(reflect.New(fieldInfo.Type.Elem()))
						}

						workList = append(workList, ptrVal.Elem())
						continue
					}
				}
			}

			// TODO: Check if ...
			// 1. fieldType is unmarshaler
			// 2. PtrTo(fieldType) is unmarshaler (current struct is heuristically addressalbe thus so are its fields)
			isUnmarshaler := false

			if !isUnmarshaler && !fieldInfo.exported {
				// Return error if intention to consider this field is explicit
				if fieldInfo.tagName != "" {
					return nil, fieldInfo.newError("non-empty name directive on invalid type")
				}

				// Otherwise skip
				continue
			}

			decodeName := fieldInfo.DecodeName()

			// TODO: More rigorous priority definition, see
			// https://golang.org/src/encoding/json/encode.go#L1196
			// for inspiration. depth > from tag > index sounds right.
			if _, collision := res[decodeName]; !collision {
				res[decodeName] = structItem{
					setOpts: fieldInfo.tagSetOpts,
					val:     workItem.Field(i),
				}
			}
		}
	}

	return res, nil
}
