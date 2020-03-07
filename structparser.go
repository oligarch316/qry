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
	sTagSep        = ","
	sTagOmit       = "-"
	sTagOmitEscape = sTagOmit + sTagSep

	sTagBaseEmbed   = "embed"
	sTagBaseRequire = "require"

	sTagSetSep = "="
)

var sTagSetLevelNames = map[string]DecodeLevel{
	"query":     LevelQuery,
	"field":     LevelField,
	"key":       LevelKey,
	"valueList": LevelValueList,
	"value":     LevelValue,
}

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
	Tagged bool

	fieldInfo
	baseTagInfo
	setTagInfo
}

// DecodeName TODO
func (sfi StructFieldInfo) DecodeName() string {
	switch {
	case sfi.TagName != "":
		return sfi.TagName
	case sfi.Name == "":
		// TODO: Is this possible?
		return ""
	}

	// utf8 is fine because: https://golang.org/ref/spec#Source_code_representation
	r, n := utf8.DecodeRuneInString(sfi.Name)
	return string(unicode.ToLower(r)) + sfi.Name[n:]
}

func (sfi StructFieldInfo) wrapError(err error) StructFieldError {
	return StructFieldError{StructFieldInfo: sfi, err: err}
}

func (sfi StructFieldInfo) newError(msg string) StructFieldError {
	return sfi.wrapError(errors.New(msg))
}

type fieldInfo struct {
	Anonymous, Exported bool
	Name                string
	Type                reflect.Type
}

func (fi fieldInfo) String() string {
	if fi.Anonymous {
		return fmt.Sprintf("%s (%s)", fi.Type, fi.Type.Kind())
	}
	return fmt.Sprintf("%s %s (%s)", fi.Name, fi.Type, fi.Type.Kind())
}

type baseTagInfo struct {
	TagName                       string
	TagEmbed, TagOmit, TagRequire bool
}

func (bti *baseTagInfo) parse(raw string) error {
	switch raw {
	case "":
		// `qry:""`
		return errors.New("empty base tag")
	case sTagOmit:
		// `qry:"-"`
		bti.TagOmit = true
		return nil
	case sTagOmitEscape:
		// `qry:"-,"`
		bti.TagName = sTagOmit
		return nil
	}

	items := strings.Split(raw, sTagSep)
	bti.TagName, items = items[0], items[1:]

	for _, item := range items {
		switch item {
		case sTagBaseEmbed:
			bti.TagEmbed = true
		case sTagBaseRequire:
			bti.TagRequire = true
		default:
			return fmt.Errorf("invalid base tag directive '%s'", item)
		}
	}

	// Ensure no incompatible tag directives
	if bti.TagEmbed {
		switch {
		case bti.TagName != "":
			return errors.New("mutually exclusive base tag directive 'embed' and non-empty name")
		case bti.TagRequire:
			return errors.New("mutually exclusive base tag directives 'embed' and 'require'")
		}
	}

	return nil
}

type setTagInfo struct {
	explicitSetOpts SetOptionsMap
	defaultSetOpts  []SetOption
}

func (sti setTagInfo) SetOptions(defaultLevel DecodeLevel) SetOptionsMap {
	if len(sti.defaultSetOpts) < 1 {
		return sti.explicitSetOpts
	}

	res := SetOptionsMap{defaultLevel: sti.defaultSetOpts}
	for level, opts := range sti.explicitSetOpts {
		res[level] = opts
	}
	return res
}

