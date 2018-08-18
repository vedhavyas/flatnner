package flatnner

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
)

type Node struct {
	Name  string
	Value string
}

func Flatten(data proto.Message) (nodes []Node, err error) {
	dv := reflect.ValueOf(data).Elem()
	return toNodes(dv.Type(), 0, dv)
}

func toNodes(t reflect.Type, i int, v reflect.Value) (nodes []Node, err error) {

	if !v.IsValid() {
		return nil, fmt.Errorf("invalid kind found: %v", v)
	}

	// unexported field. skipping
	if !v.CanInterface() {
		return nil, nil
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return toNodes(t, i, v.Elem())
	case reflect.Struct:
		nodes, err = flattenStruct(v)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten struct: %v", err)
		}
		return nodes, nil
	case reflect.Chan, reflect.UnsafePointer, reflect.Func, reflect.Uintptr:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

	return newNode(getName(t, i), fmt.Sprint(v.Interface())), nil
}

func flattenStruct(v reflect.Value) (nodes []Node, err error) {
	for i := 0; i < v.NumField(); i++ {
		n, err := toNodes(v.Type(), i, v.Field(i))
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, n...)
	}

	return nodes, nil
}

// getName returns the name of the field.
// must be of type struct else panics
func getName(t reflect.Type, i int) string {
	field := t.Field(i)
	tag, ok := field.Tag.Lookup("protobuf")
	if !ok || tag == "" {
		return field.Name
	}

	split := strings.Split(tag, ",")
	var name string
	for _, s := range split {
		if !strings.Contains(s, "name=") {
			continue
		}

		ns := strings.Split(s, "=")
		name = ns[1]
	}

	if name == "" {
		name = field.Name
	}

	return name
}

func newNode(name, value string) []Node {
	if name == "" {
		return nil
	}

	return []Node{
		{
			Name:  name,
			Value: value,
		},
	}
}
