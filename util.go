package ucfg

import (
	"fmt"
	"reflect"
	"strings"
)

type tagOptions struct {
	squash bool
}

func parseTags(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	opts := tagOptions{}
	for _, opt := range s[1:] {
		if opt == "squash" {
			opts.squash = true
		}
	}
	return s[0], opts
}

func fieldName(tagName, structName string) string {
	if tagName != "" {
		return tagName
	}
	return strings.ToLower(structName)
}

func chaseValueInterfaces(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

func chaseValuePointers(v reflect.Value) reflect.Value {
	fmt.Println("start - chaseValuePointers")
	fmt.Println(v, v.Type())
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
		fmt.Println(v, v.Type())
	}
	fmt.Println("done - chaseValuePointers")
	return v
}

func chaseValue(v reflect.Value) reflect.Value {
	fmt.Println("start - chaseValue")
	fmt.Println(v, v.Type())
	for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
		v = v.Elem()
		fmt.Println(v, v.Type())
	}
	fmt.Println("done - chaseValue")
	return v
}

func chaseTypePointers(t reflect.Type) reflect.Type {
	fmt.Println("start - chaseTypePointers")
	fmt.Println(t)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		fmt.Println(t)
	}
	fmt.Println("done - chaseTypePointers")
	return t
}
