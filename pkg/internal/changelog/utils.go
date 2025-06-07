package changelog

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func dotGet(target interface{}, prop string) (interface{}, bool) {
	fmt.Printf("DEBUG: dotGet called with prop=%q\n", prop)

	if target == nil {
		fmt.Printf("DEBUG: dotGet: target is nil\n")
		return nil, false
	}

	// Debug: Zeige Typ des Ziels
	fmt.Printf("DEBUG: dotGet: target type=%T\n", target)

	path := strings.Split(prop, ".")
	fmt.Printf("DEBUG: dotGet: split into %d properties: %v\n", len(path), path)

	if len(path) == 0 {
		fmt.Printf("DEBUG: dotGet: path is empty\n")
		return nil, false
	}

	for _, key := range path {
		fmt.Printf("DEBUG: dotGet: handling property %q\n", key)

		var value reflect.Value

		if reflect.TypeOf(target).Kind() == reflect.Ptr {
			fmt.Printf("DEBUG: dotGet: target is a pointer\n")
			value = reflect.ValueOf(target).Elem()
		} else {
			fmt.Printf("DEBUG: dotGet: target is not a pointer\n")
			value = reflect.ValueOf(target)
		}

		//nolint:staticcheck
		field := value.FieldByName(strings.Title(key))
		if !field.IsValid() {
			fmt.Printf("DEBUG: dotGet: field %q is not valid\n", key)
			return nil, false
		}

		target = field.Interface()
		fmt.Printf("DEBUG: dotGet: got value for %q: %v\n", key, target)
	}

	fmt.Printf("DEBUG: dotGet: returning final value=%v\n", target)
	return target, true
}

// TODO: dotSet ...

func assignDynamicValues(target interface{}, attrs []string, values []string) {
	rv := reflect.ValueOf(target).Elem()
	rt := rv.Type()

	for i, field := range attrs {
		if f, ok := rt.FieldByName(field); ok {
			rv.FieldByIndex(f.Index).SetString(values[i])
		}
	}
}

func compare(a interface{}, operator string, b interface{}) (bool, error) {
	at := reflect.TypeOf(a).String()
	bt := reflect.TypeOf(a).String()
	if at != bt {
		return false, fmt.Errorf("\"%s\" and \"%s\" can not be compared", at, bt)
	}

	switch at {
	case "string":
		aa := a.(string)
		bb := b.(string)
		return compareString(aa, operator, bb), nil
	case "int":
		aa := a.(int)
		bb := b.(int)
		return compareInt(aa, operator, bb), nil
	case "time.Time":
		aa := a.(time.Time)
		bb := b.(time.Time)
		return compareTime(aa, operator, bb), nil
	}

	return false, nil
}

func compareString(a string, operator string, b string) bool {
	switch operator {
	case "<":
		return a < b
	case ">":
		return a > b
	default:
		return false
	}
}

func compareInt(a int, operator string, b int) bool {
	switch operator {
	case "<":
		return a < b
	case ">":
		return a > b
	default:
		return false
	}
}

func compareTime(a time.Time, operator string, b time.Time) bool {
	switch operator {
	case "<":
		return !a.After(b)
	case ">":
		return a.After(b)
	default:
		return false
	}
}

func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}
