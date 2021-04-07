package config2structure

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"
	"strconv"
	"strings"
	//	"sort"
)

type DynamicStruct interface {
	SetDynamicType(Type string)
}

// Unmarshal full yaml into the configuration structure
func UnmarshalYaml(data []byte, config interface{}) error {

	rawmap := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(data), rawmap)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %s", err)
	}

	err = decodeMapLevels(rawmap["config"], config)
	if err != nil {
		return fmt.Errorf("decode map: %s", err)
	}

	return nil
}

// Unmarshal full yaml into the configuration structure
func UnmarshalJson(data []byte, config interface{}) error {

	rawmap := make(map[string]interface{})

	err := json.Unmarshal([]byte(data), &rawmap)
	if err != nil {
		return fmt.Errorf("unmarshal json: %s", err)
	}

	err = decodeMapLevels(rawmap["config"], config)
	if err != nil {
		return fmt.Errorf("decode map: %s", err)
	}

	return nil
}

// decodeMapLevel decodes a single map level into the config structure
func decodeMapLevels(input interface{}, output interface{}) error {
	return decode("", input, reflect.ValueOf(output).Elem())
}

// Decodes an unknown data type into a specific reflection value.
func decode(name string, input interface{}, outVal reflect.Value) error {
	var inputVal reflect.Value
	if input != nil {
		inputVal = reflect.ValueOf(input)

		// We need to check here if input is a typed nil. Typed nils won't
		// match the "input == nil" below so we check that here.
		if inputVal.Kind() == reflect.Ptr && inputVal.IsNil() {
			input = nil
		}
	}

	if input == nil {
		return fmt.Errorf("input is nil")
	}

	if !inputVal.IsValid() {
		return fmt.Errorf("input is invalid")
	}

	var err error
	outputKind := getKind(outVal)
	switch outputKind {
	case reflect.Bool:
		err = decodeBool(name, input, outVal)
	case reflect.Interface:
		err = decodeBasic(name, input, outVal)
	case reflect.String:
		err = decodeString(name, input, outVal)
	case reflect.Int:
		err = decodeInt(name, input, outVal)
	case reflect.Uint:
		err = decodeUint(name, input, outVal)
	case reflect.Float32:
		err = decodeFloat(name, input, outVal)
	case reflect.Struct:
		err = decodeStruct(name, input, outVal)
		//	case reflect.Map:
		//		err = decodeMap(name, input, outVal)
	case reflect.Ptr:
		_, err = decodePtr(name, input, outVal)
	case reflect.Slice:
		err = decodeSlice(name, input, outVal)
	case reflect.Array:
		err = decodeArray(name, input, outVal)
		//	case reflect.Func:
		//		err = decodeFunc(name, input, outVal)
	default:
		// If we reached this point then we weren't able to decode it
		return fmt.Errorf("%s: unsupported type: %s", name, outputKind)
	}

	return err
}

func decodePtr(name string, data interface{}, val reflect.Value) (bool, error) {
	// If the input data is nil, then we want to just set the output
	// pointer to be nil as well.
	isNil := data == nil
	if !isNil {
		switch v := reflect.Indirect(reflect.ValueOf(data)); v.Kind() {
		case reflect.Chan,
			reflect.Func,
			reflect.Interface,
			reflect.Map,
			reflect.Ptr,
			reflect.Slice:
			isNil = v.IsNil()
		}
	}
	if isNil {
		if !val.IsNil() && val.CanSet() {
			nilValue := reflect.New(val.Type()).Elem()
			val.Set(nilValue)
		}

		return true, nil
	}

	// Create an element of the concrete (non pointer) type and decode
	// into that. Then set the value of the pointer to this type.
	valType := val.Type()
	valElemType := valType.Elem()
	if val.CanSet() {
		realVal := val
		if realVal.IsNil() {
			realVal = reflect.New(valElemType)
		}

		if err := decode(name, data, reflect.Indirect(realVal)); err != nil {
			return false, err
		}

		val.Set(realVal)
	} else {
		if err := decode(name, data, reflect.Indirect(val)); err != nil {
			return false, err
		}
	}
	return false, nil
}

