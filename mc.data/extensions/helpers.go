package extensions

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FilterMultiple return all elements that satisfy the predicate
func FilterMultiple[T any](elements []T, predicate func(T) bool) (results []T) {
	for _, element := range elements {
		if predicate(element) {
			results = append(results, element)
		}
	}
	return
}

// FilterMultiplePtr return all pointers that satisfy the predicate
func FilterMultiplePtr[T any](elements []*T, predicate func(*T) bool) (results []*T) {
	for _, element := range elements {
		if predicate(element) {
			results = append(results, element)
		}
	}
	return
}

// FilterFirst return the first element that satisfies the predicate
func FilterFirst[T any](elements []T, predicate func(T) bool) (result T) {
	for _, element := range elements {
		if predicate(element) {
			return element
		}
	}
	return
}

// FilterFirstPtr return the first pointer that satisfies the predicate
func FilterFirstPtr[T any](elements []*T, predicate func(*T) bool) (result *T) {
	for _, element := range elements {
		if predicate(element) {
			return element
		}
	}
	return
}

// FilterSingle return the single element that satisfies the predicate.
// If zero or more than one, default T and an error is returned.
func FilterSingle[T any](elements []T, predicate func(T) bool) (T, error) {
	res := FilterMultiple(elements, predicate)

	if len(res) != 1 {
		var zero T
		return zero, fmt.Errorf("error getting single, found %d matches", len(res))
	}

	return res[0], nil
}

// FilterSinglePtr return the single pointer that satisfies the predicate.
// If zero or more than one, nil is returned.
func FilterSinglePtr[T any](elements []*T, predicate func(*T) bool) *T {
	res := FilterMultiplePtr(elements, predicate)

	if len(res) != 1 {
		return nil
	}

	return res[0]
}

// GetFields will get the attributes names within a struct using reflection
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
	return
}

// AreEqual is a simple case invariant string comparason
func AreEqual(s, c string) bool {
	return strings.EqualFold(s, c)
}

// AreNotEqual is a simple case invariant string comparason
func AreNotEqual(s, c string) bool {
	return !AreEqual(s, c)
}

// AreAllEqual checks if a slice is complised of the same element by value
func AreAllEqual[T comparable](values []T) bool {
	for i := 1; i < len(values); i++ {
		if values[i] != values[0] {
			return false
		}
	}
	return true
}

// FmtShort formats a time in a date only string
func FmtShort(t time.Time) string {
	return t.Format(time.DateOnly)
}

// FmtLong formats a time to a full date string
func FmtLong(t time.Time) string {
	return t.Format(time.RFC3339)
}

func DotProduct[T Number](a, b []T) (res T, err error) {
	if len(a) != len(b) {
		return res, fmt.Errorf("error in dotproduct, lengths of vectors are not equal")
	}

	for i, v := range a {
		res += v * b[i]
	}

	return
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Sum[T Number](inp []T) (res T) {
	for _, v := range inp {
		res += v
	}
	return
}
