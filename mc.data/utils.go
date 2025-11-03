package data

import (
	"fmt"
	"reflect"
)

func FilterMultiple[T any](elements []T, predicate func(T) bool) (results []T) {
	for _, element := range elements {
		if predicate(element) {
			results = append(results, element)
		}
	}
	return
}

func FilterFirst[T any](elements []T, predicate func(T) bool) (result T) {
	for _, element := range elements {
		if predicate(element) {
			return element
		}
	}
	return
}

func FilterSingle[T any](elements []T, predicate func(T) bool) (result T, err error) {
	res := FilterMultiple(elements, predicate)

	if len(res) != 1 {
        var zero T
        return zero, fmt.Errorf("error getting single, found %d matches", len(res))
	}

    return res[0], nil
}

func GetFields[T any](value T) (results []string, err error) {
	typ := reflect.TypeOf(value)
	if typ == nil {
		return nil, fmt.Errorf("GetFields: nil type")
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("GetFields: expected struct, got %s", typ.Kind())
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i).Name
		results = append(results, field)
	}
	return results, nil
}