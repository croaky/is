# Todo

Each file is one sockeye change. Number the files in the order the
changes land. The commit that finishes a change deletes its file.
The next file starts past the highest number already used.

Every change runs the `Checkfile` list before it opens.

## Threat model

is is a test-assertion library. It runs only inside the `go test`
binary of a module that requires it. It reads no network input and
holds no secret. The only input is the values and source text of the
caller's own tests.

So a security finding here names no outside attacker. The consequence
to weigh is a wrong verdict: a test that passes when it must fail, or
fails when it must pass. The boundary is the consumer's test suite and
the release that the suite guards. A wrong pass is the worse half,
because nobody sees it.

## Open changes

- `001-isnil-unsafe-pointer.md`: `isNil` misses a nil `unsafe.Pointer`.
  Finding `find_01HT7WExsVC5mxoFCFrpePxq`.
