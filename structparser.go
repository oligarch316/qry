package qry

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	stEmbed = "embed"
	stSep   = ","
)

type (
	structItem struct {
		setOpts []SetOption
		val     reflect.Value
	}
	structParser string
)

func formatFieldName(str string) string {
	if str == "" {
		return ""
	}

	// https://golang.org/ref/spec#Source_code_representation
	r, n := utf8.DecodeRuneInString(str)

	return string(unicode.ToLower(r)) + str[n:]
}

func parseRawTag(tag string) (name string, embed bool, setOpts []SetOption) {
	items := strings.Split(tag, stSep)
	name = items[0]

	if len(items) > 1 {
		for _, raw := range items[1:] {
			if raw == stEmbed {
				embed = true
				continue
			}

			setOpts = append(setOpts, SetOption(raw))
		}
	}

	return
}

func (sp structParser) parse(val reflect.Value) (map[string]structItem, error) {
	var (
		workList = []reflect.Value{val}
		res      = make(map[string]structItem)
	)

	for len(workList) > 0 {
		workItem := workList[0]
		workList = workList[1:]

		var (
			sType   = workItem.Type()
			nFields = sType.NumField()
		)

		for i := 0; i < nFields; i++ {
			var (
				field  = sType.Field(i)
				rawTag = field.Tag.Get(string(sp))
			)

			if rawTag == "-" {
				continue
			}

			tagName, tagEmbed, tagSetOpts := parseRawTag(rawTag)

			if tagEmbed && tagName != "" {
				return nil, fmt.Errorf("invalid struct tag: non-empty name (%q) and embed directives are mutually exclusive", tagName)
			}

			isExported := field.PkgPath == ""

			switch {
			case tagEmbed && isExported, tagName == "" && field.Anonymous:
				// For either of
				// - exported field with explicit embed tag
				// - anonymous field with no explicit name tag
				// that is ...

				// ... ignoring any pointer indirection ...
				tmpT := field.Type
				for tmpT.Kind() == reflect.Ptr {
					tmpT = tmpT.Elem()
				}

				// ... of kind struct ...
				if tmpT.Kind() == reflect.Struct {
					// ... add to the worklist and continue
					workList = append(workList, workItem.Field(i))
					continue
				}
			case !isExported:
				// Ignore all non-exported fields save in the above case
				continue
			}

			if tagName == "" {
				tagName = formatFieldName(field.Name)
			}

			// TODO: More rigorous priority definition, see
			// https://golang.org/src/encoding/json/encode.go#L1196
			// for inspiration. depth > from tag > index sounds right.
			if _, collision := res[tagName]; !collision {
				res[tagName] = structItem{
					setOpts: tagSetOpts,
					val:     workItem.Field(i),
				}
			}
		}
	}

	return res, nil
}
