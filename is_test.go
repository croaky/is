package is

import (
	"errors"
	"fmt"
	"testing"
	"unsafe"
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

// An interface field that holds a slice makes == panic, so Eq compares
// the struct by contents.
func TestEqComparesAStructWithASliceInAnInterfaceField(t *testing.T) {
	t.Parallel()
	f := &fake{}
	type box struct {
		N int
		V any
	}
	is := NewRelaxed(f)
	is.Eq(box{1, []int{1, 2}}, box{1, []int{1, 2}})
	if len(f.msgs) != 0 {
		t.Fatalf("failures = %v, want none", f.msgs)
	}
	is.Eq(box{1, []int{1, 2}}, box{1, []int{1, 3}})
	is.Eq(box{1, 2}, box{1, 2})
	if len(f.msgs) != 1 {
		t.Fatalf("failures = %v, want 1", f.msgs)
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

// Eq(x, nil) agrees with Nil(x). The any parameter boxes a nil *T into
// a non-nil interface, but that boxing is Eq's doing, not the caller's:
// in the caller's code p == nil holds, so Eq says the same. NoErr takes
// error and stays strict, which is where a typed nil is a real bug.
func TestEqSeesThroughATypedNil(t *testing.T) {
	t.Parallel()
	f := &fake{}
	is := NewRelaxed(f)
	var p *int
	var s []int
	var m map[string]int
	var e *typedErr
	var err error = e
	is.Eq(p, nil)
	is.Eq(s, nil)
	is.Eq(m, nil)
	is.Eq(err, nil)
	is.Eq(nil, p)
	if len(f.msgs) != 0 {
		t.Fatalf("failures = %v, want none", f.msgs)
	}

	v := 3
	is.Eq(&v, nil)
	is.Eq([]int{}, nil)
	is.NotEq(p, nil)
	is.NoErr(err)
	if len(f.msgs) != 4 {
		t.Fatalf("failures = %v, want 4", f.msgs)
	}
}

type typedErr struct{}

func (*typedErr) Error() string { return "typed" }

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

// handle is a named type with an unsafe.Pointer underlying type, as cgo
// and syscall code declare.
type handle unsafe.Pointer

func TestNilSeesANilUnsafePointer(t *testing.T) {
	t.Parallel()
	f := &fake{}
	var p unsafe.Pointer
	New(f).NotNil(p)
	if got, want := f.one(t), "p = nil, want non-nil"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

// Each kind that can be nil gives the same verdict in the Nil pair and
// in the Eq pair against nil.
func TestNilAgreesWithEqForEachNilableKind(t *testing.T) {
	t.Parallel()
	v := 3
	for name, c := range map[string]struct{ nilV, nonNil any }{
		"chan":           {(chan int)(nil), make(chan int)},
		"func":           {(func())(nil), func() {}},
		"map":            {map[string]int(nil), map[string]int{}},
		"pointer":        {(*int)(nil), &v},
		"slice":          {[]int(nil), []int{}},
		"unsafe pointer": {unsafe.Pointer(nil), unsafe.Pointer(&v)},
		"named unsafe":   {handle(nil), handle(unsafe.Pointer(&v))},
	} {
		f := &fake{}
		is := NewRelaxed(f)
		is.Nil(c.nilV)
		is.Eq(c.nilV, nil)
		is.NotNil(c.nonNil)
		is.NotEq(c.nonNil, nil)
		if len(f.msgs) != 0 {
			t.Errorf("%s: failures = %v, want none", name, f.msgs)
		}
		is.NotNil(c.nilV)
		is.NotEq(c.nilV, nil)
		is.Nil(c.nonNil)
		is.Eq(c.nonNil, nil)
		if len(f.msgs) != 4 {
			t.Errorf("%s: failures = %v, want 4", name, f.msgs)
		}
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

// discard is a testing.TB that drops each failure, so a benchmark of a
// failing assertion measures the assertion and not the recording.
type discard struct{ testing.TB }

func (discard) Helper()      {}
func (discard) Fatal(...any) {}
func (discard) Error(...any) {}

// account is a struct of the size a test usually compares.
type account struct {
	ID    int64
	Name  string
	Email string
	Plan  string
	Seats int
}

// The pass path runs once for each assertion in each suite that uses
// this package, so it must stay cheap.
func BenchmarkEqPass(b *testing.B) {
	is := New(discard{})
	co := account{ID: 7, Name: "Acme", Email: "ops@acme.test", Plan: "team", Seats: 12}
	want := co
	ids := []int{1, 2, 3, 4, 5, 6, 7, 8}
	wantIDs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	b.Run("int", func(b *testing.B) {
		b.ReportAllocs()
		n := 3
		for b.Loop() {
			is.Eq(n, 3)
		}
	})
	b.Run("string", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq(co.Name, "Acme")
		}
	})
	b.Run("struct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq(co, want)
		}
	})
	b.Run("slice", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq(ids, wantIDs)
		}
	})
	b.Run("typed nil", func(b *testing.B) {
		b.ReportAllocs()
		var p *int
		for b.Loop() {
			is.Eq(p, nil)
		}
	})
}

func BenchmarkNilPass(b *testing.B) {
	is := New(discard{})
	v := 3
	b.Run("untyped nil", func(b *testing.B) {
		b.ReportAllocs()
		var err error
		for b.Loop() {
			is.Nil(err)
		}
	})
	b.Run("typed nil", func(b *testing.B) {
		b.ReportAllocs()
		var p *int
		for b.Loop() {
			is.Nil(p)
		}
	})
	b.Run("unsafe pointer", func(b *testing.B) {
		b.ReportAllocs()
		var p unsafe.Pointer
		for b.Loop() {
			is.Nil(p)
		}
	})
	b.Run("not nil", func(b *testing.B) {
		b.ReportAllocs()
		p := &v
		for b.Loop() {
			is.NotNil(p)
		}
	})
}

func BenchmarkNoErrPass(b *testing.B) {
	b.ReportAllocs()
	is := New(discard{})
	var err error
	for b.Loop() {
		is.NoErr(err)
	}
}

// The fail path reads the caller's source to find the label. The parse
// of each file is cached, so this measures the search for the call.
func BenchmarkEqFail(b *testing.B) {
	is := NewRelaxed(discard{})
	co := account{Name: "Acme"}
	b.Run("labelled", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq(co.Name, "Beta")
		}
	})
	b.Run("multiline", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq(
				co.Name,
				"Beta",
			)
		}
	})
	b.Run("literal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			is.Eq("Acme", "Beta")
		}
	})
	b.Run("same print", func(b *testing.B) {
		b.ReportAllocs()
		var n int64 = 3
		for b.Loop() {
			is.Eq(3, n)
		}
	})
}

func BenchmarkTrueFail(b *testing.B) {
	b.ReportAllocs()
	is := NewRelaxed(discard{})
	n := 2
	for b.Loop() {
		is.True(n > 5)
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
