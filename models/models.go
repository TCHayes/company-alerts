package models

// CurrentUsers ... server state for users currently using our system
var CurrentUsers = make(map[string][]Company)

// CurrentArticlesMap ... server state for articles that we are consistently parsing
var CurrentArticlesMap = make(map[string][]Article)

func stringInSlice(a string, list []Company) bool {
	for _, b := range list {
		if b.Name == a {
			return true
		}
	}
	return false
}

// AddWatchedCompany adds a company to the watch list for a given user
func AddWatchedCompanies(companyNames []string, user User) {
	companies := []Company{}
	for _, c := range companyNames {
		companies = append(companies, Company{Name: c})
	}
	CurrentUsers[user.Username] = companies
	user.Companies = companies
}
