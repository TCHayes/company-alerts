package models

import (
	"gopkg.in/mgo.v2/bson"
)

// User ...
type User struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Username  string
	Password  string
	Phone     string
	Companies []Company `bson:"companies"`
}

func (u User) getUserCompanies() []Company {
	return u.Companies
}

// GetUserArticles ...
func (u User) GetUserArticles() []Article {
	var companies = u.getUserCompanies()
	var articles = []Article{}
	for _, company := range companies {
		articles = append(articles, company.GetCompanyArticles()...)
	}
	return articles
}