// This decodes a basic type (bool, int, string, etc.) and sets the
// value to "data" of that type.
func decodeBasic(name string, data interface{}, val reflect.Value) error {

	fmt.Printf("decodeBasic: data: %#v \n", data)
	fmt.Printf("decodeBasic: typeofdata: %s \n", reflect.TypeOf(data))

	if val.IsValid() && val.Elem().IsValid() {

		fmt.Println("decodeBasic: value is valid")

		elem := val.Elem()

		// If we can't address this element, then its not writable. Instead,
		// we make a copy of the value (which is a pointer and therefore
		// writable), decode into that, and replace the whole value.
		copied := false
		if !elem.CanAddr() {
			copied = true

			// Make *T
			copy := reflect.New(elem.Type())

			// *T = elem
			copy.Elem().Set(elem)

			// Set elem so we decode into it
			elem = copy
		}

		// Decode. If we have an error then return. We also return right
		// away if we're not a copy because that means we decoded directly.
		if err := decode(name, data, elem); err != nil || !copied {
			return err
		}

		// If we're a copy, we need to set te final result
		val.Set(elem.Elem())
		return nil
	}

	dataVal := reflect.ValueOf(data)

	// If the input data is a pointer, and the assigned type is the dereference
	// of that exact pointer, then indirect it so that we can assign it.
	// Example: *string to string
	if dataVal.Kind() == reflect.Ptr && dataVal.Type().Elem() == val.Type() {
		dataVal = reflect.Indirect(dataVal)
	}

	if !dataVal.IsValid() {
		dataVal = reflect.Zero(val.Type())
	}

	dataValType := dataVal.Type()
	if !dataValType.AssignableTo(val.Type()) {
		return fmt.Errorf(
			"'%s' expected type '%s', got '%s'",
			name, val.Type(), dataValType)
	}

	//	val.Set(dataVal)
	return nil
}

