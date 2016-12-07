package docopt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func errKey(key string) error {
	return fmt.Errorf("no such key: %q", key)
}
func errType(key string) error {
	return fmt.Errorf("key: %q failed type conversion", key)
}
func errStrconv(key string, convErr error) error {
	return fmt.Errorf("key: %q failed type conversion: %s", key, convErr)
}

// Opts is a map of command line options to their values, with some convenience
// methods for value type conversion (bool, float64, int, string). For example,
// to get an option value as an int:
//
// 	opts, _ := docopt.ParseDoc("Usage: sleep <seconds>")
// 	secs, _ := opts.Int("<seconds>")
//
// Additionally, Opts.Bind allows you easily populate a struct's fields with the
// values of each option value. See below for examples.
//
// Lastly, you can still treat Opts as a regular map, and do any type checking
// and conversion that you want to yourself. For example:
//
// 	if s, ok := opts["<binary>"].(string); ok {
// 		if val, err := strconv.ParseUint(s, 2, 64); err != nil { ... }
// 	}
//
// Note that any non-boolean option / flag will have a string value in the
// underlying map.
type Opts map[string]interface{}

func (o Opts) String(key string) (s string, err error) {
	v, ok := o[key]
	if !ok {
		err = errKey(key)
		return
	}
	s, ok = v.(string)
	if !ok {
		err = errType(key)
	}
	return
}

func (o Opts) Bool(key string) (b bool, err error) {
	v, ok := o[key]
	if !ok {
		err = errKey(key)
		return
	}
	b, ok = v.(bool)
	if !ok {
		err = errType(key)
	}
	return
}

func (o Opts) Int(key string) (i int, err error) {
	s, err := o.String(key)
	if err != nil {
		return
	}
	i, err = strconv.Atoi(s)
	if err != nil {
		err = errStrconv(key, err)
	}
	return
}

func (o Opts) Float64(key string) (f float64, err error) {
	s, err := o.String(key)
	if err != nil {
		return
	}
	f, err = strconv.ParseFloat(s, 64)
	if err != nil {
		err = errStrconv(key, err)
	}
	return
}

// Bind populates the fields of a given struct with matching option values.
// Each key in Opts will be mapped to an exported field of the struct pointed
// to by `v`, as follows:
//
// 	field int                     // `field` is ignored by Bind
// 	Field int `docopt:"--help"`   // `Field` mapped from key "--help"
// 	Field int `docopt:"-h"`       // `Field` mapped from key "-h"
// 	Field int                     // `Field` mapped from key "--field"
// 	F int                         // `F` mapped from key "-f"
//
// Bind will handle conversion to bool, float64, int or string types.
func (o Opts) Bind(v interface{}) error {
	structVal := reflect.ValueOf(v)
	if structVal.Kind() != reflect.Ptr {
		return newError("'v' argument is not pointer to struct type")
	}
	for structVal.Kind() == reflect.Ptr {
		structVal = structVal.Elem()
	}
	if structVal.Kind() != reflect.Struct {
		return newError("'v' argument is not pointer to struct type")
	}
	structType := structVal.Type()

	indexMap := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if isUnexportedField(field) || field.Anonymous {
			continue
		}
		tag := field.Tag.Get("docopt")
		if tag == "" {
			key := strings.ToLower(field.Name)
			if len(field.Name) == 1 {
				key = "-" + key
			} else {
				key = "--" + key
			}
			indexMap[key] = i
			continue
		}
		for _, t := range strings.Split(tag, ",") {
			indexMap[t] = i
		}
	}

	for k, v := range o {
		i, ok := indexMap[k]
		if !ok {
			if k == "--help" || k == "--version" {
				continue
			}
			return newError("mapping of %q is not found in given struct, or is an unexported field", k)
		}
		field := structVal.Field(i)
		// If the struct's field is already non-zero, then don't change it.
		if !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			continue
		}
		optVal := reflect.ValueOf(v)
		// Option value is the zero Value, so we can't get its .Type(). No need to assign anyway, so move along.
		if !optVal.IsValid() {
			continue
		}
		if !field.CanSet() {
			return newError("%q field cannot be set", structType.Field(i).Name)
		}
		// Try to assign now if able. bool and string values should be assignable already.
		if optVal.Type().AssignableTo(field.Type()) {
			field.Set(optVal)
			continue
		}
		// Try to convert the value and assign if able.
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if x, err := o.Int(k); err == nil {
				field.SetInt(int64(x))
				continue
			}
		case reflect.Float32, reflect.Float64:
			if x, err := o.Float64(k); err == nil {
				field.SetFloat(x)
				continue
			}
		}
		// TODO: Something clever (recursive?) with non-string slices.
		// case reflect.Slice:
		// 	if optVal.Kind() == reflect.Slice {
		// 		for i := 0; i < optVal.Len(); i++ {
		// 			sliceVal := optVal.Index(i)
		// 			fmt.Printf("%v", sliceVal)
		// 		}
		// 		fmt.Printf("\n")
		// 	}
		return newError("value of %q is not assignable to %q field", k, structType.Field(i).Name)
	}

	return nil
}

// isUnexportedField returns whether the field is unexported.
// isUnexportedField is to avoid the bug in versions older than Go1.3.
// See following links:
//   https://code.google.com/p/go/issues/detail?id=7247
//   http://golang.org/ref/spec#Exported_identifiers
func isUnexportedField(field reflect.StructField) bool {
	return !(field.PkgPath == "" && unicode.IsUpper(rune(field.Name[0])))
}
