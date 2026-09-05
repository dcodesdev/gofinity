# Struct Embedding

Go has no inheritance. What it has instead is **embedding**: a field declared
with a type but no name.

```go
type Base struct {
	ID   int
	Name string
}

type Article struct {
	Base           // embedded: no field name, just the type
	Title string
}
```

`Article` now has a field literally called `Base`, plus a shortcut: `Base`'s own
fields are **promoted**, so `a.ID` works and means `a.Base.ID`. Nothing is
copied into `Article` and no new type relationship is created. An `Article` is
not a `Base` and cannot be used where one is wanted; it just knows a shorter way
to spell one of its fields.

## Building one

An embedded field's key in a composite literal is its **type name**:

```go
a := Article{
	Base:  Base{ID: 7, Name: "base"},
	Title: "Embedding",
}
```

You cannot set a promoted field from the outer literal - `Article{ID: 7}` does
not compile. The nesting is real; only the selector is flattened.

Assignment through a promoted field does work, though, because a promoted
selector is a full selector:

```go
a.ID = 8       // same as a.Base.ID = 8
```

## Shadowing and ambiguity

Promotion is a lookup, and the rule is **shallowest wins**. A field the outer
struct declares itself is at depth 0 and beats anything embedded, which is at
depth 1 or more:

```go
type Feature struct {
	Article        // Article, Base and Meta all live in here
	Name string    // depth 0: this is what f.Name means
}
```

The shadowed fields are not gone. `f.Article.Base.Name` still reaches the one
underneath. Shadowing changes what the short spelling means, nothing else.

When two candidates sit at the **same** depth, neither wins. That is not an
error in the type - the struct is perfectly legal - but the *selector* is:

```go
type Article struct {
	Base           // has Name
	Meta           // also has Name
	Title string
}

a.Name    // compile error: ambiguous selector a.Name
a.ID      // fine, only Base has it
a.Tags    // fine, only Meta has it
```

So an ambiguity costs you only the ambiguous name, and you fix it by spelling
out which one you meant. This is exactly why embedding is composition rather
than inheritance: there is no method resolution order and no rule about which
parent wins, because the compiler refuses to have an opinion.

The same promotion rules will apply to **methods** in the next lesson, and that
is where embedding earns its keep - an embedded type's methods become the outer
type's methods, which is how a struct satisfies an interface by embedding
something that already does.

## Task

`main.go` declares `Base`, `Meta`, `Article` and `Feature` for you. Fill in the
seven functions.

## Hints

- `NewArticle` must name the embedded types as keys: `Base: Base{...}`.
- `Names` is the one that cannot use the short spelling. `a.Name` will not
  compile at all - use `a.Base.Name` and `a.Meta.Name`.
- `AddTag` can use the promoted `a.Tags` on both sides of the assignment,
  because only `Meta` declares it.
- In `FeatureNames`, `f.Name` is the feature's own, and `f.Base.Name` reaches
  through `Article` to `Base` in one step, because `Base` itself promotes.
- If you get "ambiguous selector", you have found the trap the exercise is
  about: name the level you meant.
