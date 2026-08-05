# is

Test assertions for Go that name what they check and print both values,
labelled from the caller's source. Nothing outside the standard library.

```go
func TestRename(t *testing.T) {
	is := is.New(t)

	co, err := Rename(id, "Acme")

	is.NoErr(err)
	is.Eq(co.Name, "Acme")
}
```

A failure reads:

```
rename_test.go:9: co.Name = Beta, want Acme
```

The label is the source text of the first argument, so there is no
message to write and none to drift. Seven assertions, and the name is
the documentation:

- `Eq`, `NotEq` for comparing two values. The default.
- `NoErr`, `HasErr` for errors.
- `Nil`, `NotNil` for nil checks.
- `True` for a predicate with no want to name, like
  `is.True(strings.Contains(s, x))`. The escape hatch, not the default.

`New(t)` fails fatally on the first failure. `NewRelaxed(t)` records and
keeps going, for a test that reports every mismatch in a loop.

Use `Nil` rather than `Eq(x, nil)`: a typed nil boxed into `any` is a
non-nil interface, so `Eq` reports a nil `*T` as unequal to nil.

See `doc.go` for why `Eq` takes `any`, and how the label is read.

## It works under -trimpath

`go test -trimpath` makes `runtime.Caller` return a module-relative path
that will not open, which is where the obvious implementation of this
silently degrades to "assertion failed" in CI and nowhere else. This
resolves `filepath.Base`, which does open, because `go test` runs each
binary in its own package directory.

`gotest.tools/v3/assert` repairs a relative path only when it detects a
Bazel test, so under `go test -trimpath` its parse fails too.

## This repo is a mirror

Development happens on [cibot](https://dancroak.com/cmd/cibot/), a
self-hosted review and CI server, which holds the branches. GitHub
receives `main` and the tags, so `go get` works and a commit hash is
browsable, and pull requests are closed because there is nothing here to
merge into.

`main` arrives with each merge. A tag is deliberate:

```sh
scripts/tag v0.1.0
```

That pushes the tag to the module path and nowhere else. cibot has no
use for one.

## License

MIT
