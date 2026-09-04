package is

import (
	"fmt"
	"reflect"
	"testing"
)

// A asserts against a testing.TB.
type A struct {
	tb      testing.TB
	relaxed bool
}

// New returns an A that fails the test fatally on the first failure.
func New(tb testing.TB) *A {
	tb.Helper()
	return &A{tb: tb}
}

// NewRelaxed returns an A that records failures and keeps going.
func NewRelaxed(tb testing.TB) *A {
	tb.Helper()
	return &A{tb: tb, relaxed: true}
}

// Eq fails when got and want are unequal.
func (a *A) Eq(got, want any) {
	a.tb.Helper()
	if equal(got, want) {
		return
	}
	g, w := describe(got, want)
	a.fail(label(2, "Eq", 0), g, "want "+w)
}

// NotEq fails when got and want are equal. Both values are equal by
// then, so only one prints.
func (a *A) NotEq(got, want any) {
	a.tb.Helper()
	if !equal(got, want) {
		return
	}
	a.fail(label(2, "NotEq", 0), fmt.Sprintf("%v", got), "want anything else")
}

// NoErr fails when err is non-nil.
func (a *A) NoErr(err error) {
	a.tb.Helper()
	if err == nil {
		return
	}
	a.fail(label(2, "NoErr", 0), fmt.Sprintf("%v", err), "want no error")
}

// HasErr fails when err is nil.
func (a *A) HasErr(err error) {
	a.tb.Helper()
	if err != nil {
		return
	}
	a.fail(label(2, "HasErr", 0), "nil", "want an error")
}

// Nil fails when v is non-nil. Eq(v, nil) agrees, because equal routes
// a nil operand through isNil too. Nil says what the check is.
func (a *A) Nil(v any) {
	a.tb.Helper()
	if isNil(v) {
		return
	}
	a.fail(label(2, "Nil", 0), fmt.Sprintf("%v", v), "want nil")
}

// NotNil fails when v is nil.
func (a *A) NotNil(v any) {
	a.tb.Helper()
	if !isNil(v) {
		return
	}
	a.fail(label(2, "NotNil", 0), "nil", "want non-nil")
}

// True fails when cond is false. Use it only where there is no want to
// name; prefer Eq for a comparison.
func (a *A) True(cond bool) {
	a.tb.Helper()
	if cond {
		return
	}
	src := label(2, "True", 0)
	if src == "" {
		a.report("assertion failed")
		return
	}
	a.report(src + " = false")
}

// fail reports "<label> = <got>, <want>", degrading to "got <got>,
// <want>" when the label could not be read. The values always print;
// only the label depends on the source.
func (a *A) fail(label, got, want string) {
	a.tb.Helper()
	if label == "" {
		a.report("got " + got + ", " + want)
		return
	}
	a.report(label + " = " + got + ", " + want)
}

func (a *A) report(msg string) {
	a.tb.Helper()
	if a.relaxed {
		a.tb.Error(msg)
		return
	}
	a.tb.Fatal(msg)
}

// equal keeps =='s pointer identity where both values are comparable and
// falls back to reflect.DeepEqual where == is undefined, so slices and
// maps compare by contents. A nil operand takes neither path: both sides
// go through isNil, so a typed nil boxed into any compares equal to nil
// rather than to the non-nil interface holding it.
func equal(got, want any) bool {
	if got == nil || want == nil {
		return isNil(got) && isNil(want)
	}
	gv, wv := reflect.ValueOf(got), reflect.ValueOf(want)
	if gv.Comparable() && wv.Comparable() {
		return gv.Equal(wv)
	}
	return reflect.DeepEqual(got, want)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// describe formats both values, adding types when the values print the
// same. Eq takes any, so Eq(int, int64) builds and then fails with two
// values that read identically.
func describe(got, want any) (string, string) {
	g, w := fmt.Sprintf("%v", got), fmt.Sprintf("%v", want)
	if g != w {
		return g, w
	}
	return fmt.Sprintf("%v (%T)", got, got), fmt.Sprintf("%v (%T)", want, want)
}
