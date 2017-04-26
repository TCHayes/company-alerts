package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/getlatest", getLatestHandler)
	http.ListenAndServe(":8080", nil)
}

var currentUsers = []User{}
var currentArticlesMap = make(map[string][]Article)

type Article struct {
	Author      string
	Title       string
	Description string
	Url         string
	UrlToImage  string
	PublishedAt string
}

type NewsApiResponse struct {
	Status   string
	Source   string
	SortBy   string
	Articles []Article
}

type Company struct {
	Name   string
	Ticker string
}

type User struct {
	Username  string
	Password  string
	Companies []Company
}

func getLatestHandler(w http.ResponseWriter, r *http.Request) {
	resp, _ := http.Get("https://newsapi.org/v1/articles?source=techcrunch&sortBy=latest&apiKey=bd7079b419d3439ca765e70919837e9d")
	defer resp.Body.Close()
	target := NewsApiResponse{}
	_ = json.NewDecoder(resp.Body).Decode(&target)
	fmt.Println(target.Articles)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (c Company) getCompanyArticles() []Article {
	return currentArticlesMap[c.Name]
}

func (u User) getUserCompanies() []Company {
	return u.Companies
}

func (u User) getUserArticles() []Article {
	var companies = u.getUserCompanies()
	var articles = []Article{}
	for _, company := range companies {
		articles = append(articles, company.getCompanyArticles()...)
	}
	return articles
}

func addWatchedCompany(companyName string, userId string) {
	var isInList = stringInSlice(companyName, currentCompanyMap[userId])
	if !isInList {
		currentCompanyMap[userId] = append(currentCompanyMap[userId], companyName)
	}
}
