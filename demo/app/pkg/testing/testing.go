package testing

import (
	"reflect"
	"regexp"
	"testing"
)

func Ok(t *testing.T, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func Equals(t *testing.T, got any, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func Match(t *testing.T, got string, want string) {
	t.Helper()
	m, _ := regexp.MatchString(want, got)

	if !m {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func IsNull(t *testing.T, got any) {
	t.Helper()
	if got != nil {
		t.Errorf("expected nil but got %v", got)
	}
}

func IsNotNull(t *testing.T, got any) {
	t.Helper()
	if got == nil {
		t.Errorf("expected not nil but got %v", got)
	}
}

