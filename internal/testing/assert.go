// Package testing provides test assertions.
package testing

import (
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Assert provides fluent test assertions.
type Assert struct {
	t testing.TB
}

// NewAssert creates a new assertion helper.
func NewAssert(t testing.TB) *Assert {
	return &Assert{t: t}
}

// Equal asserts two values are equal.
func (a *Assert) Equal(got, want interface{}, msg ...string) {
	a.t.Helper()
	if !reflect.DeepEqual(got, want) {
		a.t.Errorf("Expected equal:\n  got:  %v\n  want: %v%s", got, want, formatMsg(msg))
	}
}

// NotEqual asserts two values are not equal.
func (a *Assert) NotEqual(got, notWant interface{}, msg ...string) {
	a.t.Helper()
	if reflect.DeepEqual(got, notWant) {
		a.t.Errorf("Expected not equal:\n  got:     %v\n  notWant: %v%s", got, notWant, formatMsg(msg))
	}
}

// True asserts a condition is true.
func (a *Assert) True(cond bool, msg ...string) {
	a.t.Helper()
	if !cond {
		a.t.Errorf("Expected true, got false%s", formatMsg(msg))
	}
}

// False asserts a condition is false.
func (a *Assert) False(cond bool, msg ...string) {
	a.t.Helper()
	if cond {
		a.t.Errorf("Expected false, got true%s", formatMsg(msg))
	}
}

// Nil asserts a value is nil.
func (a *Assert) Nil(v interface{}, msg ...string) {
	a.t.Helper()
	if v != nil {
		a.t.Errorf("Expected nil, got %v%s", v, formatMsg(msg))
	}
}

// NotNil asserts a value is not nil.
func (a *Assert) NotNil(v interface{}, msg ...string) {
	a.t.Helper()
	if v == nil {
		a.t.Errorf("Expected non-nil value%s", formatMsg(msg))
	}
}

// Empty asserts a value is empty.
func (a *Assert) Empty(v interface{}, msg ...string) {
	a.t.Helper()
	if !isEmpty(v) {
		a.t.Errorf("Expected empty, got %v%s", v, formatMsg(msg))
	}
}

// NotEmpty asserts a value is not empty.
func (a *Assert) NotEmpty(v interface{}, msg ...string) {
	a.t.Helper()
	if isEmpty(v) {
		a.t.Errorf("Expected non-empty value%s", formatMsg(msg))
	}
}

// Contains asserts a string contains a substring.
func (a *Assert) Contains(s, substr string, msg ...string) {
	a.t.Helper()
	if !strings.Contains(s, substr) {
		a.t.Errorf("Expected string to contain %q:\n  string: %s%s", substr, s, formatMsg(msg))
	}
}

// NotContains asserts a string does not contain a substring.
func (a *Assert) NotContains(s, substr string, msg ...string) {
	a.t.Helper()
	if strings.Contains(s, substr) {
		a.t.Errorf("Expected string to not contain %q:\n  string: %s%s", substr, s, formatMsg(msg))
	}
}

// HasPrefix asserts a string has a prefix.
func (a *Assert) HasPrefix(s, prefix string, msg ...string) {
	a.t.Helper()
	if !strings.HasPrefix(s, prefix) {
		a.t.Errorf("Expected string to have prefix %q:\n  string: %s%s", prefix, s, formatMsg(msg))
	}
}

// HasSuffix asserts a string has a suffix.
func (a *Assert) HasSuffix(s, suffix string, msg ...string) {
	a.t.Helper()
	if !strings.HasSuffix(s, suffix) {
		a.t.Errorf("Expected string to have suffix %q:\n  string: %s%s", suffix, s, formatMsg(msg))
	}
}

// Greater asserts a value is greater than another (integers only).
func (a *Assert) Greater(got, min int, msg ...string) {
	a.t.Helper()
	if !(got > min) {
		a.t.Errorf("Expected %d > %d%s", got, min, formatMsg(msg))
	}
}

// GreaterOrEqual asserts a value is greater than or equal to another (integers only).
func (a *Assert) GreaterOrEqual(got, min int, msg ...string) {
	a.t.Helper()
	if !(got >= min) {
		a.t.Errorf("Expected %d >= %d%s", got, min, formatMsg(msg))
	}
}

// Less asserts a value is less than another (integers only).
func (a *Assert) Less(got, max int, msg ...string) {
	a.t.Helper()
	if !(got < max) {
		a.t.Errorf("Expected %d < %d%s", got, max, formatMsg(msg))
	}
}

// LessOrEqual asserts a value is less than or equal to another (integers only).
func (a *Assert) LessOrEqual(got, max int, msg ...string) {
	a.t.Helper()
	if !(got <= max) {
		a.t.Errorf("Expected %d <= %d%s", got, max, formatMsg(msg))
	}
}

// FloatGreater asserts a float is greater than another.
func (a *Assert) FloatGreater(got, min float64, msg ...string) {
	a.t.Helper()
	if !(got > min) {
		a.t.Errorf("Expected %f > %f%s", got, min, formatMsg(msg))
	}
}

// FloatLess asserts a float is less than another.
func (a *Assert) FloatLess(got, max float64, msg ...string) {
	a.t.Helper()
	if !(got < max) {
		a.t.Errorf("Expected %f < %f%s", got, max, formatMsg(msg))
	}
}

// ApproxEqual asserts two float values are approximately equal.
func (a *Assert) ApproxEqual(got, want, epsilon float64, msg ...string) {
	a.t.Helper()
	if math.Abs(got-want) > epsilon {
		a.t.Errorf("Expected approximately %v, got %v (epsilon=%v)%s", want, got, epsilon, formatMsg(msg))
	}
}

// NoError asserts no error occurred.
func (a *Assert) NoError(err error, msg ...string) {
	a.t.Helper()
	if err != nil {
		a.t.Errorf("Unexpected error: %v%s", err, formatMsg(msg))
	}
}

// Error asserts an error occurred.
func (a *Assert) Error(err error, msg ...string) {
	a.t.Helper()
	if err == nil {
		a.t.Errorf("Expected error, got nil%s", formatMsg(msg))
	}
}

// ErrorContains asserts an error contains a substring.
func (a *Assert) ErrorContains(err error, substr string, msg ...string) {
	a.t.Helper()
	if err == nil {
		a.t.Errorf("Expected error containing %q, got nil%s", substr, formatMsg(msg))
		return
	}
	if !strings.Contains(err.Error(), substr) {
		a.t.Errorf("Expected error to contain %q:\n  error: %v%s", substr, err, formatMsg(msg))
	}
}

// Len asserts a collection has the expected length.
func (a *Assert) Len(v interface{}, expected int, msg ...string) {
	a.t.Helper()
	actual := reflect.ValueOf(v).Len()
	if actual != expected {
		a.t.Errorf("Expected length %d, got %d%s", expected, actual, formatMsg(msg))
	}
}

// In asserts a value is in a collection.
func (a *Assert) In(item interface{}, collection interface{}, msg ...string) {
	a.t.Helper()
	if !contains(collection, item) {
		a.t.Errorf("Expected %v to be in %v%s", item, collection, formatMsg(msg))
	}
}

// NotIn asserts a value is not in a collection.
func (a *Assert) NotIn(item interface{}, collection interface{}, msg ...string) {
	a.t.Helper()
	if contains(collection, item) {
		a.t.Errorf("Expected %v to not be in %v%s", item, collection, formatMsg(msg))
	}
}

// Panic asserts a function panics.
func (a *Assert) Panic(fn func(), msg ...string) {
	a.t.Helper()
	defer func() {
		if r := recover(); r == nil {
			a.t.Errorf("Expected panic, but function did not panic%s", formatMsg(msg))
		}
	}()
	fn()
}

// NotPanic asserts a function does not panic.
func (a *Assert) NotPanic(fn func(), msg ...string) {
	a.t.Helper()
	defer func() {
		if r := recover(); r != nil {
			a.t.Errorf("Unexpected panic: %v%s", r, formatMsg(msg))
		}
	}()
	fn()
}

// Match asserts a string matches a regex pattern.
func (a *Assert) Match(s, pattern string, msg ...string) {
	a.t.Helper()
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		a.t.Errorf("Invalid regex pattern %q: %v%s", pattern, err, formatMsg(msg))
		return
	}
	if !matched {
		a.t.Errorf("Expected string to match %q:\n  string: %s%s", pattern, s, formatMsg(msg))
	}
}

