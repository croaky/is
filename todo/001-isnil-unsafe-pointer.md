# 001: isNil sees a nil unsafe.Pointer

Finding: `find_01HT7WExsVC5mxoFCFrpePxq` (low, CWE-697).
Title: `security: make isNil see a nil unsafe.Pointer`.

## Need

`NotNil(p)` and `NotEq(p, nil)` pass when `p` is a nil
`unsafe.Pointer`. `Nil(p)` and `Eq(p, nil)` fail on the same value.
The cause is the `isNil` function in `is.go`. Its kind switch lists
`Chan`, `Func`, `Interface`, `Map`, `Pointer`, and `Slice`. It does not
list `reflect.UnsafePointer`, so the `default` arm returns false.
`equal` sends a nil operand through `isNil`, so the `Eq` pair has the
same fault.

## Judgment

- Attacker: none. A developer writes a test on cgo, syscall, or
  low-level memory code.
- Consequence: a wrong pass. A regression that returns a nil handle
  stays green in the consumer's suite.
- Boundary: the consumer's test suite and the release it guards.

Outcome: fix it. The fix adds one kind to one `case` line. It adds no
new part, no dependency, and no measurable test time. Removal is not
possible, because `Nil` and `NotNil` are the product. Accepting the
risk saves almost nothing, and a wrong pass is the failure this
package exists to prevent.

## Change

1. Red: add tests in `is_test.go` for a nil `unsafe.Pointer` and for a
   nil value of a named type with that underlying type. Assert the
   exact failure text:
   - `NotNil(p)` fails with `p = nil, want non-nil`.
   - `NotEq(p, nil)` fails.
   - `Nil(p)` and `Eq(p, nil)` pass.
   - A non-nil `unsafe.Pointer` gives the opposite results.

   Also pin each nil-able kind (chan, func, map, pointer, slice, unsafe
   pointer) through both pairs, if the current tests do not already.
2. Green: add `reflect.UnsafePointer` to the `case` in `isNil`.
3. Keep the `default` arm. It returns false without a panic for kinds
   that cannot be nil.

The `unsafe` import in the test is standard library. The rule "nothing
outside the standard library" stays true.

## Release cost

This change makes `NotNil` and `NotEq(x, nil)` stricter on one kind.
A consumer test that passed on a nil `unsafe.Pointer` now fails. That
test was wrong. `Nil` and `Eq(x, nil)` become looser on the same kind,
so no correct test turns red. Before tagging, run the `replace` check
in `AGENTS.md` against each user. Expect zero or few failures, because
`unsafe.Pointer` values are rare outside cgo.

No `doc.go` change is needed. The text on `Nil` and `isNil` stays true.
