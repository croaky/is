# Agents guide

is holds Go test assertions that name what they check and print both
values. See `README.md` for why it exists and `doc.go` for the design.

## Architecture

One package at the repo root, standard library only. Two files:

- `is.go` — the seven assertions, `A`, and the two constructors. Each
  assertion decides pass or fail and hands three strings to `fail`.
- `label.go` — the source lookup: caller frame to the text of an
  argument. Knows nothing about assertions beyond a helper's name.

`doc.go` is documentation only.

The split is the point. The assertions are obvious and the lookup is
not, so a change to how a failure reads is one file and a change to how
the source is found is the other.

## Checks

The root `Checkfile` is the list, and CI runs it on every push. Run the
same things before committing, since a check that fails locally fails
there:

```sh
goimports -local "$(go list -m)" -w .
go vet ./...
go test -trimpath -race -cover ./...
git ls-files -z '*.go' | xargs -0 gopls check -severity=hint
```

Keep `-trimpath` on the local run. It is the case this package exists to
handle, and without it a broken lookup passes everywhere except CI.

Nothing outside the standard library is imported. A test-assertion
package that pulls in a dependency puts it in the module graph of every
repo whose tests it serves.

## Tests

Red/green TDD. `is_test.go` asserts on the failure text, because the text
is the product: a test that only checked pass and fail would not have
caught the label. Its `fake` records what would have been reported, since
`Fatal` cannot `Goexit` inside a test of `Fatal`.

- Every exported helper needs a test. There is no `deadcode` job here,
  because a library's API is dead until someone imports it, so the tests
  are what say the surface is real.
- A label test needs the failing form: assert the exact string, not that
  a failure happened. `co.Name = Acme, want Beta` is the behavior.
- Cover the degraded paths too. A literal argument suppresses the label,
  a mismatched callee refuses it, and both are what keep a wrong label
  from being invented.

## Documentation

`doc.go` says why: why `Eq` takes `any`, why `Nil` exists beside `Eq`,
why `True` is the escape hatch, and how the label survives `-trimpath`.
A change to any of those updates it in the same commit. The code says
what.

Plans live in `todo/planned` and are reviewed as ordinary changes. A
numeric prefix means land in this order; unnumbered siblings are
parallel and pickable anytime. Delete a plan's text as it ships rather
than leaving a record of work already done.

## Commits

- Prefix with what the change acts on: `is:`, `label:`, `doc:`, `todo:`,
  `ci:`.
- Imperative mood, lowercase except proper nouns. Hard-wrap at 72.
- Include _why_, not just _what_. See `git log` for examples.
- Sign your work with a `Co-Authored-By` trailer.

## Releases

cibot is origin and holds no tags. `scripts/tag vX.Y.Z` publishes one
annotated tag to the GitHub mirror, which is what a `go get` resolves.

No user should discover a break from a tag. `scripts/tag` refuses a
dirty tree and a `main` behind the farmer's, but it cannot know whether
the code is any good to the repos that will fetch it. So before tagging,
point each user at the working tree:

```sh
go mod edit -replace github.com/croaky/is=/path/to/is
go build ./... && go test ./...
```

The `replace` is a local experiment and must not be committed.

Bump each user in the same sitting. Two versions of the assertions in
use at once means the next change has to reason about both.

A tightening here is louder than in most libraries: this package decides
whether a suite passes, so a stricter `Eq` turns green tests red in every
repo that fetches it. Converting a large suite onto `Eq` has already
cost a few hundred literals typed for exactly that reason. Weigh a
change to comparison against the size of the suites that will run it,
and prefer a vet pass that flags a call to a runtime rule that fails
one.
