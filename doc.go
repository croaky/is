// Package is holds test assertions and nothing else.
//
// Each assertion names what it checks and prints both values. The label
// is read from the caller's source, so no message is written by hand:
//
//	is := is.New(t)
//	is.Eq(co.Name, "Acme")   // co.Name = Beta, want Acme
//	is.NoErr(err)            // err = sql: no rows, want no error
//	is.True(strings.Contains(s, "x"))
//
// The variable shadows the package deliberately, so a call site reads as
// English.
//
// New fails fatally on the first failure. NewRelaxed records the failure
// and lets the test continue, for a test that reports every mismatch in
// a loop.
//
// # Choosing a helper
//
//   - Eq, NotEq for comparing two values. This is the default; prefer it
//     over True.
//   - NoErr, HasErr for errors.
//   - Nil, NotNil for nil checks. Eq(x, nil) agrees, because equal
//     routes a nil operand through isNil too, but the pair names the
//     check: NotNil prints "want non-nil" where NotEq(x, nil) prints
//     "want anything else".
//   - True only for a predicate with no want to name.
//
// True takes an arbitrary bool because something has to:
// True(strings.Contains(s, x)) has no want. It is the escape hatch, not
// the default, and nothing stops True(got == want) from compiling. Only
// a vet pass would.
//
// # Types
//
// Eq takes any rather than a type parameter, because a method may not
// have type parameters and these are methods. That gives up compile-time
// type checking: Eq(gotInt64, 3) builds and then fails, comparing int64
// against int. When the two values format identically the failure prints
// %T as well, which is what makes such a call findable. Type the literal
// rather than converting the value under test.
//
// Comparison uses reflect.Value.Equal where both values are comparable,
// which keeps =='s pointer identity, and reflect.DeepEqual where == is
// undefined, so slices and maps compare by contents. A nil operand
// takes neither: both sides go through the same nil test the Nil helper
// uses, so an interface holding a typed nil compares equal to nil.
//
// # Reading the label
//
// The label is the source text of the assertion's first argument, so the
// values always print and only the label depends on reading the file. It
// degrades to "got X, want Y", and is suppressed for a literal argument,
// where "[]int{1, 2} = []int{1, 2}" is noise.
//
// The lookup survives -trimpath, which is worth stating because that is
// where the obvious implementation fails. Under -trimpath runtime.Caller
// returns a module-relative path that will not open, so resolve
// filepath.Base, which does open: go test runs each binary in its own
// package directory. It then parses the whole file and takes the call
// node covering the line rather than reading the line's own text, so a
// call spread over several lines keeps its label, and requires that node
// to be a call to the helper itself, so a same-named file from another
// package is rejected rather than misread.
//
// gotest.tools/v3/assert has the same two good ideas and repairs a
// relative path only when it detects a Bazel test, attributing relative
// paths to Bazel rather than to -trimpath, so under go test -trimpath
// its parse fails too. matryer/is prints the trailing comment instead of
// the argument, which relocates a hand-written message rather than
// removing it.
package is
