package main

import (
	"slices"
	"testing"
)

func sample() Article {
	return NewArticle(7, "base-name", "meta-name", "Embedding", []string{"go", "structs"})
}

func TestNewArticle(t *testing.T) {
	a := sample()
	if a.Base.ID != 7 {
		t.Errorf("Base.ID = %d, want 7", a.Base.ID)
	}
	if a.Base.Name != "base-name" {
		t.Errorf("Base.Name = %q, want %q", a.Base.Name, "base-name")
	}
	if a.Meta.Name != "meta-name" {
		t.Errorf("Meta.Name = %q, want %q", a.Meta.Name, "meta-name")
	}
	if !slices.Equal(a.Meta.Tags, []string{"go", "structs"}) {
		t.Errorf("Meta.Tags = %v, want [go structs]", a.Meta.Tags)
	}
	if a.Title != "Embedding" {
		t.Errorf("Title = %q, want %q", a.Title, "Embedding")
	}

	// The embedded types are keys in the literal, so an Article built from
	// nothing has two usable zero-valued halves.
	var zero Article
	if zero.ID != 0 || zero.Base.Name != "" || zero.Meta.Tags != nil {
		t.Errorf("the zero Article is not zero throughout: %+v", zero)
	}
}

func TestArticleID(t *testing.T) {
	if got := ArticleID(sample()); got != 7 {
		t.Errorf("ArticleID = %d, want 7", got)
	}
	if got := ArticleID(Article{}); got != 0 {
		t.Errorf("ArticleID(zero) = %d, want 0", got)
	}
}

func TestNames(t *testing.T) {
	base, meta := Names(sample())
	if base != "base-name" {
		t.Errorf("base name = %q, want %q", base, "base-name")
	}
	if meta != "meta-name" {
		t.Errorf("meta name = %q, want %q - the two Name fields must not be confused", meta, "meta-name")
	}
}

func TestAddTag(t *testing.T) {
	a := sample()
	got := AddTag(a, "embedding")
	if !slices.Equal(got.Tags, []string{"go", "structs", "embedding"}) {
		t.Errorf("AddTag tags = %v, want [go structs embedding]", got.Tags)
	}
	if len(a.Tags) != 2 {
		t.Errorf("AddTag grew the caller's article to %d tags - it takes a copy", len(a.Tags))
	}

	// Appending to a nil promoted slice works, the same as any other nil slice.
	empty := AddTag(Article{}, "first")
	if !slices.Equal(empty.Tags, []string{"first"}) {
		t.Errorf("AddTag on a zero Article = %v, want [first]", empty.Tags)
	}
}

func TestNewFeature(t *testing.T) {
	f := NewFeature(sample(), "own-name")
	if f.Name != "own-name" {
		t.Errorf("Feature.Name = %q, want %q", f.Name, "own-name")
	}
	if f.Article.Title != "Embedding" {
		t.Errorf("Feature.Article.Title = %q, want %q", f.Article.Title, "Embedding")
	}
	// Promotion goes through as many levels as it needs to, as long as nothing
	// shallower shadows the name.
	if f.Title != "Embedding" {
		t.Errorf("f.Title = %q, want the promoted %q", f.Title, "Embedding")
	}
	if f.ID != 7 {
		t.Errorf("f.ID = %d, want the promoted 7", f.ID)
	}
}

func TestFeatureNames(t *testing.T) {
	own, base := FeatureNames(NewFeature(sample(), "own-name"))
	if own != "own-name" {
		t.Errorf("own name = %q, want %q - the shallowest field wins", own, "own-name")
	}
	if base != "base-name" {
		t.Errorf("base name = %q, want %q", base, "base-name")
	}

	// Feature's own Name shadows the deeper ones rather than replacing them:
	// they are all still there, reachable by their full path.
	f := NewFeature(sample(), "shadow")
	if f.Article.Base.Name != "base-name" || f.Article.Meta.Name != "meta-name" {
		t.Errorf("shadowed names were lost: %+v", f)
	}
}

func TestRender(t *testing.T) {
	if got, want := Render(sample()), "7/meta-name: Embedding"; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
	if got, want := Render(Article{}), "0/: "; got != want {
		t.Errorf("Render(zero) = %q, want %q", got, want)
	}
}
