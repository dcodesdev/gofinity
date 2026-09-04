package main

import "testing"

// Running `go test` at image build time caches the test harness and `go vet`
// results too, not just the packages `main` imports.
func TestPrewarm(t *testing.T) {
	t.Run("subtest", func(t *testing.T) {
		if 1+1 != 2 {
			t.Fatal("arithmetic is broken")
		}
	})
}
