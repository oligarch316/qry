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
	if sfi.anonymous {
		return fmt.Sprintf("%s (%s)", sfi.Type, sfi.Type.Kind())
	}
	return fmt.Sprintf("%s %s (%s)", sfi.Name, sfi.Type, sfi.Type.Kind())
}

func (sfi StructFieldInfo) newError(msg string) StructFieldError {
	return StructFieldError{
		StructFieldInfo: sfi,
		err:             errors.New(msg),
	}
}

// ConfigStructParse TODO
type ConfigStructParse struct{ TagName string }

type structItem struct {
	setOpts []SetOption
	val     reflect.Value
}

type structParser struct {
	ConfigStructParse
	checkUnmarshaler func(reflect.Type) bool
}

func newStructParser(cfg ConfigStructParse, checkUnmarshaler func(reflect.Type) bool) *structParser {
	return &structParser{
		ConfigStructParse: cfg,
		checkUnmarshaler:  checkUnmarshaler,
	}
}

func (sp structParser) loadFieldInfo(field reflect.StructField) (*StructFieldInfo, bool) {
	rawTag := field.Tag.Get(sp.TagName)
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
		// TODO: is continue on empty string worthwhile here?
		// - no change in correctness, just eliminate a useless arg to childWithSetMode()
		// - special mention only because it is technically part of the spec given
		//   parsing a key of name "-" requires a tag `qry:"-,"` (like encoding/json)
		if item == sTagDirectiveEmbed {
			res.tagEmbed = true
			continue
		}
		res.tagSetOpts = append(res.tagSetOpts, SetOption(item))
	}

	return res, false
}

func (sp structParser) canUnmarshal(sfi *StructFieldInfo) bool {
	/*
			NOTE:
			1. We cannot take the .Interface() of unexported fields, so for our
		       purposes, all unexported fields are not unmarshalers
			2. Root struct heuristacally addressable
		       ==implies=> field is addressable
		       ==implies=> PtrTo check required
	*/
	return !sfi.exported && (sp.checkUnmarshaler(sfi.Type) || sp.checkUnmarshaler(reflect.PtrTo(sfi.Type)))
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
				case !fieldInfo.anonymous && !fieldInfo.exported:
					/*
						NOTE:
						We could *almost* disallow 'embed' on anonymous fields,
						allowing the implicit embed logic to handle all
						anonymous cases.

						However, for the gotcha, see the tests:
						> `pathological/anonymous exported embedded unmarshaler`
						> `pathological/anonymous unexported embedded unmarshaler`
					*/
					return nil, fieldInfo.newError("embed directive on non-anonymous unexported field")
				}

				switch fieldInfo.Type.Kind() {
				case reflect.Struct:
					workList = append(workList, workItem.Field(i))
					continue
				case reflect.Ptr:
					if !fieldInfo.exported {
						return nil, fieldInfo.newError("embed directive on unexported pointer field")
					}

					elemType := fieldInfo.Type.Elem()
					if elemType.Kind() != reflect.Struct {
						return nil, fieldInfo.newError("embed directive on invalid type")
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
			if fieldInfo.anonymous && fieldInfo.tagName == "" && !sp.canUnmarshal(fieldInfo) {
				/*
					NOTE:
					We could *almost* dispense with the unmarshaler check, given
					anonymous field is unmarshaler ==> struct is unmarshaler.

					However, for the gotcha, see the test:
					> `pathological/anonymous exported unmarshaler`
				*/

				switch fieldInfo.Type.Kind() {
				case reflect.Struct:
					workList = append(workList, workItem.Field(i))
					continue
				case reflect.Ptr:
					if fieldInfo.exported {
						elemType := fieldInfo.Type.Elem()
						if elemType.Kind() == reflect.Struct {
							ptrVal := workItem.Field(i)
							if ptrVal.IsNil() {
								ptrVal.Set(reflect.New(elemType))
							}

							workList = append(workList, ptrVal.Elem())
							continue
						}
					}
				}
			}

			if !fieldInfo.exported {
				// Return error if there's explicit intention to consider this field
				if fieldInfo.tagName != "" {
					return nil, fieldInfo.newError("non-empty name directive on unexported field")
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
