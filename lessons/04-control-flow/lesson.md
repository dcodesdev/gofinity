# Control flow

Three keywords steer every Go program: `if`, `switch` and `for`. There is no
`while`, no `do`, no ternary, no comprehension and no exception handling. The
list is short because each one was designed to have a single obvious form, and
the payoff is that control flow in unfamiliar Go code reads the same as control
flow in yours.

## if

No parentheses around the condition, and braces are never optional:

```go
if score >= 90 {
	return "A"
} else if score >= 80 {
	return "B"
} else {
	return "C"
}
```

The condition must be a `bool`. Go has no truthiness at all: `if n` where `n` is
an `int` does not compile, nor does `if s` for a string or `if p` for a pointer.
You write `if n != 0`, `if s != ""`, `if p != nil`. That is three more
characters and one less category of bug, because there is never a question of
which values count as false.

The opening brace has to sit on the same line as the `if`. This is not a style
preference you can argue with; Go inserts semicolons at the end of lines, so a
brace on its own line terminates the statement and the code will not compile.

### The short statement

An `if` may run one statement before its condition, separated by a semicolon:

```go
if err := doWork(); err != nil {
	return err
}
```

`err` exists only inside the `if` and its `else`. This is the single most common
shape in real Go code, and its point is scope: the error variable does not leak
into the rest of the function where a later branch might reuse it by accident.
`switch` and `for` take the same short statement.

## switch

Go's `switch` fixes the two things everyone dislikes about the C version.

**Cases do not fall through.** There is no `break`, and forgetting one cannot
cost you an afternoon of debugging. A case ends where the next one begins.

**A case can list several values**, which collapses a stack of near-identical
branches into one line:

```go
switch day {
case "sat", "sun":
	return "weekend"
case "mon", "tue", "wed", "thu", "fri":
	return "weekday"
default:
	return "unknown"
}
```

`default` is optional and can go anywhere, though the bottom is conventional. A
switch that matches nothing and has no `default` simply does nothing.

### switch with no expression

Leave the expression out and the switch compares against `true`, so every case
becomes a condition of its own:

```go
switch {
case score < 0 || score > 100:
	return "?"
case score >= 90:
	return "A"
case score >= 80:
	return "B"
default:
	return "F"
}
```

This is the same logic as an if/else if chain with the repetition removed, and
it is the form to reach for whenever you are testing ranges of one value. Cases
are still evaluated top to bottom, so ordering matters exactly as much as it
does in a chain: put the out-of-range guard below `score >= 90` and `101` never
reaches it.

### fallthrough

The old behaviour is still available when you actually want it, as an explicit
statement:

```go
switch tier {
case "gold":
	perks = append(perks, "lounge")
	fallthrough
case "silver":
	perks = append(perks, "priority")
	fallthrough
case "bronze":
	perks = append(perks, "points")
}
```

`fallthrough` must be the last statement in its case, it cannot appear in the
final case, and it jumps into the next case's body **without testing that
case's condition**. It is rare. The one shape where it earns its keep is a
ladder like this, where each level genuinely includes everything below it.

There is a third form, the type switch, which switches on a value's dynamic
type. It needs interfaces to be interesting, so it waits for that concept.

## for

Go has exactly one loop keyword, and it covers every shape by dropping clauses.

**Three clauses** - init, condition, post - is the counter:

```go
for i := 1; i <= n; i++ {
	total += i
}
```

`i` is scoped to the loop and does not exist after it.

**One clause** is what other languages spell `while`:

```go
for n != 1 {
	n /= 2
}
```

**No clauses** loops forever, and you leave it with `break` or `return`:

```go
for {
	p *= 2
	if p >= n {
		break
	}
}
```

**`range`** walks a collection. Over a slice it yields index and value:

```go
for i, v := range items { }
for i := range items { }      // index only
for _, v := range items { }   // value only
```

The value is a **copy** of the element, so assigning to `v` changes nothing in
the slice. Reach for `items[i] = ...` when you mean to write back. Ranging over
a `nil` or empty slice runs the body zero times, which is usually the special
case you were about to write an `if` for.

`range` also works over maps, strings, channels, and over an integer:
`for i := range 10` counts 0 to 9. Ranging a string yields runes rather than
bytes, which the basic types concept covered.

### break, continue and labels

`break` ends the innermost loop; `continue` skips to its next iteration, running
the post statement on the way. In a three-clause loop the counter still
advances, so `continue` is safe; in a condition-only loop, a `continue` above
the line that advances the variable is an infinite loop.

When you need to leave an *outer* loop from inside an inner one, label it:

```go
search:
	for r, row := range grid {
		for _, cell := range row {
			if cell == target {
				break search
			}
		}
	}
```

A plain `break` there would only end the inner loop and the search would carry
on with the next row, quietly returning the last match instead of the first.
The label names a loop; it is not a `goto`, and it is the honest alternative to
threading a `found` flag through two conditions.

Go does have `goto`. You will not need it.

## What is missing, and why

- **No ternary.** `a ? b : c` does not exist. Write the `if`, or a small helper
  function. Go's answer is that a three-line `if` is never the thing making your
  code hard to read.
- **No `while`.** `for cond` is the same loop with a shorter name.
- **No exceptions.** Errors are values you return and check, which is its own
  concept two lessons from now. `panic` exists for genuinely unrecoverable
  states and is not flow control.

## Further reading

- [Statements](https://go.dev/ref/spec#Statements) - `if`, `switch`, `for`,
  `break` and `continue`, labelled statements included, in one place.
- [For statements](https://go.dev/ref/spec#For_statements) - the three loop forms
  and what `range` yields for each kind of value.
- [Switch statements](https://go.dev/ref/spec#Switch_statements) - expressionless
  switches, several cases per clause, and where `fallthrough` is legal.
- [Effective Go: control structures](https://go.dev/doc/effective_go#control-structures) -
  the idioms behind the rules, including why there is no ternary.

## Practise

Three challenges. The first is `if`, `else if` and the modulo operator in the
smallest program that needs all three. The second is every form of `switch`,
including the one place `fallthrough` belongs. The third asks for six loops in
six different shapes, ending with the labelled break.
