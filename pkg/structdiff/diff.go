package structdiff

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	myers "github.com/pkg/diff"
)

// Diff is a struct differ.
type Diff struct {
	a      any
	b      any
	ignore []string
}

// DiffOption is functional opt for Diff.
type DiffOption func(*Diff)

// New constructs a new Diff.
func New(a, b any, opts ...DiffOption) *Diff {
	d := Diff{
		a: a,
		b: b,
	}

	for _, o := range opts {
		o(&d)
	}
	return &d
}

// WithIgnoreList updates the field path to ignore.
func WithIgnoreList(ig []string) DiffOption {
	return func(d *Diff) {
		d.ignore = ig
	}
}

// Get returns the diff between two structs.
func (d *Diff) Get() map[string]string {
	diffs := make(map[string]string)

	if d.a == nil || d.b == nil {
		return diffs
	}
	if reflect.TypeOf(d.a) != reflect.TypeOf(d.b) {
		return diffs
	}
	if reflect.TypeOf(d.a).Kind() != reflect.Struct {
		return diffs
	}

	ignoreList := map[string]struct{}{}
	for _, f := range d.ignore {
		ignoreList[f] = struct{}{}
	}

	chunks := getChunks(d.a, d.b, ignoreList)
	groupedChunks := make(map[string][]chunk)
	for _, c := range chunks {
		parent := c.Field
		if idx := strings.LastIndex(parent, "."); idx != -1 {
			parent = parent[:idx] // Extract the parent path.
		}
		groupedChunks[parent] = append(groupedChunks[parent], c)
	}
	for parent, chunks := range groupedChunks {
		var delta string
		if len(chunks) == 1 {
			delta = chunks[0].diff()
		} else {
			delta = getGroupedDiff(parent, chunks)
		}
		if !isEmptyDiff(delta) {
			diffs[parent] += delta
		}
	}
	return diffs
}

type chunk struct {
	Field string
	Typ   reflect.Kind
	From  any
	To    any

	ignore map[string]struct{}
}

func (c *chunk) diff() string {
	a := serialize(c.Typ, c.From, c.ignore)
	b := serialize(c.Typ, c.To, c.ignore)

	d, _ := doDiff(c.Field, a, b)
	return d
}

func getChunks(a, b any, ignore map[string]struct{}) []chunk {
	diffs := make([]chunk, 0)

	atyp := reflect.TypeOf(a)
	btyp := reflect.TypeOf(b)
	aval := reflect.ValueOf(a)
	bval := reflect.ValueOf(b)

	if atyp == nil && btyp == nil {
		return diffs
	}

	for i := range atyp.NumField() {
		field := atyp.Field(i)
		afield := aval.Field(i)
		bfield := bval.Field(i)

		if !afield.IsValid() || !bfield.IsValid() {
			continue
		}
		if _, ok := ignore[field.Name]; ok {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Slice:
			maxLen := max(bfield.Len(), afield.Len())
			for j := range maxLen {
				var fromItem, toItem any

				// Handle variation in slice length.
				if j < bfield.Len() {
					fromItem = bfield.Index(j)
					if bfield.CanInterface() {
						fromItem = bfield.Index(j).Interface()
					}
				}
				if j < afield.Len() {
					toItem = afield.Index(j)
					if afield.CanInterface() {
						toItem = afield.Index(j).Interface()
					}
				}

				diffs = append(diffs, chunk{
					Field:  field.Name,
					Typ:    field.Type.Elem().Kind(),
					From:   fromItem,
					To:     toItem,
					ignore: ignore,
				})
			}
		default:
			diffs = append(diffs, chunk{
				Field:  field.Name,
				Typ:    field.Type.Kind(),
				From:   bfield,
				To:     afield,
				ignore: ignore,
			})
		}
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

// getActualDiff returns +/- changes on the diff.
func getActualDiff(diff string) string {
	var out strings.Builder

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "@@") {
			continue
		}
		if line[0] == '+' || line[0] == '-' {
			out.WriteString(line + "\n")
		}
	}
	return out.String()
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

func getGroupedDiff(parent string, chunks []chunk) string {
	var (
		out       strings.Builder
		changeset strings.Builder
		added     int
		removed   int
	)

	for _, c := range chunks {
		actualDiff := getActualDiff(c.diff())
		added, removed = func() (int, int) {
			a, r := countChanges(actualDiff)
			return added + a, removed + r
		}()
		changeset.WriteString(actualDiff)
	}

	out.WriteString(fmt.Sprintf("--- a/%s\n", parent))
	out.WriteString(fmt.Sprintf("+++ b/%s\n", parent))
	out.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", removed, removed, added, added))
	out.WriteString(changeset.String())
	return out.String()
}

func countChanges(diff string) (int, int) {
	var (
		added   int
		removed int
	)

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '+':
			added++
		case '-':
			removed++
		}
	}
	return added, removed
}

func serialize(kind reflect.Kind, data any, ignore map[string]struct{}) string {
	if data == nil {
		return ""
	}
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
		return serializeSlice(data, ignore)
	case reflect.Ptr:
		return serializePtr(data, ignore)
	case reflect.Struct:
		return serializeStruct(data, ignore)
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

func serializeSlice(data any, ignore map[string]struct{}) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}
	if !v.IsValid() {
		return ""
	}

	var items []string
	for i := range v.Len() {
		val := v.Index(i)
		if !val.CanInterface() {
			items = append(items, fmt.Sprintf("%d: %x", i, val))
			continue
		}
		kind := reflect.TypeOf(val.Interface()).Kind()

		item := serialize(kind, val, ignore)
		if item != "" {
			items = append(items, item)
		}
	}
	return strings.Join(items, "\n")
}

func serializePtr(data any, ignore map[string]struct{}) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}
	if !v.IsValid() {
		return ""
	}
	if v.IsNil() {
		return ""
	}
	elm := v.Elem()

	if !v.CanInterface() {
		return serialize(elm.Kind(), elm, ignore)
	}
	return serialize(elm.Kind(), elm.Interface(), ignore)
}

func serializeStruct(data any, ignore map[string]struct{}) string {
	v, ok := data.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(data)
	}
	if !v.IsValid() {
		return ""
	}
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return ""
	}

	var items []string
	for i := range t.NumField() {
		var kind reflect.Kind

		field := t.Field(i)
		val := v.Field(i)

		if _, ok := ignore[field.Name]; ok {
			continue
		}

		if val.CanInterface() {
			kind = reflect.TypeOf(val.Interface()).Kind()
		} else {
			kind = field.Type.Kind()
		}

		item := serialize(kind, val, ignore)
		if item != "" {
			items = append(items, fmt.Sprintf("%s: %s", field.Name, item))
		}
	}
	return strings.Join(items, "\n")
}
