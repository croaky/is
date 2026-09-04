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

Use `Nil` and `NotNil` for a nil check. `Eq` sees through a typed nil
too, so `Eq(x, nil)` passes, but the name and the failure say less:
`NotEq(x, nil)` prints `want anything else` where `NotNil` prints
`want non-nil`.

See `doc.go` for why `Eq` takes `any`, and how the label is read.

## GitHub repo is a mirror

Development happens on [cibot](https://dancroak.com/cmd/cibot/), a
self-hosted review and CI server, which holds in progress branches.
GitHub receives `main` and the tags so `go get` works.

## License

MIT
