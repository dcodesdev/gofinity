package main

import (
	"context"
	"errors"
	"testing"
)

func TestRequestIDRoundTrips(t *testing.T) {
	ctx := WithRequestID(context.Background(), "r1")
	got, ok := RequestID(ctx)
	if !ok {
		t.Fatal("RequestID reported no id on a context that carries one")
	}
	if got != "r1" {
		t.Errorf("RequestID = %q, want %q", got, "r1")
	}
}

func TestRequestIDIsAbsentOnABareContext(t *testing.T) {
	got, ok := RequestID(context.Background())
	if ok {
		t.Errorf("RequestID(Background()) reported %q, want no id", got)
	}
	if got != "" {
		t.Errorf("RequestID(Background()) = %q, want the empty string alongside false", got)
	}
}

func TestRequestIDTellsEmptyFromAbsent(t *testing.T) {
	// An empty id was still set by somebody, and that is a different fact from
	// nobody having set one. Store what you were given.
	got, ok := RequestID(WithRequestID(context.Background(), ""))
	if !ok {
		t.Error("RequestID reported no id after WithRequestID(ctx, \"\"), want an empty id that is present")
	}
	if got != "" {
		t.Errorf("RequestID = %q, want the empty string that was stored", got)
	}
}

func TestAStringKeyCannotReachTheRequestID(t *testing.T) {
	// Somebody else's package, storing under the most obvious key there is.
	//nolint
	other := context.WithValue(context.Background(), "requestID", "theirs")

	if got, ok := RequestID(other); ok {
		t.Errorf(`RequestID found %q under a plain string key - your keys must be an unexported type nobody else can name`, got)
	}

	// And the other direction: our value must not be readable by guessing.
	ours := WithRequestID(context.Background(), "ours")
	if v := ours.Value("requestID"); v != nil {
		t.Errorf(`ctx.Value("requestID") = %v, want nil - a string key must not find your value`, v)
	}
}

func TestUserRoundTrips(t *testing.T) {
	want := User{Name: "Ada", Admin: true}
	got, ok := UserFrom(WithUser(context.Background(), want))
	if !ok {
		t.Fatal("UserFrom reported no user on a context that carries one")
	}
	if got != want {
		t.Errorf("UserFrom = %+v, want %+v", got, want)
	}
}

func TestUserIsAbsentOnABareContext(t *testing.T) {
	got, ok := UserFrom(context.Background())
	if ok {
		t.Errorf("UserFrom(Background()) reported %+v, want no user", got)
	}
	if got != (User{}) {
		t.Errorf("UserFrom(Background()) = %+v, want the zero User - a caller that ignores ok must not get an admin", got)
	}
}

func TestTheTwoValuesDoNotOverwriteEachOther(t *testing.T) {
	ctx := WithUser(WithRequestID(context.Background(), "r1"), User{Name: "Ada"})

	if id, ok := RequestID(ctx); !ok || id != "r1" {
		t.Errorf("RequestID = (%q, %v) after a user was added, want (\"r1\", true) - each value needs its own key", id, ok)
	}
	if u, ok := UserFrom(ctx); !ok || u.Name != "Ada" {
		t.Errorf("UserFrom = (%+v, %v), want Ada", u, ok)
	}
}

func TestADerivedValueShadowsWithoutChangingTheParent(t *testing.T) {
	parent := WithRequestID(context.Background(), "outer")
	child := WithRequestID(parent, "inner")

	if id, ok := RequestID(child); !ok || id != "inner" {
		t.Errorf("the child sees %q, want %q", id, "inner")
	}
	if id, ok := RequestID(parent); !ok || id != "outer" {
		t.Errorf("the parent sees %q after a child overrode it, want %q - contexts are immutable", id, "outer")
	}
}

func TestValuesSurviveCancellationWrappers(t *testing.T) {
	base := WithUser(WithRequestID(context.Background(), "r1"), User{Name: "Ada"})
	ctx, cancel := context.WithCancel(base)
	defer cancel()

	// Value lookup walks up the chain, so a cancellable child still sees
	// everything its ancestors carry.
	if id, ok := RequestID(ctx); !ok || id != "r1" {
		t.Errorf("RequestID through a WithCancel child = (%q, %v), want (\"r1\", true)", id, ok)
	}
	if u, ok := UserFrom(ctx); !ok || u.Name != "Ada" {
		t.Errorf("UserFrom through a WithCancel child = (%+v, %v), want Ada", u, ok)
	}
}

