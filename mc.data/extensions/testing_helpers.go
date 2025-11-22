package extensions

import "testing"

func AssertAreEqual[T comparable](t *testing.T, name string, expected T, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, actual)
	}
}

func AssertNillability[T comparable](t *testing.T, name string, expected bool, actual *T) {
	t.Helper()
	if (actual == nil) != expected {
		t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, (actual == nil))
	}
}
