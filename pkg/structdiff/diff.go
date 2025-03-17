package structdiff

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	myers "github.com/pkg/diff"
)

type chunk struct {
	Field string
	Typ   reflect.Kind
	From  any
	To    any
}

func (c *chunk) diff() string {
	a := serialize(c.Typ, c.From)
	b := serialize(c.Typ, c.To)

	d, _ := doDiff(c.Field, a, b)
	return d
}

// Get returns the diff between two structs.
func Get(a, b any) map[string]string {
	diffs := make(map[string]string)

	if a == nil || b == nil {
		return diffs
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return diffs
	}
	if reflect.TypeOf(a).Kind() != reflect.Struct {
		return diffs
	}

	chunks := getChunks(a, b)
	for _, c := range chunks {
		d := c.diff()
		if !isEmptyDiff(d) {
			diffs[c.Field] = d
		}
	}
	return diffs
}

func getChunks(a, b any) []chunk {
	diffs := make([]chunk, 0)

	atyp := reflect.TypeOf(a)
	aval := reflect.ValueOf(a)
	bval := reflect.ValueOf(b)

	for i := range atyp.NumField() {
		field := atyp.Field(i)

		diffs = append(diffs, chunk{
			Field: field.Name,
			Typ:   field.Type.Kind(),
			From:  bval.Field(i),
			To:    aval.Field(i),
		})
	}
	return diffs
}

// doDiff calculates the diff between two strings using myers diff algorithm.
func doDiff(file, a, b string) (string, error) {
	w := &bytes.Buffer{}

	err := myers.Text("a/"+file, "b/"+file, a, b, w)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

// isEmptyDiff checks if the diff contains any meaningful changes.
func isEmptyDiff(diff string) bool {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "@@") {
			continue
		}
		if line[0] == '+' || line[0] == '-' {
			return false
		}
	}
	return true
}

func serialize(kind reflect.Kind, data any) string {
	switch kind {
	case reflect.Int:
		return serializeInt(data)
	case reflect.String:
		return serializeStr(data)
	case reflect.Bool:
		return serializeBool(data)
	case reflect.Float32, reflect.Float64:
		return serializeFloat(data)
	case reflect.Slice:
		return serializeSlice(data)
	case reflect.Ptr:
		return serializePtr(data)
	case reflect.Struct:
		return serializeStruct(data)
	}
	return ""
}

func serializeInt(data any) string {
	return fmt.Sprintf("%d", data)
}

func serializeStr(data any) string {
	return fmt.Sprintf("%s", data)
}

func serializeBool(data any) string {
	return fmt.Sprintf("%t", data)
}

func serializeFloat(data any) string {
	return fmt.Sprintf("%g", data)
}

func serializeSlice(data any) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}

	var items []string
	for i := range v.Len() {
		val := v.Index(i)
		if !val.CanInterface() {
			items = append(items, fmt.Sprintf("%d: %x", i, val))
			continue
		}
		kind := reflect.TypeOf(val.Interface()).Kind()

		item := serialize(kind, val)
		if item != "" {
			items = append(items, fmt.Sprintf("%d.%s", i, item))
		}
	}
	return strings.Join(items, "\n")
}

func serializePtr(data any) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}

	if v.IsNil() {
		return ""
	}
	elm := v.Elem()

	if !v.CanInterface() {
		return serialize(elm.Kind(), elm)
	}
	return serialize(elm.Kind(), elm.Interface())
}

func serializeStruct(data any) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}
	t := v.Type()

	var items []string
	for i := range v.Type().NumField() {
		var kind reflect.Kind

		field := t.Field(i)
		val := v.Field(i)

		if val.CanInterface() {
			kind = reflect.TypeOf(val.Interface()).Kind()
		} else {
			kind = field.Type.Kind()
		}

		item := serialize(kind, val)
		if item != "" {
			items = append(items, fmt.Sprintf("%s: %s", field.Name, item))
		}
	}
	return strings.Join(items, "\n")
}
