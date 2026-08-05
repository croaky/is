package is

import (
	"errors"
	"fmt"
	"testing"
)

// fake records what an assertion would have reported. Fatal can't
// Goexit here, so a call returns normally after failing.
type fake struct {
	testing.TB
	msgs []string
}

func (f *fake) Helper() {}

func (f *fake) Fatal(args ...any) {
	f.msgs = append(f.msgs, fmt.Sprint(args...))
}

func (f *fake) Error(args ...any) {
	f.msgs = append(f.msgs, fmt.Sprint(args...))
}

func (f *fake) one(t *testing.T) string {
	t.Helper()
	if len(f.msgs) != 1 {
		t.Fatalf("failures = %d, want 1: %v", len(f.msgs), f.msgs)
	}
	return f.msgs[0]
}

func TestEqLabelsTheGotExpression(t *testing.T) {
	t.Parallel()
	f := &fake{}
	co := struct{ Name string }{Name: "Acme"}
	New(f).Eq(co.Name, "Beta")
	if got, want := f.one(t), "co.Name = Acme, want Beta"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestEqLabelsACallExpression(t *testing.T) {
	t.Parallel()
	f := &fake{}
	New(f).Eq(len([]int{1, 2, 3}), 5)
	if got, want := f.one(t), "len([]int{1, 2, 3}) = 3, want 5"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestEqKeepsTheLabelForAMultilineCall(t *testing.T) {
	t.Parallel()
	f := &fake{}
	co := struct{ Name string }{Name: "Acme"}
	New(f).Eq(
		co.Name,
		"Beta",
	)
	if got, want := f.one(t), "co.Name = Acme, want Beta"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestEqDropsTheLabelForALiteral(t *testing.T) {
	t.Parallel()
	f := &fake{}
	New(f).Eq("Acme", "Beta")
	if got, want := f.one(t), "got Acme, want Beta"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestEqPrintsTypesWhenTheValuesPrintTheSame(t *testing.T) {
	t.Parallel()
	f := &fake{}
	var n int64 = 3
	New(f).Eq(3, n)
	if got, want := f.one(t), "got 3 (int), want 3 (int64)"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestEqComparesSlicesByContents(t *testing.T) {
	t.Parallel()
	f := &fake{}
	New(f).Eq([]int{1, 2}, []int{1, 2})
	if len(f.msgs) != 0 {
		t.Fatalf("failures = %v, want none", f.msgs)
	}
}

func TestEqComparesPointersByIdentity(t *testing.T) {
	t.Parallel()
	f := &fake{}
	New(f).Eq(new(3), new(3))
	if len(f.msgs) != 1 {
		t.Fatalf("failures = %v, want 1", f.msgs)
	}
}

func TestNotEq(t *testing.T) {
	t.Parallel()
	f := &fake{}
	name := "Acme"
	New(f).NotEq(name, "Acme")
	if got, want := f.one(t), "name = Acme, want anything else"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestNoErr(t *testing.T) {
	t.Parallel()
	f := &fake{}
	err := errors.New("boom")
	New(f).NoErr(err)
	if got, want := f.one(t), "err = boom, want no error"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestHasErr(t *testing.T) {
	t.Parallel()
	f := &fake{}
	var err error
	New(f).HasErr(err)
	if got, want := f.one(t), "err = nil, want an error"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestNilSeesThroughATypedNil(t *testing.T) {
	t.Parallel()
	f := &fake{}
	var p *int
	New(f).Nil(p)
	if len(f.msgs) != 0 {
		t.Fatalf("failures = %v, want none", f.msgs)
	}
	New(f).NotNil(p)
	if got, want := f.one(t), "p = nil, want non-nil"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestNil(t *testing.T) {
	t.Parallel()
	f := &fake{}
	v := 3
	New(f).Nil(&v)
	if got := f.one(t); got == "" {
		t.Fatal("want a failure message")
	}
}

func TestTruePrintsTheExpression(t *testing.T) {
	t.Parallel()
	f := &fake{}
	n := 2
	New(f).True(n > 5)
	if got, want := f.one(t), "n > 5 = false"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestTrueFallsBackWithNoExpression(t *testing.T) {
	t.Parallel()
	f := &fake{}
	New(f).True(false)
	if got, want := f.one(t), "assertion failed"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestNewRelaxedKeepsGoing(t *testing.T) {
	t.Parallel()
	f := &fake{}
	is := NewRelaxed(f)
	for _, n := range []int{1, 2} {
		is.Eq(n, 0)
	}
	if len(f.msgs) != 2 {
		t.Fatalf("failures = %v, want 2", f.msgs)
	}
}

func TestLabelRejectsAMismatchedCallee(t *testing.T) {
	t.Parallel()
	// The line holds a call, but not to Eq, so the label is refused
	// rather than read from the wrong node.
	if got := label(1, "Eq", 0); got != "" {
		t.Fatalf("label = %q, want empty", got)
	}
}