// NotMatch asserts a string does not match a regex pattern.
func (a *Assert) NotMatch(s, pattern string, msg ...string) {
	a.t.Helper()
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		a.t.Errorf("Invalid regex pattern %q: %v%s", pattern, err, formatMsg(msg))
		return
	}
	if matched {
		a.t.Errorf("Expected string to not match %q:\n  string: %s%s", pattern, s, formatMsg(msg))
	}
}

// FileExists asserts a file exists.
func (a *Assert) FileExists(path string, msg ...string) {
	a.t.Helper()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		a.t.Errorf("Expected file to exist: %s%s", path, formatMsg(msg))
	}
}

// FileNotExists asserts a file does not exist.
func (a *Assert) FileNotExists(path string, msg ...string) {
	a.t.Helper()
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		a.t.Errorf("Expected file to not exist: %s%s", path, formatMsg(msg))
	}
}

// Fail fails the test immediately.
func (a *Assert) Fail(msg string) {
	a.t.Helper()
	a.t.Fatal(msg)
}

// Fatal fails the test with formatted message.
func (a *Assert) Fatal(args ...interface{}) {
	a.t.Helper()
	a.t.Fatal(args...)
}

// Fatalf fails the test with formatted message.
func (a *Assert) Fatalf(format string, args ...interface{}) {
	a.t.Helper()
	a.t.Fatalf(format, args...)
}

// Log logs a message.
func (a *Assert) Log(args ...interface{}) {
	a.t.Log(args...)
}

// Logf logs a formatted message.
func (a *Assert) Logf(format string, args ...interface{}) {
	a.t.Logf(format, args...)
}

// Helper functions

func formatMsg(msg []string) string {
	if len(msg) == 0 {
		return ""
	}
	return "\n  msg: " + strings.Join(msg, " ")
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

func contains(collection, item interface{}) bool {
	cv := reflect.ValueOf(collection)
	iv := reflect.ValueOf(item)

	switch cv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < cv.Len(); i++ {
			if reflect.DeepEqual(cv.Index(i).Interface(), iv.Interface()) {
				return true
			}
		}
	case reflect.Map:
		return cv.MapIndex(iv).IsValid()
	case reflect.String:
		s, ok := item.(string)
		if !ok {
			return false
		}
		return strings.Contains(cv.String(), s)
	}
	return false
}
