package config2structure

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// global convenience types and functions

type submap map[string]interface{}

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

//import (
//	"fmt"
//	"testing"
//)
//
//type Config struct {
//	ConfigElmArr ConfigElm `c2s:"config"`
//}
//
//type ConfigElm struct {
//	Name        string      `c2s:"name"`
//	DynElmArray []DynConfig `c2s:"dynelement,dynamic=type"`
//}
//
//type DynConfig struct {
//	Type   string      `c2s:"type"`
//	Config interface{} `c2s:"config"`
//}
//
//func (dc *DynConfig) SetDynamicType(Type string) {
//	switch Type {
//	case "myfloat":
//		{
//			dc.Config = MyFloatConfig{}
//		}
//	case "myint":
//		{
//			dc.Config = MyIntConfig{}
//		}
//	case "null":
//		{
//			dc.Config = nil
//		}
//	}
//}
//
//type MyFloatConfig struct {
//	Key   string  `c2s:"keyfloat"`
//	Value float64 `c2s:"valuefloat"`
//}
//
//type MyIntConfig struct {
//	Key   string `c2s:"keyint"`
//	Value int    `c2s:"valueint"`
//}
//
//func TestYamlUnmarshal(t *testing.T) {
//	t.Parallel()
//	var datastruct = `
//        config:
//          name: "myconfig1"
//          dynelement:
//            - type: "myfloat"
//              config:
//                keyfloat: "chiavefloat"
//                valuefloat: 23.2
//            - type: "null"
//              config: ''
//       `
//	var cfg = Config{}
//	err := UnmarshalYaml([]byte(datastruct), &cfg)
//	if err != nil {
//		t.Fatalf("got an err: %s", err)
//	}
//
//	fmt.Printf("unmarshalled config: %#v \n", cfg)
//	//	fmt.Printf("subconf type: %s \n", reflect.TypeOf(cfg.DynElm[0].Config))
//	//	fmt.Printf("access float: %s \n", cfg.DynElm[0].Config.(MyFloatConfig).Key)
//
//	//	if cfg.Name != "myconfig" {
//	//		t.Errorf("string does not match: %s", cfg.Name)
//	//	}
//}
//
//func TestJsonlUnmarshal(t *testing.T) {
//	t.Parallel()
//	var datastruct = `
//        {
//          "config": {
//            "name": "myconfig1",
//            "dynelement": [
//              {
//                "type": "myfloat",
//                "config": {
//                  "keyfloat": "chiavefloat",
//                  "valuefloat": 23.2
//                }
//              },
//              {
//                "type": "null",
//                "config": ""
//              }
//            ]
//          }
//        }        `
//	var cfg = Config{}
//	err := UnmarshalJson([]byte(datastruct), &cfg)
//	if err != nil {
//		t.Fatalf("got an err: %s", err)
//	}
//
//	fmt.Printf("unmarshalled config: %#v \n", cfg)
//	//	fmt.Printf("subconf type: %s \n", reflect.TypeOf(cfg.DynElm.Config))
//	//	fmt.Printf("access float: %s \n", cfg.DynElm.Config.(MyFloatConfig).Key)
//
//	//	if cfg.Name != "myconfig" {
//	//		t.Errorf("string does not match: %s", cfg.Name)
//	//	}
//}
