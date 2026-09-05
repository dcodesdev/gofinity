# Zero Values

Go has no uninitialised memory. A variable declared without a value is not
garbage and it is not undefined: it is that type's **zero value**, and the
compiler guarantees it.

| Type | Zero value |
| --- | --- |
| `int`, `int64`, `uint`, ... | `0` |
| `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| slice, map, pointer, function, channel, interface | `nil` |

That guarantee is load-bearing. It is why `var total int` needs no `= 0`, why
`append` to a `nil` slice just works, and why so much Go code has no
constructor: the zero value is already useful.

## Task

Implement three functions in `main.go`.

### `ZeroReport() string`

Declare one variable of each type with `var` and **no initial value**, then
return exactly these seven lines, joined by `"\n"` with no trailing newline:

```
int: 0
float64: 0
string: ""
bool: false
slice: [] nil=true len=0
map: map[] nil=true len=0
pointer: <nil>
```

Print the string with `%q` so the empty one is visible as `""`. Everything else
is `%v`, which prints a `nil` slice as `[]`, a `nil` map as `map[]` and a `nil`
pointer as `<nil>`.

### `SumOrZero(nums []int) int`

Add up the numbers. A `nil` slice must sum to `0` **without a special case**:
ranging over `nil` is legal and runs zero times.

### `GrowFromNil(nums ...int) []int`

Start from `var out []int`, which is `nil`, `append` each number to it and
return it. With no arguments the result stays `nil`.

## Hints

- A `nil` slice and an empty slice behave the same for `len`, `range` and
  `append`. They differ only when you compare against `nil`, which is why the
  report prints both `nil=` and `len=`.
- `append` is the reason `nil` is a usable starting point: it allocates the
  backing array on first use, so `var out []int` needs no `make`.
- `%v` chooses a sensible format for whatever you give it. `%q` quotes a
  string. Getting `int: %!v(MISSING)` in the failure output means the verb
  count and the argument count disagree.
- One `fmt.Sprintf` with a multi-line format string is the shortest way to
  build the report. A raw literal in backticks would also work, but backticks
  cannot contain `\n` escapes, so an ordinary quoted string is simpler here.
- `var` is the right declaration in this challenge precisely because `:=`
  cannot express "give me the zero value": `:=` always needs a value on the
  right.
