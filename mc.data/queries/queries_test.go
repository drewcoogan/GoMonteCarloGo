package queries

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestQueryHelperAllStringsRecursive(t *testing.T) {
	// collect all query paths in QueryHelper
	var paths []string
	collectQueryPaths(reflect.ValueOf(QueryHelper), &paths)

	if len(paths) == 0 {
		t.Fatal("no query paths in QueryHelper found")
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			if content := Get(path); content == "" {
				t.Errorf("query file %q is empty", path)
			}
		})
	}

	// raw count of .sql files in queries folder
	root := "../queries"
	count := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
			count++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Error walking the path: %v", err)
	}

	// this verified that all of the sql files that are present in the project are also present in the QueryHelper, I am going to force a 1:1 relationship
	if count != len(paths) {
		t.Fatalf("number of .sql files in %s does not match number of query paths in QueryHelper (%d != %d)", root, count, len(paths))
	}
}

// collectQueryPaths recursively walks v (a struct or pointer to struct) and
// appends every string field value to paths. Used to find all query file paths
// in QueryHelper and its substructure.
func collectQueryPaths(v reflect.Value, paths *[]string) {
	// .NumFields() will bomb if it doesnt get a struct, but we know it will be a struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if field.Kind() == reflect.String {
			if s := field.String(); s != "" {
				*paths = append(*paths, s)
			}
		} else {
			collectQueryPaths(field, paths)
		}
	}
}
