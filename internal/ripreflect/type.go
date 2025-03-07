package ripreflect

import "reflect"

func TagFromType(v any) (string, bool) {
	if v == nil {
		return "", false
	}

	t := reflect.TypeOf(v)
	for {
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			t = t.Elem()
		default:
			return t.Name(), true
		}
	}
}