func (sti *setTagInfo) parse(raw string) error {
	if raw == "" {
		// `qrySet:""`
		return errors.New("empty set tag")
	}

	items := strings.Split(raw, sTagSep)
	for _, item := range items {
		var (
			split = strings.SplitN(item, sTagSetSep, 2)
			err   error
		)

		if len(split) == 1 {
			err = sti.parseDefaultOpt(split[0])
		} else {
			err = sti.parseExplicitOpt(split[0], split[1])
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (sti *setTagInfo) parseDefaultOpt(rawOpt string) error {
	if opt := SetOption(rawOpt); opt.valid() {
		sti.defaultSetOpts = append(sti.defaultSetOpts, opt)
		return nil
	}
	return fmt.Errorf("invalid set tag option '%s'", rawOpt)
}

func (sti *setTagInfo) parseExplicitOpt(rawLevel, rawOpt string) error {
	var (
		level = sTagSetLevelNames[rawLevel]
		opt   = SetOption(rawOpt)
	)

	switch {
	case !level.validInput():
		return fmt.Errorf("invalid set tag level '%s'", rawLevel)
	case !opt.valid():
		return fmt.Errorf("invalid set tag option '%s'", rawOpt)
	}

	sti.explicitSetOpts[level] = append(sti.explicitSetOpts[level], opt)
	return nil
}

// ConfigStructParse TODO
type ConfigStructParse struct{ BaseTagName, SetTagName string }

type structItem struct {
	val reflect.Value
	setTagInfo
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

func (sp structParser) parseField(field reflect.StructField) (*StructFieldInfo, error) {
	var (
		res = &StructFieldInfo{
			fieldInfo: fieldInfo{
				Anonymous: field.Anonymous,
				Exported:  field.PkgPath == "",
				Name:      field.Name,
				Type:      field.Type,
			},
		}

		rawBaseTag, rawSetTag       string
		baseTagExists, setTagExists bool
	)

	if rawBaseTag, baseTagExists = field.Tag.Lookup(sp.BaseTagName); baseTagExists {
		if err := res.baseTagInfo.parse(rawBaseTag); err != nil {
			return nil, res.wrapError(err)
		}
	}

	if rawSetTag, setTagExists = field.Tag.Lookup(sp.SetTagName); setTagExists {
		if err := res.setTagInfo.parse(rawSetTag); err != nil {
			return nil, res.wrapError(err)
		}
	}

	// Ensure no incompatible settings between base and set tags
	if setTagExists {
		switch {
		case res.TagOmit:
			return nil, res.newError("mutually exclusive base tag name '-' (omit) and set tag options")
		case res.TagEmbed:
			return nil, res.newError("mutually exclusive base tag directive 'embed' and set tag options")
		}
	}

	res.Tagged = baseTagExists || setTagExists
	return res, nil
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
	return !sfi.Exported && (sp.checkUnmarshaler(sfi.Type) || sp.checkUnmarshaler(reflect.PtrTo(sfi.Type)))
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
			fieldInfo, fieldErr := sp.parseField(sType.Field(i))
			switch {
			case fieldErr != nil:
				return nil, fieldErr
			case fieldInfo.TagOmit:
				continue
			}

			// Explicit embed
			if fieldInfo.TagEmbed {
				if !fieldInfo.Anonymous && !fieldInfo.Exported {
					/*
						NOTE:
						We could *almost* disallow 'embed' on anonymous fields,
						allowing the implicit embed logic to handle all
						anonymous cases.

						However, for the gotcha, see the tests:
						> `pathological/anonymous exported embedded unmarshaler`
						> `pathological/anonymous unexported embedded unmarshaler`
					*/
					return nil, fieldInfo.newError("'embed' directive on non-anonymous unexported field")
				}

				switch fieldInfo.Type.Kind() {
				case reflect.Struct:
					// Unexported structs are fine as we can work with their zero values directly
					workList = append(workList, workItem.Field(i))
					continue
				case reflect.Ptr:
					if !fieldInfo.Exported {
						// Unexported pointers are not ok, as we need to set them if they're nil (zero value)
						return nil, fieldInfo.newError("'embed' directive on unexported pointer field")
					}

					elemType := fieldInfo.Type.Elem()
					if elemType.Kind() != reflect.Struct {
						return nil, fieldInfo.newError("'embed' directive on invalid type")
					}

					ptrVal := workItem.Field(i)
					if ptrVal.IsNil() {
						ptrVal.Set(reflect.New(elemType))
					}

					workList = append(workList, ptrVal.Elem())
					continue
				}

				return nil, fieldInfo.newError("'embed' directive on invalid type")
			}

			// Possible implicit embed
			if fieldInfo.Anonymous && !fieldInfo.Tagged && !sp.canUnmarshal(fieldInfo) {
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
					if fieldInfo.Exported {
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

			if !fieldInfo.Exported {
				// Return error if there's explicit intention to consider this field
				if fieldInfo.Tagged {
					return nil, fieldInfo.newError("tag on unexported field")
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
					val:        workItem.Field(i),
					setTagInfo: fieldInfo.setTagInfo,
				}
			}
		}
	}

	return res, nil
}
