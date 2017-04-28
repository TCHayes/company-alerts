package models

// Company struct ...
type Company struct {
	Name string `bson:"name"`
}

// GetCompanyArticles ...
func (c Company) GetCompanyArticles() []Article {
	return CurrentArticlesMap[c.Name]
}
