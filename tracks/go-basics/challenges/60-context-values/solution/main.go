package main

import (
	"context"
	"errors"
	"fmt"
)

// contextKey is the type of every key this file stores in a context.
//
// It is unexported, so no other package can construct one, so no other package
// can collide with these keys or read them by accident. A plain string would
// be readable and writable by anyone who guessed the same word.
type contextKey int

const (
	requestIDKey contextKey = iota
	userKey
)

// User is who a request is running as.
type User struct {
	Name  string
	Admin bool
}

// ErrForbidden is returned when a request is not allowed to go any further.
var ErrForbidden = errors.New("forbidden")

// WithRequestID returns a child of ctx carrying id.
//
// It stores id even when it is empty: "present and empty" and "absent" are
// different answers, and only the caller knows which one matters.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request id ctx carries, and whether it carries one.
//
// ctx.Value returns an any, so it needs a type assertion - the two-value form,
// because a missing value is nil and a nil any asserts to nothing.
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// WithUser returns a child of ctx carrying u.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// UserFrom returns the user ctx carries, and whether it carries one.
//
// With no user it returns the zero User, which is not an admin - the safe
// answer if a caller ignores the second result.
func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

// Describe renders what ctx carries as one line for a log:
//
//	request=none user=anonymous
//	request=r1 user=Ada
//	request=r1 user=Ada (admin)
//
// A missing request id is "none" and a missing user is "anonymous".
func Describe(ctx context.Context) string {
	id, ok := RequestID(ctx)
	if !ok {
		id = "none"
	}
	user := "anonymous"
	if u, ok := UserFrom(ctx); ok {
		user = u.Name
		if u.Admin {
			user += " (admin)"
		}
	}
	return fmt.Sprintf("request=%s user=%s", id, user)
}

// Handler is a unit of work that runs under a context.
type Handler func(ctx context.Context) error

// Middleware wraps a Handler in another Handler.
type Middleware func(next Handler) Handler

// Chain wraps h in every middleware and returns the result.
//
// The first middleware listed is the outermost, so Chain(h, a, b) runs a, then
// b, then h - the order they are written is the order they run. Chain with no
// middleware returns h itself.
func Chain(h Handler, mw ...Middleware) Handler {
	// Wrapping from the last middleware backwards leaves mw[0] on the outside,
	// which is the order the call site reads in.
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// RequireAdmin is a Middleware that calls next only when ctx carries an admin
// user, and otherwise returns ErrForbidden without running it at all.
func RequireAdmin(next Handler) Handler {
	return func(ctx context.Context) error {
		u, ok := UserFrom(ctx)
		if !ok || !u.Admin {
			return ErrForbidden
		}
		return next(ctx)
	}
}

func main() {
	ctx := WithUser(WithRequestID(context.Background(), "r1"), User{Name: "Ada", Admin: true})
	fmt.Println(Describe(ctx))

	h := Chain(func(ctx context.Context) error {
		fmt.Println("handling", Describe(ctx))
		return nil
	}, RequireAdmin)

	fmt.Println(h(ctx))
	fmt.Println(h(context.Background()))
}
