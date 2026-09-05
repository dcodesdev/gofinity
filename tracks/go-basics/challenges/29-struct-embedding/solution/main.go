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
	// An embedded field's key in a literal is its type name, and there is no
	// way to set a promoted field from the outer literal directly.
	return Article{
		Base:  Base{ID: id, Name: baseName},
		Meta:  Meta{Name: metaName, Tags: tags},
		Title: title,
	}
}

// ArticleID returns the article's ID through the promoted field.
func ArticleID(a Article) int {
	// ID appears in exactly one embedded struct, so it promotes: a.Base.ID
	// would work too and says the same thing.
	return a.ID
}

// Names returns the two Name fields the article carries, base first.
func Names(a Article) (base, meta string) {
	// a.Name does not compile: "ambiguous selector a.Name". Both candidates sit
	// at the same depth, so neither wins and the compiler refuses to guess.
	return a.Base.Name, a.Meta.Name
}

// AddTag returns a copy of a with tag appended to its promoted Tags field.
func AddTag(a Article, tag string) Article {
	// Tags is unambiguous, so the promoted selector works for assignment as
	// well as for reading. a is a copy, so this cannot touch the caller's.
	a.Tags = append(a.Tags, tag)
	return a
}

// NewFeature wraps an article, giving the feature its own name.
func NewFeature(a Article, name string) Feature {
	return Feature{Article: a, Name: name}
}

// FeatureNames returns the feature's own name and the Base.Name buried two
// levels down inside it.
func FeatureNames(f Feature) (own, base string) {
	// f.Name is unambiguous now: depth 0 beats anything deeper, so Feature's
	// own field shadows both of the ones inside the embedded Article.
	// f.Base still promotes through Article, which is why the second one does
	// not have to be spelled f.Article.Base.Name.
	return f.Name, f.Base.Name
}

// Render formats an article as "ID/Meta.Name: Title".
func Render(a Article) string {
	return fmt.Sprintf("%d/%s: %s", a.ID, a.Meta.Name, a.Title)
}

func main() {
	a := NewArticle(7, "base", "meta", "Embedding", []string{"go"})
	fmt.Println(Render(a))
	fmt.Println(FeatureNames(NewFeature(a, "own")))
}
