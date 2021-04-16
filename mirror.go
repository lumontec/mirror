package mirror

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type DynamicStruct interface {
	SetDynamicType(Type string)
}

// Unmarshal full yaml into the configuration structure
func UnmarshalYaml(data []byte, config interface{}) error {

	rawmap := make(map[string]interface{})

	err := yaml.Unmarshal([]byte(data), rawmap)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %s", err)
	}

	err = decodeMapLevels(rawmap, config)
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

	err = decodeMapLevels(rawmap, config)
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
		fallthrough
	case reflect.Float64:
		err = decodeFloat(name, input, outVal)
	case reflect.Struct:
		err = decodeStruct(name, input, outVal)
	case reflect.Ptr:
		_, err = decodePtr(name, input, outVal)
	case reflect.Slice:
		err = decodeSlice(name, input, outVal)
	case reflect.Array:
		err = decodeArray(name, input, outVal)
	default:
		// If we reached this point then we weren't able to decode it
		return fmt.Errorf("%s: unsupported type: %s", name, outputKind)
	}

	return err
}

func decodeBool(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	if dataKind != reflect.Bool {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	val.SetBool(dataVal.Bool())
	return nil
}

func decodeInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	if dataKind != reflect.Int {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	val.SetInt(dataVal.Int())
	return nil
}

func decodeUint(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	if dataKind != reflect.Uint {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	val.SetUint(dataVal.Uint())
	return nil
}

func decodeFloat(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	if dataKind != reflect.Float64 {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	val.SetFloat(dataVal.Float())
	return nil
}

func decodeString(name string, data interface{}, val reflect.Value) error {

	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataKind := getKind(dataVal)

	if dataKind != reflect.String {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
			name, val.Type(), dataVal.Type(), data)
	}

	val.SetString(dataVal.String())
	return nil
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

	if val.IsValid() && val.Elem().IsValid() {
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

func decodeStruct(name string, data interface{}, val reflect.Value) error {

	dataVal := reflect.Indirect(reflect.ValueOf(data))

	// If the type of the value to write to and the data match directly,
	// then we just set it directly instead of recursing into the structure.
	if dataVal.Type() == val.Type() {
		val.Set(dataVal)
		return nil
	}

	dataValKind := dataVal.Kind()

	if dataValKind != reflect.Map {
		return fmt.Errorf(
			"'%s' expected a map, got unconvertible type '%s', value: '%v'",
			name, dataVal.Type(), data)
	}

	return decodeStructFromMap(name, dataVal, val)
}

func decodeStructFromMap(name string, dataVal, val reflect.Value) error {

	dataValType := dataVal.Type()
	if kind := dataValType.Key().Kind(); kind != reflect.String && kind != reflect.Interface {
		return fmt.Errorf(
			"'%s' needs a map with string keys, has '%s' keys",
			name, dataValType.Key().Kind())
	}

	dataValKeys := make(map[reflect.Value]struct{})
	dataValKeysUnused := map[string]interface{}{}

	for _, dataValKey := range dataVal.MapKeys() {
		dataValKeys[dataValKey] = struct{}{}
		dataValKeysUnused[dataValKey.String()] = struct{}{}
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

		// look for tags
		tags := field.Tag.Get("mirror")
		tagSlice := strings.Split(tags, ",")
		tagValue := tagSlice[0]

		tagDynamic := ""
		if len(tagSlice) > 1 {
			tagDynamic = tagSlice[1]
		}

		if tagValue == "" {
			errors = append(errors, "missing `mirror` tag for struct field: "+fieldName)
		}

		// cast to type if tagDynamic is present
		if tagDynamic != "" {

			if !strings.HasPrefix(tagDynamic, "dynamic=") {
				return fmt.Errorf("'%s' invalid dynamic selector tag", name)
			}

			selectValue := strings.Split(tagDynamic, "=")[1]
			rawMapSelectKey := reflect.ValueOf(selectValue)
			rawMapKey := reflect.ValueOf(tagValue)
			rawMapVal := dataVal.MapIndex(rawMapKey)

			if rawMapVal.Elem().Kind() == reflect.Slice {
				// Get slice type
				valType := fieldValue.Type()
				valElemType := valType.Elem()
				sliceType := reflect.SliceOf(valElemType)
				// Create a new slice over the prev
				valSlice := fieldValue
				if valSlice.IsNil() {
					// Make a new slice to hold our result, same size as the original data.
					valSlice = reflect.MakeSlice(sliceType, rawMapVal.Elem().Len(), rawMapVal.Elem().Len())
				}

				// Cast dynamic type for each element of slice
				for i := 0; i < rawMapVal.Elem().Len(); i++ {
					rawMapSelectVal := rawMapVal.Elem().Index(i).Elem().MapIndex(rawMapSelectKey)

					if !rawMapSelectVal.IsValid() {
						errors = append(errors, "map value not found in slice element for dynamic selector: "+selectValue)
						continue
					}

					valSlice.Index(i).Addr().Interface().(DynamicStruct).SetDynamicType(rawMapSelectVal.Interface().(string))
				}

				// Finally, set the value to the slice we built up
				fieldValue.Set(valSlice)

			} else {
				rawMapSelectVal := rawMapVal.Elem().MapIndex(rawMapSelectKey)

				if !rawMapSelectVal.IsValid() {
					errors = append(errors, "map value not found for dynamic selector: "+selectValue)
					continue
				}

				fieldValue.Addr().Interface().(DynamicStruct).SetDynamicType(rawMapSelectVal.Interface().(string))
			}

		}

		rawMapKey := reflect.ValueOf(tagValue)
		rawMapVal := dataVal.MapIndex(rawMapKey)

		if !rawMapVal.IsValid() {
			errors = append(errors, "map value not found for key: "+tagValue)
			continue
		}

		if !fieldValue.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !fieldValue.CanSet() {
			errors = append(errors, "cannot set field: "+fieldName+" likely unexported")
			continue
		}

		// Delete the key we're using from the unused map so we stop tracking
		delete(dataValKeysUnused, rawMapKey.String())

		// If the name is empty string, then we're at the root, and we
		// don't dot-join the fields.
		if name != "" {
			fieldName = name + "." + fieldName
		}

		if err := decode(fieldName, rawMapVal.Interface(), fieldValue); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	// Emit error if unused keys slice
	dataValKeysUnusedString := []string{}
	for key := range dataValKeysUnused {
		dataValKeysUnusedString = append(dataValKeysUnusedString, key)
	}
	sort.Strings(dataValKeysUnusedString)
	if len(dataValKeysUnusedString) > 0 {
		errors = append(errors, "detected unused keys: "+strings.Join(dataValKeysUnusedString, " "))
	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float64
	default:
		return kind
	}
}
