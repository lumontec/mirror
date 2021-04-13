package config2structure

import (
	//	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// global convenience types and functions

type smap map[string]interface{}

func TestDecodeBool(t *testing.T) {
	t.Parallel()

	type Test struct {
		name string
		data interface{}
		want bool
		err  bool
	}

	tests_ok := []struct {
		name string
		data interface{}
		want bool
		err  bool
	}{
		{"bool 1", true, true, false},
		{"bool 2", false, false, false},
		{"bool 3", 1, false, true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var value bool
			val := reflect.ValueOf(&value).Elem()
			err := decodeBool(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, value)
			}
		})
	}
}

func TestDecodeInt(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want int
		err  bool
	}{
		{"int 1", 1, 1, false},
		{"int 2", 2147483649, 2147483649, false},
		{"int 3", -2147483649, -2147483649, false},
		{"int 4", int64(1), 1, false},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var value int
			val := reflect.ValueOf(&value).Elem()
			err := decodeInt(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, value)
			}

		})
	}
}

func TestDecodeUint(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want uint
		err  bool
	}{
		{"uint 1", uint(1), 1, false},
		{"uint 2", uint(2147483649), 2147483649, false},
		{"uint 3", int(2147483649), 2147483649, true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var value uint
			val := reflect.ValueOf(&value).Elem()
			err := decodeUint(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, value)
			}

		})
	}
}

func TestDecodeFloat(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want float64
		err  bool
	}{
		{"float 1", 1.0, 1.0, false},
		{"float 2", float64(10.0), float64(10.0), false},
		{"float 3", float32(10.0), 10.0, false},
		{"float 4", int(10.0), 9.0, true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var value float64
			val := reflect.ValueOf(&value).Elem()
			err := decodeFloat(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, value)
			}

		})
	}
}

func TestDecodeString(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want string
		err  bool
	}{
		{"string 1", "string", "string", false},
		{"string 2", 1, "string", true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var value string
			val := reflect.ValueOf(&value).Elem()
			err := decodeString(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, value)
			}

		})
	}
}

func TestDecodePtr(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want interface{}
		ret  bool
		err  bool
	}{
		{"ptr 1", int(1), int(1), false, false},
		{"ptr 2", "string", "string", false, false},
		{"ptr 3", nil, nil, true, false},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// This generates an empty value ptr on copy of tt.data type
			value := tt.data
			refval := reflect.ValueOf(&value).Elem()
			if tt.data != nil {
				refval.Set(reflect.Zero(reflect.TypeOf(tt.data)))
			} else {
				value = nil
			}
			ptr := &value

			valptr := reflect.ValueOf(ptr)
			ret, err := decodePtr(tt.name, tt.data, valptr)

			if tt.err {
				assert.Error(t, err)
			} else {

				assert.Equal(t, tt.want, *ptr)
				assert.Equal(t, tt.ret, ret)
			}

		})
	}
}

func TestDecodeSlice(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want []int
		err  bool
	}{
		{"slice 1", []int{1, 1}, []int{1, 1}, false},
		{"slice 2", []int{}, []int{}, false},
		{"slice 3", nil, nil, false},
		{"slice 4", 100, []int{}, true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var slc []int

			val := reflect.ValueOf(&slc).Elem()
			err := decodeSlice(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, val.Interface())
			}

		})
	}
}

func TestDecodeArray(t *testing.T) {
	t.Parallel()

	tests_ok := []struct {
		name string
		data interface{}
		want [2]int
		err  bool
	}{
		{"slice 1", [2]int{1, 1}, [2]int{1, 1}, false},
		{"slice 2", [2]int{}, [2]int{}, false},
		{"slice 3", [1]int{}, [2]int{}, false},
		{"slice 4", 100, [2]int{1, 1}, true},
	}
	for _, tt := range tests_ok {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var slc [2]int

			val := reflect.ValueOf(&slc).Elem()
			err := decodeArray(tt.name, tt.data, val)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, val.Interface())
			}

		})
	}
}

func TestDecodeStructFromMapSimple(t *testing.T) {

	type ExtraTyp struct {
		Twitter string `c2s:"twitter"`
	}

	type Person struct {
		Name   string   `c2s:"name"`
		Age    int      `c2s:"age"`
		Emails []string `c2s:"emails"`
		Extra  ExtraTyp `c2s:"extra"`
	}

	input := map[string]interface{}{
		"name":   "lumontec",
		"age":    91,
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "lumontec",
		},
	}

	var want = Person{
		Name:   "lumontec",
		Age:    91,
		Emails: []string{"one", "two", "three"},
		Extra: ExtraTyp{
			Twitter: "lumontec",
		},
	}

	var result Person
	val := reflect.ValueOf(&result).Elem()
	err := decodeStructFromMap("struct", reflect.Indirect(reflect.ValueOf(input)), val)

	assert.NoError(t, err)
	assert.Equal(t, want, val.Interface())
}

func TestDecodeStructFromMapErrors(t *testing.T) {

	type ExtraTyp struct {
		Twitter string `c2s:"twit"`
		Medium  string `c2s:"med"`
	}

	type Person struct {
		Name   string   `c2:"name"`
		Age    int      `c2s:"age"`
		Emails []string `c2s:"emails"`
		Extra  ExtraTyp `c2s:"extra"`
	}

	input := map[string]interface{}{
		"name":   "lumontec",
		"age":    91,
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "lumontec",
			"medium":  "lumontec",
		},
	}

	wanterr := &Error{
		Errors: []string{"missing `c2s` tag for struct field: Name", "map value not found for key: ", "map value not found for key: twit", "map value not found for key: med"},
	}

	var result Person
	val := reflect.ValueOf(&result).Elem()
	err := decodeStructFromMap("struct", reflect.Indirect(reflect.ValueOf(input)), val)

	assert.Error(t, err)
	assert.Equal(t, wanterr, err)
}
