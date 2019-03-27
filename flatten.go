package structs

import "reflect"

// This is a modified version of the code at https://github.com/doublerebel/bellows that is able to handle slices, array, and actually handle json tags on structs if structs.DefaultTagName is set to json.

// Flatten takes a map or struct and flattens any nested maps inside of it so that the result is a map that is one level deep, except for arrays.
// If there is an array located inside the map structure, each element will have flatten ran on it, such that an array of objects will be transformed into an array of flattened objects.
// Example: input: {"nested":{"i": "am"}, "array": [{"nested":{"am": "i"}}]}
// Example: output: {"nested.i":"am", "array": [{"nested.am": "i"}]}
// Note flatten will infinitely recurse if given cyclic structures. It also only handles maps that have a key of type string.
// These methods do not work with structs with unexported fields (such as time.Time)
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
		anySlice := make([]interface{}, 0, original.Len())
		for i := 0; i < original.Len(); i++ {
			elem := original.Index(i)
			// If this array element is a nil pointer, skip over it
			if !reflect.Indirect(elem).IsValid() {
				continue
			} else {
				elem = reflect.Indirect(elem)
			}
			if !elem.CanInterface() {
				continue
			}
			// Dereference an interface before switching on the type.
			if elem.Kind() == reflect.Interface {
				elem = reflect.ValueOf(elem.Interface())
				if !elem.CanInterface() {
					continue
				}
			}
			switch elem.Kind() {
			case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
				anySlice = append(anySlice, Flatten(elem.Interface()))
			default:
				anySlice = append(anySlice, elem.Interface())
			}
		}
		m[prefix] = anySlice
	default:
		if prefix != "" {
			m[prefix] = value
		}
	}
}
