package flatnner

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ex = struct {
	ValueA string `protobuf:"bytes,1,opt,name=valueA,proto3" json:"valueA,omitempty"`
	ValueB int
	ValueC bool
	ValueD float32

	// unexported
	valueC bool
}{
	ValueA: "Hello",
	ValueB: 1,
	ValueC: true,
	ValueD: 22.2,
}

func Test_getName(t *testing.T) {
	tests := []struct {
		i    int
		want string
	}{
		{
			i:    0,
			want: "valueA",
		},

		{
			i:    1,
			want: "ValueB",
		},
	}

	st := reflect.ValueOf(ex).Type()
	for _, c := range tests {
		got := getName(st, c.i)
		if got != c.want {
			t.Fatalf("%s != %s", got, c.want)
		}
	}
}

func Test_toNodes_simple_types(t *testing.T) {
	st := reflect.ValueOf(ex).Type()

	tests := []interface{}{
		true,
		int(1),
		int8(2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
		float32(1.0),
		float64(2.0),
		"abcde",
	}

	for _, c := range tests {
		ns, err := toNodes(st, 0, reflect.ValueOf(c))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, len(ns), 1)
		assert.Equal(t, ns[0].Name, "valueA")
		assert.Equal(t, ns[0].Value, fmt.Sprintf("%v", c))
	}
}

func Test_toNodes_simple_struct(t *testing.T) {
	sv := reflect.ValueOf(ex)
	ns, err := flattenStruct(sv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, len(ns), 4)
	want := []Node{
		{
			Name:  "valueA",
			Value: "Hello",
		},

		{
			Name:  "ValueB",
			Value: "1",
		},

		{
			Name:  "ValueC",
			Value: "true",
		},

		{
			Name:  "ValueD",
			Value: "22.2",
		},
	}

	assert.Equal(t, ns, want)
}

func Test_flatten_nested_struct(t *testing.T) {
	nestEx := struct {
		Ex struct {
			ValueA string
			Ex1    struct {
				ValueE string
			}
		}
		ValueB int
		ValueC bool
		ValueD float32
	}{
		Ex: struct {
			ValueA string
			Ex1    struct{ ValueE string }
		}{ValueA: "abc", Ex1: struct{ ValueE string }{ValueE: "cde"}},
		ValueB: 1,
		ValueC: true,
		ValueD: 22.2,
	}
	sv := reflect.ValueOf(nestEx)
	ns, err := flattenStruct(sv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, 5, len(ns))
	want := []Node{
		{
			Name:  "ValueA",
			Value: "abc",
		},

		{
			Name:  "ValueE",
			Value: "cde",
		},

		{
			Name:  "ValueB",
			Value: "1",
		},

		{
			Name:  "ValueC",
			Value: "true",
		},

		{
			Name:  "ValueD",
			Value: "22.2",
		},
	}

	assert.Equal(t, ns, want)
}

func Test_toNodes_nested_struct_pointer(t *testing.T) {
	nestEx := struct {
		Ex struct {
			ValueA string
			Ex1    *struct {
				ValueE string
			}
		}
		ValueB int
		ValueC bool
		ValueD float32
	}{
		Ex: struct {
			ValueA string
			Ex1    *struct{ ValueE string }
		}{ValueA: "abc", Ex1: &struct{ ValueE string }{ValueE: "cde"}},
		ValueB: 1,
		ValueC: true,
		ValueD: 22.2,
	}

	sv := reflect.ValueOf(nestEx)
	ns, err := flattenStruct(sv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, 5, len(ns))
	want := []Node{
		{
			Name:  "ValueA",
			Value: "abc",
		},

		{
			Name:  "ValueE",
			Value: "cde",
		},

		{
			Name:  "ValueB",
			Value: "1",
		},

		{
			Name:  "ValueC",
			Value: "true",
		},

		{
			Name:  "ValueD",
			Value: "22.2",
		},
	}

	assert.Equal(t, ns, want)
}

func Test_toNodes_slice_maps(t *testing.T) {
	tests := []interface{}{
		[]int{1, 2, 3},
		[]uint8{1, 2, 3},
		[]string{"a", "b", "c"},
		[]byte{1, 2, 255},
		[][]int{{1}, {2}},
		map[string]string{"1": "2"},
		map[byte]interface{}{1: 1},
		map[byte]interface{}{1: "abc"},
		map[byte]interface{}{1: true},
		map[byte]interface{}{1: []int{12}},
		map[byte]interface{}{1: [][]int{{1}, {2}}},
		map[byte]interface{}{1: map[string]string{"1": "2"}},
	}

	st := reflect.ValueOf(ex).Type()
	for _, c := range tests {
		ns, err := toNodes(st, 0, reflect.ValueOf(c))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, 1, len(ns))
		assert.Equal(t, fmt.Sprint(c), ns[0].Value)
	}
}

func Test_toNodes_slice_struct(t *testing.T) {
	test := []interface{}{
		struct{ A string }{A: "a"},
		&struct{ B int }{B: 1},
		struct{ C *struct{ D int } }{C: &struct{ D int }{D: 1}},
		&struct{ E *struct{ F int } }{E: &struct{ F int }{F: 1}},
		&struct{ G struct{ H int } }{G: struct{ H int }{H: 1}},
	}

	want := []Node{
		{
			Name:  "A",
			Value: "a",
		},
		{
			Name:  "B",
			Value: "1",
		},
		{
			Name:  "D",
			Value: "1",
		},
		{
			Name:  "F",
			Value: "1",
		},
		{
			Name:  "H",
			Value: "1",
		},
	}

	v := reflect.ValueOf(&test)
	nodes, err := toNodes(v.Type(), 0, v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, 5, len(nodes))
	assert.Equal(t, want, nodes)
}