func decodeBool(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch {
	case dataKind == reflect.Bool:
		val.SetBool(dataVal.Bool())
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func decodeInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetInt(dataVal.Int())
	case dataKind == reflect.Uint:
		val.SetInt(int64(dataVal.Uint()))
	case dataKind == reflect.Float32:
		val.SetInt(int64(dataVal.Float()))
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func decodeUint(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		i := dataVal.Int()
		val.SetUint(uint64(i))
	case dataKind == reflect.Uint:
		val.SetUint(dataVal.Uint())
	case dataKind == reflect.Float32:
		f := dataVal.Float()
		val.SetUint(uint64(f))
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func decodeFloat(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetFloat(float64(dataVal.Int()))
	case dataKind == reflect.Uint:
		val.SetFloat(float64(dataVal.Uint()))
	case dataKind == reflect.Float32:
		val.SetFloat(dataVal.Float())
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	return nil
}

func decodeSlice(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	valType := val.Type()
	valElemType := valType.Elem()
	sliceType := reflect.SliceOf(valElemType)

	// If we have a non array/slice type then we first attempt to convert.
	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
		return fmt.Errorf(
			"'%s': source data must be an array or slice, got %s", name, dataValKind)
	}

	// If the input value is nil, then don't allocate since empty != nil
	if dataVal.IsNil() {
		return nil
	}

	valSlice := val
	if valSlice.IsNil() {
		// Make a new slice to hold our result, same size as the original data.
		valSlice = reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())
	}

	// Accumulate any errors
	errors := make([]string, 0)

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		for valSlice.Len() <= i {
			valSlice = reflect.Append(valSlice, reflect.Zero(valElemType))
		}
		currentField := valSlice.Index(i)

		fieldName := name + "[" + strconv.Itoa(i) + "]"
		if err := decode(fieldName, currentData, currentField); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	// Finally, set the value to the slice we built up
	val.Set(valSlice)

	// If there were errors, we return those
	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

func decodeArray(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	valType := val.Type()
	valElemType := valType.Elem()
	arrayType := reflect.ArrayOf(valType.Len(), valElemType)

	valArray := val

	if valArray.Interface() == reflect.Zero(valArray.Type()).Interface() {
		// Check input type
		if dataValKind != reflect.Array && dataValKind != reflect.Slice {
			return fmt.Errorf(
				"'%s': source data must be an array or slice, got %s", name, dataValKind)

		}
		if dataVal.Len() > arrayType.Len() {
			return fmt.Errorf(
				"'%s': expected source data to have length less or equal to %d, got %d", name, arrayType.Len(), dataVal.Len())

		}

		// Make a new array to hold our result, same size as the original data.
		valArray = reflect.New(arrayType).Elem()
	}

	// Accumulate any errors
	errors := make([]string, 0)

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valArray.Index(i)

		fieldName := name + "[" + strconv.Itoa(i) + "]"
		if err := decode(fieldName, currentData, currentField); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	// Finally, set the value to the array we built up
	val.Set(valArray)

	// If there were errors, we return those
	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

func decodeString(name string, data interface{}, val reflect.Value) error {

	fmt.Println("decodeString")

	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	converted := true
	switch {
	case dataKind == reflect.String:
		val.SetString(dataVal.String())
	default:
		converted = false
	}

	if !converted {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}
	return nil
}

func decodeStruct(name string, data interface{}, val reflect.Value) error {

	fmt.Println("decoding struct")

	dataVal := reflect.Indirect(reflect.ValueOf(data))

	// If the type of the value to write to and the data match directly,
	// then we just set it directly instead of recursing into the structure.
	if dataVal.Type() == val.Type() {
		val.Set(dataVal)
		return nil
	}

	dataValKind := dataVal.Kind()
	switch dataValKind {
	case reflect.Map:
		return decodeStructFromMap(name, dataVal, val)

		//	case reflect.Struct:
		//		// Not the most efficient way to do this but we can optimize later if
		//		// we want to. To convert from struct to struct we go to map first
		//		// as an intermediary.
		//
		//		// Make a new map to hold our result
		//		mapType := reflect.TypeOf((map[string]interface{})(nil))
		//		mval := reflect.MakeMap(mapType)
		//
		//		// Creating a pointer to a map so that other methods can completely
		//		// overwrite the map if need be (looking at you decodeMapFromMap). The
		//		// indirection allows the underlying map to be settable (CanSet() == true)
		//		// where as reflect.MakeMap returns an unsettable map.
		//		addrVal := reflect.New(mval.Type())
		//
		//		reflect.Indirect(addrVal).Set(mval)
		//		if err := decodeMapFromStruct(name, dataVal, reflect.Indirect(addrVal), mval); err != nil {
		//			return err
		//		}
		//
		//		result := decodeStructFromMap(name, reflect.Indirect(addrVal), val)
		//		return result

	default:
		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
	}
}

//func decodeMapFromStruct(name string, dataVal reflect.Value, val reflect.Value, valMap reflect.Value) error {
//
//	fmt.Println("decodeMapFromStruct")
//
//	typ := dataVal.Type()
//	for i := 0; i < typ.NumField(); i++ {
//		// Get the StructField first since this is a cheap operation. If the
//		// field is unexported, then ignore it.
//		f := typ.Field(i)
//		if f.PkgPath != "" {
//			continue
//		}
//
//		// Next get the actual value of this field and verify it is assignable
//		// to the map value.
//		v := dataVal.Field(i)
//		if !v.Type().AssignableTo(valMap.Type().Elem()) {
//			return fmt.Errorf("cannot assign type '%s' to map value field of type '%s'", v.Type(), valMap.Type().Elem())
//		}
//
//		tagValue := f.Tag.Get("c2s")
//		keyName := f.Name
//
//		// Determine the name of the key in the map
//		if index := strings.Index(tagValue, ","); index != -1 {
//			if tagValue[:index] == "-" {
//				continue
//			}
//			// If "omitempty" is specified in the tag, it ignores empty values.
//			if strings.Index(tagValue[index+1:], "omitempty") != -1 && isEmptyValue(v) {
//				continue
//			}
//
//			keyName = tagValue[:index]
//		} else if len(tagValue) > 0 {
//			if tagValue == "-" {
//				continue
//			}
//			keyName = tagValue
//		}
//
//		switch v.Kind() {
//		// this is an embedded struct, so handle it differently
//		case reflect.Struct:
//			x := reflect.New(v.Type())
//			x.Elem().Set(v)
//
//			vType := valMap.Type()
//			vKeyType := vType.Key()
//			vElemType := vType.Elem()
//			mType := reflect.MapOf(vKeyType, vElemType)
//			vMap := reflect.MakeMap(mType)
//
//			// Creating a pointer to a map so that other methods can completely
//			// overwrite the map if need be (looking at you decodeMapFromMap). The
//			// indirection allows the underlying map to be settable (CanSet() == true)
//			// where as reflect.MakeMap returns an unsettable map.
//			addrVal := reflect.New(vMap.Type())
//			reflect.Indirect(addrVal).Set(vMap)
//
//			err := decode(keyName, x.Interface(), reflect.Indirect(addrVal))
//			if err != nil {
//				return err
//			}
//
//			// the underlying map may have been completely overwritten so pull
//			// it indirectly out of the enclosing value.
//			vMap = reflect.Indirect(addrVal)
//
//			valMap.SetMapIndex(reflect.ValueOf(keyName), vMap)
//
//		default:
//			valMap.SetMapIndex(reflect.ValueOf(keyName), v)
//		}
//	}
//
//	if val.CanAddr() {
//		val.Set(valMap)
//	}
//
//	return nil
//}

func decodeStructFromMap(name string, dataVal, val reflect.Value) error {

	fmt.Println("decodeStructFromMap")

	dataValType := dataVal.Type()
	if kind := dataValType.Key().Kind(); kind != reflect.String && kind != reflect.Interface {
		return fmt.Errorf(
			"'%s' needs a map with string keys, has '%s' keys",
			name, dataValType.Key().Kind())
	}

	dataValKeys := make(map[reflect.Value]struct{})
	dataValKeysUnused := make(map[interface{}]struct{})
	for _, dataValKey := range dataVal.MapKeys() {
		fmt.Println("ranged key:", dataValKey)
		dataValKeys[dataValKey] = struct{}{}
		dataValKeysUnused[dataValKey.Interface()] = struct{}{}
	}

	errors := make([]string, 0)

	type field struct {
		field reflect.StructField
		val   reflect.Value
	}

	fields := []field{}
	structVal := val
	structType := structVal.Type()

	// Fill slice with all struct field values
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldVal := structVal.Field(i)
		fields = append(fields, field{fieldType, fieldVal})
	}

	// Fill each field with respective map value
	for _, f := range fields {
		field, fieldValue := f.field, f.val
		fieldName := field.Name

		tags := field.Tag.Get("c2s")
		fmt.Println("tags:", tags)
		tagSlice := strings.Split(tags, ",")
		tagValue := tagSlice[0]
		tagDynamic := ""
		if len(tagSlice) > 1 {
			tagDynamic = tagSlice[1]
		}

		// use tag if present
		if tagValue != "" {
			fieldName = tagValue
		} else {
			return fmt.Errorf("tag value not set")
		}

		// cast to type if tagDynamic is present
		if tagDynamic != "" {
			selectSlice := strings.Split(tagDynamic, "=")
			selectValue := selectSlice[1]
			rawMapSelectKey := reflect.ValueOf(selectValue)
			rawMapKey := reflect.ValueOf(fieldName)
			rawMapVal := dataVal.MapIndex(rawMapKey)
			rawMapSelectVal := rawMapVal.Elem().MapIndex(rawMapSelectKey)
			fmt.Printf("set select: %s \n", rawMapSelectVal.Interface().(string))
			fieldValue.Addr().Interface().(DynamicStruct).SetDynamicType(rawMapSelectVal.Interface().(string))
			fmt.Printf("set field: %#v \n", fieldValue)
		}

		rawMapKey := reflect.ValueOf(fieldName)
		rawMapVal := dataVal.MapIndex(rawMapKey)

		if !rawMapVal.IsValid() {
			panic("raw map is not valid")
		}

		if !fieldValue.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !fieldValue.CanSet() {
			return fmt.Errorf("can not set field")
		}

		// Delete the key we're using from the unused map so we stop tracking
		delete(dataValKeysUnused, rawMapKey.Interface())

		// If the name is empty string, then we're at the root, and we
		// don't dot-join the fields.
		if name != "" {
			fieldName = name + "." + fieldName
		}

		if err := decode(fieldName, rawMapVal.Interface(), fieldValue); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	//	// If we have a "remain"-tagged field and we have unused keys then
	//	// we put the unused keys directly into the remain field.
	//	if remainField != nil && len(dataValKeysUnused) > 0 {
	//		// Build a map of only the unused values
	//		remain := map[interface{}]interface{}{}
	//		for key := range dataValKeysUnused {
	//			remain[key] = dataVal.MapIndex(reflect.ValueOf(key)).Interface()
	//		}
	//
	//		// Decode it as-if we were just decoding this map onto our map.
	//		if err := decodeMap(name, remain, remainField.val); err != nil {
	//			errors = appendErrors(errors, err)
	//		}
	//
	//		// Set the map to nil so we have none so that the next check will
	//		// not error (ErrorUnused)
	//		dataValKeysUnused = nil
	//	}
	//
	//	if len(dataValKeysUnused) > 0 {
	//		keys := make([]string, 0, len(dataValKeysUnused))
	//		for rawKey := range dataValKeysUnused {
	//			keys = append(keys, rawKey.(string))
	//		}
	//		sort.Strings(keys)
	//
	//		err := fmt.Errorf("'%s' has invalid keys: %s", name, strings.Join(keys, ", "))
	//		errors = appendErrors(errors, err)
	//	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

//func decodeMapFromMap(name string, dataVal reflect.Value, val reflect.Value, valMap reflect.Value) error {
//	valType := val.Type()
//	valKeyType := valType.Key()
//	valElemType := valType.Elem()
//
//	// Accumulate errors
//	errors := make([]string, 0)
//
//	// If the input data is empty, then we just match what the input data is.
//	if dataVal.Len() == 0 {
//		if dataVal.IsNil() {
//			if !val.IsNil() {
//				val.Set(dataVal)
//			}
//		} else {
//			// Set to empty allocated value
//			val.Set(valMap)
//		}
//
//		return nil
//	}
//
//	for _, k := range dataVal.MapKeys() {
//		fieldName := name + "[" + k.String() + "]"
//
//		// First decode the key into the proper type
//		currentKey := reflect.Indirect(reflect.New(valKeyType))
//		if err := decode(fieldName, k.Interface(), currentKey); err != nil {
//			errors = appendErrors(errors, err)
//			continue
//		}
//
//		// Next decode the data into the proper type
//		v := dataVal.MapIndex(k).Interface()
//		currentVal := reflect.Indirect(reflect.New(valElemType))
//		if err := decode(fieldName, v, currentVal); err != nil {
//			errors = appendErrors(errors, err)
//			continue
//		}
//
//		valMap.SetMapIndex(currentKey, currentVal)
//	}
//
//	// Set the built up map to the value
//	val.Set(valMap)
//
//	// If we had errors, return those
//	if len(errors) > 0 {
//		return &Error{errors}
//	}
//
//	return nil
//}
//
//func decodeMap(name string, data interface{}, val reflect.Value) error {
//	valType := val.Type()
//	valKeyType := valType.Key()
//	valElemType := valType.Elem()
//
//	// By default we overwrite keys in the current map
//	valMap := val
//
//	// If the map is nil or we're purposely zeroing fields, make a new map
//	if valMap.IsNil() {
//		// Make a new map to hold our result
//		mapType := reflect.MapOf(valKeyType, valElemType)
//		valMap = reflect.MakeMap(mapType)
//	}
//
//	// Check input type and based on the input type jump to the proper func
//	dataVal := reflect.Indirect(reflect.ValueOf(data))
//	switch dataVal.Kind() {
//	case reflect.Map:
//		return decodeMapFromMap(name, dataVal, val, valMap)
//
//	case reflect.Struct:
//		return decodeMapFromStruct(name, dataVal, val, valMap)
//
//	case reflect.Array, reflect.Slice:
//		fallthrough
//
//	default:
//		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
//	}
//}

func isEmptyValue(v reflect.Value) bool {
	switch getKind(v) {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}
