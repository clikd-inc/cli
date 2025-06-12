package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// AssignDynamicValues weist dynamisch Werte zu Struct-Feldern zu
func AssignDynamicValues(target interface{}, attrs []string, values []string) {
	rv := reflect.ValueOf(target).Elem()
	rt := rv.Type()

	for i, field := range attrs {
		if i < len(values) { // Sicherheitsprüfung
			if f, ok := rt.FieldByName(field); ok {
				rv.FieldByIndex(f.Index).SetString(values[i])
			}
		}
	}
}

// ConvNewline konvertiert verschiedene Zeilenumbruchformate in das angegebene Format
func ConvNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

// DotGet ist eine Hilfsfunktion zum Zugriff auf Struct-Felder über einen Punkt-Pfad
// z.B. DotGet(commit, "Author.Name") würde commit.Author.Name zurückgeben
func DotGet(object interface{}, fieldPath string) (interface{}, bool) {
	if object == nil {
		return nil, false
	}

	fields := strings.Split(fieldPath, ".")
	value := reflect.ValueOf(object)

	for _, field := range fields {
		for {
			if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
				if value.IsNil() {
					return nil, false
				}
				value = value.Elem()
			} else {
				break
			}
		}

		if value.Kind() != reflect.Struct {
			return nil, false
		}

		value = value.FieldByName(field)
		if !value.IsValid() {
			return nil, false
		}
	}

	return value.Interface(), true
}

// JoinAndQuoteMeta verbindet Strings und escaped Metazeichen für reguläre Ausdrücke
func JoinAndQuoteMeta(list []string, sep string) string {
	arr := make([]string, len(list))
	for i, s := range list {
		arr[i] = regexp.QuoteMeta(s)
	}
	return strings.Join(arr, sep)
}

// Compare vergleicht zwei Werte mit einem Operator
func Compare(a interface{}, operator string, b interface{}) (bool, error) {
	at := reflect.TypeOf(a).String()
	bt := reflect.TypeOf(b).String()
	if at != bt {
		return false, fmt.Errorf("\"%s\" and \"%s\" can not be compared", at, bt)
	}

	switch at {
	case "string":
		aa := a.(string)
		bb := b.(string)
		return CompareString(aa, operator, bb), nil
	case "int":
		aa := a.(int)
		bb := b.(int)
		return CompareInt(aa, operator, bb), nil
	case "time.Time":
		aa := a.(time.Time)
		bb := b.(time.Time)
		return CompareTime(aa, operator, bb), nil
	}

	return false, nil
}

// CompareString vergleicht zwei Strings mit einem Operator
func CompareString(a string, operator string, b string) bool {
	switch operator {
	case "<":
		return a < b
	case ">":
		return a > b
	case "==":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

// CompareInt vergleicht zwei Integers mit einem Operator
func CompareInt(a int, operator string, b int) bool {
	switch operator {
	case "<":
		return a < b
	case ">":
		return a > b
	case "==":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

// CompareTime vergleicht zwei Zeitwerte mit einem Operator
func CompareTime(a time.Time, operator string, b time.Time) bool {
	switch operator {
	case "<":
		return !a.After(b)
	case ">":
		return a.After(b)
	case "==":
		return a.Equal(b)
	case "!=":
		return !a.Equal(b)
	default:
		return false
	}
}
