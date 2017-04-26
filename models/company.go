package models

// Company struct ...
type Company struct {
	Name   string
	Ticker string
}

// GetCompanyArticles ...
func (c Company) GetCompanyArticles() []Article {
	return CurrentArticlesMap[c.Name]
}
