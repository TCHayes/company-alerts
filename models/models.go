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
func AddWatchedCompany(companyName string, user User) {
	var isInList = stringInSlice(companyName, CurrentUsers[user.Username])
	if !isInList {
		CurrentUsers[user.Username] = append(CurrentUsers[user.Username], Company{Name: companyName})
	}
}
