package ripreflect

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// FindEntityID will find a struct field named ID or a public struct field
// with the `rip:id` struct tag.
// If there are many `rip:id` struct tags in the struct, it will return the first one.
func FindEntityID(entity any) (value reflect.Value, fieldName string, err error) {
	v := reflect.ValueOf(entity)
	kind := v.Kind()
	if kind == reflect.Pointer {
		v = v.Elem()
	}

	idVal := v.FieldByName("ID")
	fieldFound := "ID"

	var zero reflect.Value
	if idVal != zero {
		return idVal, fieldFound, nil
	}

	t := reflect.TypeOf(entity)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if HasRIPIDField(f) {
			// we found a `rip:id`, we'll set this one
			idVal = v.Field(i)
			fieldFound = f.Name
			return idVal, fieldFound, nil
		}

	}

	return idVal, fieldFound, errors.New("no ID field found")

}

func FieldIDString(entity any) string {
	v, _, err := FindEntityID(entity)
	if err != nil {
		return "<NO ID FIELD FOUND>"
	}

	return fmt.Sprintf("%v", v)
}

func FieldIDName(entity any) string {
	_, fieldName, err := FindEntityID(entity)
	if err != nil {
		return "<NO ID FIELD FOUND>"
	}

	return fieldName
}

func HasRIPIDField(f reflect.StructField) bool {
	ripTag, ok := f.Tag.Lookup("rip")
	if !ok {
		return false
	}

	tagValues := strings.Split(ripTag, ",")
	if slices.Contains(tagValues, "id") {
		return true
	}
	return false
}