func TestDescribe(t *testing.T) {
	ada := User{Name: "Ada"}
	admin := User{Name: "Ada", Admin: true}
	cases := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"nothing", context.Background(), "request=none user=anonymous"},
		{"id only", WithRequestID(context.Background(), "r1"), "request=r1 user=anonymous"},
		{"user only", WithUser(context.Background(), ada), "request=none user=Ada"},
		{"both", WithUser(WithRequestID(context.Background(), "r1"), ada), "request=r1 user=Ada"},
		{"admin", WithUser(WithRequestID(context.Background(), "r1"), admin), "request=r1 user=Ada (admin)"},
	}
	for _, c := range cases {
		if got := Describe(c.ctx); got != c.want {
			t.Errorf("Describe(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestChainWithNoMiddlewareIsTheHandler(t *testing.T) {
	called := false
	h := Chain(func(context.Context) error {
		called = true
		return nil
	})
	if err := h(context.Background()); err != nil {
		t.Errorf("the handler returned %v, want nil", err)
	}
	if !called {
		t.Error("Chain with no middleware never called the handler")
	}
}

func TestChainRunsMiddlewareInTheOrderItIsWritten(t *testing.T) {
	var order []string
	mark := func(name string) Middleware {
		return func(next Handler) Handler {
			return func(ctx context.Context) error {
				order = append(order, name)
				return next(ctx)
			}
		}
	}

	h := Chain(func(context.Context) error {
		order = append(order, "handler")
		return nil
	}, mark("a"), mark("b"))

	if err := h(context.Background()); err != nil {
		t.Fatalf("the chain returned %v, want nil", err)
	}
	want := []string{"a", "b", "handler"}
	if len(order) != len(want) {
		t.Fatalf("the chain ran %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("the chain ran %v, want %v - the first middleware listed is the outermost", order, want)
		}
	}
}

func TestChainPassesTheContextItIsGiven(t *testing.T) {
	// A middleware that adds a value must be visible to everything inside it.
	tag := func(next Handler) Handler {
		return func(ctx context.Context) error {
			return next(WithRequestID(ctx, "added"))
		}
	}

	var seen string
	h := Chain(func(ctx context.Context) error {
		seen, _ = RequestID(ctx)
		return nil
	}, tag)

	if err := h(context.Background()); err != nil {
		t.Fatalf("the chain returned %v, want nil", err)
	}
	if seen != "added" {
		t.Errorf("the handler saw request id %q, want %q - pass ctx down rather than the one you were built with", seen, "added")
	}
}

func TestChainReturnsTheHandlerError(t *testing.T) {
	fail := errors.New("boom")
	pass := func(next Handler) Handler { return next }
	h := Chain(func(context.Context) error { return fail }, pass, pass)
	if err := h(context.Background()); !errors.Is(err, fail) {
		t.Errorf("the chain returned %v, want the handler's own error", err)
	}
}

func TestRequireAdminLetsAnAdminThrough(t *testing.T) {
	ctx := WithUser(context.Background(), User{Name: "Ada", Admin: true})
	called := false
	h := Chain(func(context.Context) error {
		called = true
		return nil
	}, RequireAdmin)

	if err := h(ctx); err != nil {
		t.Errorf("RequireAdmin returned %v for an admin, want nil", err)
	}
	if !called {
		t.Error("RequireAdmin did not call the handler for an admin")
	}
}

func TestRequireAdminStopsEveryoneElse(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
	}{
		{"no user", context.Background()},
		{"not an admin", WithUser(context.Background(), User{Name: "Ada"})},
	}
	for _, c := range cases {
		called := false
		h := Chain(func(context.Context) error {
			called = true
			return nil
		}, RequireAdmin)

		if err := h(c.ctx); !errors.Is(err, ErrForbidden) {
			t.Errorf("RequireAdmin with %s returned %v, want ErrForbidden", c.name, err)
		}
		if called {
			t.Errorf("RequireAdmin with %s ran the handler anyway - a rejected request must not reach it", c.name)
		}
	}
}
