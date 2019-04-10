package structs

import (
	"fmt"
	"reflect"
)

// This is a modified version of the code at https://github.com/doublerebel/bellows that is able to handle slices, array, and actually handle json tags on structs if structs.DefaultTagName is set to json.

// Copyright © 2016 Charles Phillips <charles@doublerebel.com>.
// MIT License
// https://github.com/doublerebel/bellows/blob/master/LICENSE
// Note flatten will infinitely recurse if given cyclic structures.
// These methods do not work with structs with unexported fields (such as time.Time)
// Flatten takes a map or struct and flattens any nested maps inside of it so that the result is a map that is one level deep.
// If there is a nested slice of maps/structs, flatten will recursively call itself on them.
func Flatten(value interface{}) map[string]interface{} {
	return FlattenPrefixed(value, "")
}
func FlattenPrefixed(value interface{}, prefix string) map[string]interface{} {
	m := make(map[string]interface{})
	FlattenPrefixedToResult(value, prefix, m)
	return m
}
func FlattenPrefixedToResult(value interface{}, prefix string, m map[string]interface{}) {
	base := ""
	if prefix != "" {
		base = prefix + "."
	}
	if value == nil {
		value = ""
	}
	original := reflect.ValueOf(value)
	kind := original.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		original = reflect.Indirect(original)
		kind = original.Kind()
	}
	//If its a struct, use structs.Map to convert it to a map[string]interface{} and respect json tags as well
	if kind == reflect.Struct {
		original = reflect.ValueOf(Map(value))
		kind = original.Kind()
	}
	if !original.IsValid() {
		if prefix != "" {
			m[prefix] = nil
		}
		return
	}
	t := original.Type()
	switch kind {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			break
		}
		for _, childKey := range original.MapKeys() {
			childValue := original.MapIndex(childKey)
			FlattenPrefixedToResult(childValue.Interface(), base+childKey.String(), m)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < original.Len(); i++ {
			elem := original.Index(i)
			FlattenPrefixedToResult(elem.Interface(), fmt.Sprintf("%s%d", base, i), m)
		}
	default:
		if prefix != "" {
			m[prefix] = value
		}
	}
}
