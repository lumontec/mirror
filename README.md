# mirror

mirror is a dead simple go library capable of reflecting complex dynamic json/yaml configurations into equivalent structures accessible at runtime. 


### Features

* **based on go tags**: simply map your struct members with equivalent go tags 
```go
type Config struct {
  Name string      `mirror:"name"`
}
```

* **produces detailed error report**: will output meaningful error messages in case any key is not matched 
```bash
        * detected unused keys: emails name
        * detected unused keys: medium twitter
        * map value not found for key: 
        * map value not found for key: med
        * map value not found for key: twit
        * missing `mirror` tag for struct field: Name
```

* **dynamic configuration**: supports parsing of complex kubernetes style declarative yaml configurations
Your config will is assertable at runtime:
```go
// Access dynamic fields through type assertion based on Type key
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
* **support for both json and yaml**
```go
config := Config{}

err := UnmarshalYaml([]byte(yamlContent), &config)
...
err := UnmarshalJson([]byte(jsonContent), &config)
...

```


## Example (simple and boring)

We simply map yaml configuration into equivalent struct:
**config_simple.yaml**
```yaml
config:
  name: "myconfig"
```

Declare the struct by mapping struct Name field to yaml name key:
**schema.go**
```go
type Config struct {
  Name string      `mirror:"name"`
}
```

Now mirror the config into the struct and just consume your config struct:
**main.go**
```go
import "github.com/lumontec/mirror"
...

// Read yaml file content 
yamlContent, err := ioutil.ReadFile(absPath)
...

// Mirror the yaml content into our Config object
config := Config{}
err := UnmarshalYaml([]byte(yamlContent), &config)

// Consume your configuration 
fmt.Printf("Config.Name=%s",config.Name)
```


## Example (dynamic configuration = more fun)

We demonstrate how to handle a dynamic configuration.
Key **dynamicelement** points to polymorphic subconfiguration wich can be of several types: "myfloat", "myint", as follow ...

**config_float.yaml**
```yaml
config:
  name: "myconfig"
  dynelement:
    type: "myfloat"
    config:
      keyfloat: "keyname" 
      valuefloat: 1.3 
```

**config_int.yaml**
```yaml
config:
  name: "myconfig"
  dynelement:
    type: "myint"
    config:
      keyint: "keyname" 
      valueint: 1 
```

We define an abstract schema, implementing the **mirror.DynamicStruct** interface for the **dynelement** by defining the function **SetDynamicType(string)** to return our specialized type depending on the string parameter value

**schema.go**
```go
type Config struct {
  Name string      `mirror:"name"`
  DynElm DynConfig `mirror:"dynelement,dynamic=type"`  // we add the dynamic selector, required by mirror library, sets selector key = type
}

type DynConfig struct {
  Type string        `mirror:"type"`
  Config interface{} `mirror:"config"`
}

// SetDynamicType implements mirror library DynamicStruct 
// interface, this method is called during the mirroring 
// process in order to set the Config interface to the 
// correct relative type
func (dc *DynConfig) SetDynamicType (Type string) {
  switch Type {
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
  Key string          `mirror:"keyfloat"`
  ValueFloat float64  `mirror:"valuefloat"`
}

type MyIntConfig struct {
  Key string    `mirror:"keyint"`
  ValueInt int  `mirror:"valueint"`
}
```

Then we can consume our configuration as such

**main.go**
```go
import "github.com/lumontec/mirror"
...

// Read yaml file content 
yamlContent, err := ioutil.ReadFile(absPath)
...

// Mirror the yaml content into our Config object
config := Config{}
err := UnmarshalYaml([]byte(yamlContent), &config)

// Access dynamic fields through type assertion based on Type key
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


