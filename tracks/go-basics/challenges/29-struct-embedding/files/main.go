package main

import "fmt"

// Base and Meta both declare a Name field. That is deliberate: it is what makes
// the promoted a.Name ambiguous inside Article.
type Base struct {
	ID   int
	Name string
}

type Meta struct {
	Name string
	Tags []string
}

// Article embeds both. ID and Tags promote cleanly, because each appears once.
// Name does not: writing a.Name is a compile error, and it has to be spelled
// a.Base.Name or a.Meta.Name.
type Article struct {
	Base
	Meta
	Title string
}

// Feature embeds an Article and declares its own Name. The shallower field
// wins, so f.Name is Feature's own and the ambiguity two levels down never
// comes up.
type Feature struct {
	Article
	Name string
}

// NewArticle builds an Article. baseName goes in Base.Name, metaName in
// Meta.Name, and the tags are stored as given.
func NewArticle(id int, baseName, metaName, title string, tags []string) Article {
	// TODO
	return Article{}
}

// ArticleID returns the article's ID through the promoted field.
func ArticleID(a Article) int {
	// TODO
	return 0
}

// Names returns the two Name fields the article carries, base first.
func Names(a Article) (base, meta string) {
	// TODO
	return "", ""
}

// AddTag returns a copy of a with tag appended to its promoted Tags field.
func AddTag(a Article, tag string) Article {
	// TODO
	return Article{}
}

// NewFeature wraps an article, giving the feature its own name.
func NewFeature(a Article, name string) Feature {
	// TODO
	return Feature{}
}

// FeatureNames returns the feature's own name and the Base.Name buried two
// levels down inside it.
func FeatureNames(f Feature) (own, base string) {
	// TODO
	return "", ""
}

// Render formats an article as "ID/Meta.Name: Title".
func Render(a Article) string {
	// TODO
	return ""
}

func main() {
	a := NewArticle(7, "base", "meta", "Embedding", []string{"go"})
	fmt.Println(Render(a))
	fmt.Println(FeatureNames(NewFeature(a, "own")))
}
