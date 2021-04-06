# config2struct ( WIP )

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
      value: 1.3 
```

**config_int.yaml**
```yaml
config:
  name: "myconfig"
  dynelement:
    type: "myint"
    config:
      key: "keyname" 
      value: 1 
```

We define an abstract schema, implementing the c2s.DynamicStruct interface for the ***dynelement*** by defining the function ***SetDynamicType()***

**schema.go***
```go
type Config struct {
  Name string `c2s: "name"`
  DynElm DynConfig `c2s: "dynelement,dynamic"`  // we add the dynamic tag, required by c2s library
}

type DynConfig struct {
  Type string `c2s: "type"`
  Config interface{}
}

func (dc *DynConfig) SetDynamicType () {
  switch dc.Type {
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
  Key string `c2s: "key"`
  Value float64 `c2s: "value"`
}

type MyIntConfig struct {
  Key string `c2s: "key"`
  Value int `c2s: "value"`
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


