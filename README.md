# config2struct

This is a simple go library to parse complex dynamic json/yaml configurations into equivalent structures easily accessible at runtime

## Example

Key ***dynamicelement*** is polymorphic configuration wich can be of several types: "myfloat", "myint", ..

**config_float.yaml**
```yaml
config:
  name: "myconfig"
  dynelement:
    type: "myfloat"
    config:
      key: "keyname" 
      valuefloat: 1.3 
```

**config_int.yaml**
```yaml
config:
  name: "myconfig"
  dynelement:
    type: "myint"
    config:
      key: "keyname" 
      valueint: 1 
```

We define an abstract schema, implementing the c2s.DynamicStruct interface for the ***dynelement*** by defining the function ***SetDynamicType(string)***

**schema.go**
```go
type Config struct {
  Name string      `c2s:"name"`
  DynElm DynConfig `c2s:"dynelement,dynamic=type"`  // we add the dynamic selector, required by c2s library, sets selector key = type
}

type DynConfig struct {
  Type string        `c2s:"type"`
  Config interface{} `c2s:"config"`
}

func (dc *DynConfig) SetDynamicType (Type string) {
  switch Type {
  case "myfloat": 
    {
      dc.Config = MyFloatConfig
    }
  case "myint": 
    {
      dc.Config = MyIntConfig
    }
  }
}

type MyFloatConfig struct {
  Key string          `c2s:"key"`
  ValueFloat float64  `c2s:"valuefloat"`
}

type MyIntConfig struct {
  Key string    `c2s:"key"`
  ValueInt int  `c2s:"valueint"`
}
```

Then we can consume our configuration as such

**main.go**
```go
import "github.com/lumontec/config2struct"

... // Unmarshal the yaml into empty Config object

var data = `
config:
  name: "myconfig"
  dynelement:
    type: "myint"
    config:
      key: "keyname" 
      value: 1 
`
config := Config{}
err := UnmarshalYaml([]byte(data), &config)

... // Access dynamic fields based on type

switch config.DynElm.Type {
  case "myfloat": 
    {
      floatcfg := config.DynElm.Config(MyFloatConfig)
      // Do stuff with my float configuration
    }
  case "myint": 
    {
      intcfg := config.DynElm.Config(MyIntConfig)
      // Do stuff with my int configuration
    }
    ...
}

```


