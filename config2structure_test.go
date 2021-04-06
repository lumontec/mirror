package config2structure

import (
	//	"encoding/json"
	//	"io"
	//	"reflect"
	//	"sort"
	//	"strings"
	"fmt"
	"testing"
)

type Config struct {
	Name   string    `c2s:"name"`
	DynElm DynConfig `c2s:"dynelement,dynamic"`
}

type DynConfig struct {
	Type   string      `c2s:"type"`
	Config interface{} `c2s:"config"`
}

func (dc DynConfig) SetDynamicType() {
	switch dc.Type {
	case "myfloat":
		{
			dc.Config = MyFloatConfig{}
		}
	case "myint":
		{
			dc.Config = MyIntConfig{}
		}
	}
}

type MyFloatConfig struct {
	Key   string  `c2s:"keyfloat"`
	Value float64 `c2s:"valuefloat"`
}

type MyIntConfig struct {
	Key   string `c2s:"keyin"`
	Value int    `c2s:"valueint"`
}

//func TestStringDecode(t *testing.T) {
//	t.Parallel()
//	var datastring = `
//        key: "ciao"
//        `
//	cfg := ""
//	err := UnmarshalYaml([]byte(datastring), &cfg)
//	if err != nil {
//		t.Fatalf("got an err: %s", err)
//	}
//
//	fmt.Printf("unmarshalled config: %#v \n", cfg)
//
//	if cfg != "ciao" {
//		t.Errorf("string does not match: %s", cfg)
//	}
//}

func TestStructDecode(t *testing.T) {
	t.Parallel()
	var datastruct = `
        config:
          name: "myconfig"
          dynelement:
            type: "myint"
            config:
              keyint: "chiave"
              valueint: 23
        `
	var cfg = Config{}
	err := UnmarshalYaml([]byte(datastruct), &cfg)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	fmt.Printf("unmarshalled config: %#v \n", cfg)

	//	if cfg.Name != "myconfig" {
	//		t.Errorf("string does not match: %s", cfg.Name)
	//	}
}
