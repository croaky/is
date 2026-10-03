# Agents guide

is holds Go test assertions that name what they check and print both
values. See `README.md` for why it exists and `doc.go` for the design.

## Writing

Write every word in ASD-STE100 Simplified Technical English (STE):
Markdown docs, code comments, commit messages, change descriptions,
and replies in an agent conversation.
See <https://en.wikipedia.org/wiki/Simplified_Technical_English>.

STE is a controlled English for technical writing: one meaning per
word, one idea per sentence, and the actor named. It is not a house
style. It exists so every reader reads a sentence the same way. That
includes a tired reader, a reader in a second language, and an agent
that matches on words.

- One idea per sentence. Keep an instruction to 20 words and a
  description to 25.
- Active voice, present tense, and the actor named: say what acts,
  rather than writing "the token is refused".
- One word, one meaning. Keep a term the same everywhere rather than
  varying it for tone.
- Use the simple verb, not a noun made from it: "run the formatter",
  not "perform execution of the formatter".
- Cut what carries nothing: "simply", "just", "note that", "in order
  to".
- Put a warning or a limit before the step it applies to.
- STE limits a sentence, not a text. Keep each sentence short, but
  keep the sentence that defines a term or connects a cause to its
  effect.
- STE permits technical names. Name the helper, the label, or the
  function, not a vague noun such as "the snapshot".

Apply it to prose, not to code: an identifier, a command, and a quoted
error message stay as they are.

## Architecture

One package at the repo root, standard library only. Two files:

- `is.go` — the seven assertions, `A`, and the two constructors. Six of
  the seven hand three strings to `fail`, which formats the one line.
  `True` reports its own, because it has no want to print.
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
go run golang.org/x/tools/cmd/goimports@v0.45.0 -local "$(go list -m)" -w .
go vet ./...
go test -trimpath -buildvcs=false -race -cover ./...
git ls-files -z '*.go' | xargs -0 gopls check -severity=hint
go run golang.org/x/tools/cmd/deadcode@v0.45.0 -test ./...
git ls-files -z 'scripts/*' '*.sh' | xargs -0 shellcheck
dprint fmt
```

`goimports` and `dprint` write here, where CI runs `goimports -l` and
`dprint check`. Fix the formatting in the change, since a CI job that
rewrote a file would have nowhere to put it.

Keep `-trimpath` on the local run. It is the case this package exists to
handle, and without it a broken lookup passes everywhere except CI.

Nothing outside the standard library is imported. A test-assertion
package that pulls in a dependency puts it in the module graph of every
repo whose tests it serves.

## Tests

Red/green TDD. `is_test.go` asserts on the failure text, because the text
is the product: a test that only checked pass and fail would not have
caught the label. Its `fake` records what would have been reported, since
`Fatal` cannot `Goexit` inside a test of `Fatal`. `external_test.go`
asserts the same way from `package is_test`, which is the nearest a
single repo gets to a caller in another module.

- Every exported helper needs a test, and the `deadcode` check enforces
  it. `deadcode -test` roots at the test binaries, so a helper no test
  calls is reported. A library's API is dead to a compiler until another
  module fetches it, so the tests are what say the surface is real.
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

## Changes

Work happens on a sockeye change. `soc checkout` allocates one and
prints a worktree; `soc edit` sets its title and description. Do the
edit before the code, not after. A change with neither is a blank row on
the dashboard and a blank `soc show`, so nobody looking at either can
tell what it is or whether it overlaps what they are about to start. A
rough sentence beats an empty one, and the description gets rewritten
before the merge anyway.

After a push, read the checks with `git push && soc show --wait`
rather than sleeping and then reading. The server holds the request open
and answers within a second of the last check, so a sleep is either time
spent waiting for an answer that already arrived or too short to reach
one. Too short is the worse half: a `soc show` that lands before the
push is recorded reports the previous commit's checks, green, about the
wrong code. `--wait` follows the commit in the worktree it runs in,
exits nonzero when a check failed, and gives up after ten minutes
(`--timeout`).

## Commits

- Prefix with what the change acts on: `is:`, `label:`, `doc:`, `ci:`.
- Imperative mood, lowercase except proper nouns. Hard-wrap at 72.
- Include _why_, not just _what_. See `git log` for examples.
- Write the subject for a teammate who has not read the code. Name the
  helper, the label, or the output, not a term the change makes up.
- Write the body as a change description (below). The first push of a
  branch with one commit copies the commit message to the change.
- Sign your work with a `Co-Authored-By` trailer.

## Change descriptions

sockeye squashes a change into one commit whose message is the
change's title and description (`soc edit`). Write the description as
that commit message.

Write for an engineer reading it a year from now. They know Go and
Git. They did not see your conversation, and they do not have the diff
open.

Put the most important fact first. A reader who stops after any
paragraph has the most important part so far. Use this order, and
leave out a part that does not apply:

1. The need. Name who uses what, what went wrong or was missing, and
   what the change does about it.
2. What a user sees now. Name the helper or the failure message.
   Say what stops happening and what a user does differently. If no
   user sees the change, say what it protects: a test, a release, or a
   cost.
3. How it works. Name each part at its first mention, with its
   kind: "the `New` Go function", "the `lint` check".
   Give the cause of a bug or the rule a feature applies, and the
   numbers that size the effect.
4. Next steps: a command to run, a follow-up plan by its path, or a
   known gap.

Most changes fit in three to five short paragraphs. To keep a
description short and clear:

- Define a project term at its first use, or use a plainer word.
- Leave out file lists, line counts, and code sizes. The diff shows
  them.
- Leave out what a reader of `main` cannot use: options you did not
  take, each edge case the tests cover, and "tests pass". Put a design
  argument in a code comment or a plan, and give its path.
- Use plain paragraphs. Git strips a `#` line as a comment, so a
  Markdown header disappears.
- Hard-wrap at 72 columns. Don't backslash-escape backticks. Use a
  quoted heredoc such as `<<'EOF'` with `soc edit`.
- Keep `Co-Authored-By` on the commits, not in the description. The
  merge collects the trailers from the commits it squashes.

## Releases

sockeye is origin and holds no tags. `scripts/tag vX.Y.Z` publishes one
annotated tag to the GitHub mirror, which is what a `go get` resolves.

No user should discover a break from a tag. `scripts/tag` refuses a
dirty tree and a `main` behind the server's, but it cannot know whether
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
