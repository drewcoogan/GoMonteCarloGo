package queries

import (
	"reflect"
	"testing"
)


func TestQueryHelperAllStringsRecursive(t *testing.T) {
	var paths []string
	collectQueryPaths(reflect.ValueOf(QueryHelper), &paths)

	if len(paths) == 0 {
		t.Fatal("collectQueryPaths found no query paths in QueryHelper")
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			content := Get(path)
			if content == "" {
				t.Errorf("query file %q is empty or missing", path)
			}
		})
	}
}

// collectQueryPaths recursively walks v (a struct or pointer to struct) and
// appends every string field value to paths. Used to find all query file paths
// in QueryHelper and its substructure.
func collectQueryPaths(v reflect.Value, paths *[]string) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			if s := field.String(); s != "" {
				*paths = append(*paths, s)
			}
		case reflect.Struct:
			collectQueryPaths(field, paths)
		case reflect.Pointer:
			if field.Elem().Kind() == reflect.Struct {
				collectQueryPaths(field, paths)
			}
		}
	}
}