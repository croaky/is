package is_test

// The label is read from the caller's file, and the caller is normally
// in another module entirely. These tests assert from outside package
// is, which is the closest a single repo gets to that: a different
// package, a different file, and the helper reached through a qualified
// name rather than a bare one.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/croaky/is"
)

// tb records what an assertion would have reported. Fatal cannot Goexit
// here, so a call returns normally after failing.
type tb struct {
	testing.TB
	msgs []string
}

func (f *tb) Helper() {}

func (f *tb) Fatal(args ...any) {
	f.msgs = append(f.msgs, fmt.Sprint(args...))
}

func (f *tb) Error(args ...any) {
	f.msgs = append(f.msgs, fmt.Sprint(args...))
}

func (f *tb) one(t *testing.T) string {
	t.Helper()
	if len(f.msgs) != 1 {
		t.Fatalf("failures = %d, want 1: %v", len(f.msgs), f.msgs)
	}
	return f.msgs[0]
}

func TestLabelReadsAQualifiedCall(t *testing.T) {
	t.Parallel()
	f := &tb{}
	co := struct{ Name string }{Name: "Acme"}
	is.New(f).Eq(co.Name, "Beta")
	if got, want := f.one(t), "co.Name = Acme, want Beta"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestLabelReadsThroughAVariable(t *testing.T) {
	t.Parallel()
	f := &tb{}
	is := is.New(f)
	err := errors.New("boom")
	is.NoErr(err)
	if got, want := f.one(t), "err = boom, want no error"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestLabelReadsAMultilineQualifiedCall(t *testing.T) {
	t.Parallel()
	f := &tb{}
	total := 3
	is.New(f).Eq(
		total,
		5,
	)
	if got, want := f.one(t), "total = 3, want 5"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}
