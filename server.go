package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TCHayes/company-alerts/dao"
	"github.com/TCHayes/company-alerts/models"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var session *mgo.Session

type addAlertRequest struct {
	Username string
	Company  string
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func createResponse(data interface{}, w *http.ResponseWriter) {
	response := bson.M{"data": data, "success": true}
	json.NewEncoder(*w).Encode(&response)
	return
}

func createFailureResponse(reason string, w *http.ResponseWriter) {
	response := bson.M{"reason": reason, "success": false}
	json.NewEncoder(*w).Encode(&response)
}

func bsonify(r *http.Request) bson.M {
	body := bson.M{}
	check(json.NewDecoder(r.Body).Decode(&body))
	return body
}

func handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := session.DB("company-alerts").C("user")
	request := bsonify(r)
	doc := models.User{}
	readSuccess := dao.ReadOne(c, request, &doc)
	if readSuccess {
		if doc.Username == request["username"] && doc.Password == request["password"] {
			createResponse(doc, &w)
			return
		}
		createFailureResponse("Could not find a matching username and password combination", &w)
		return
	}
	createFailureResponse("Something went wrong", &w)
	return
}

func handleRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := session.DB("company-alerts").C("user")
	request := bsonify(r)

	existsInDB := dao.CheckExisting(c, request, models.User{})
	if !existsInDB {
		writeSuccess := dao.Create(c, request)
		if writeSuccess {
			createResponse(request, &w)
			return
		}
		createFailureResponse("Something went wrong creating your user account", &w)
		return
	}
	createFailureResponse("User already exists in database", &w)
}

func handleAddalert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := session.DB("company-alerts").C("user")

	body := addAlertRequest{}
	_ = json.NewDecoder(r.Body).Decode(&body)

	user := models.User{}
	readSuccess := dao.ReadOne(c, bson.M{"username": body.Username}, &user)
	if readSuccess == true {
		models.AddWatchedCompany(body.Company, user)
		createResponse(bson.M{"success": true}, &w)
		return
	}
	createFailureResponse("Something went wrong", &w)
	return
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func sendAlert(user string, company string) {
	// TODO send twilio text
	return
}
func evaluateSentiment(text string) bool {
	return true
}

func alertLoop(t time.Time) {
	for user, companies := range models.CurrentUsers {
		var userAlertMap = make(map[string]bool)
		for _, company := range companies {
			articles := company.GetCompanyArticles()
			var text = ""
			for _, articles := range articles {
				text += articles.Description
			}
			if evaluateSentiment(text) {
				userAlertMap[company.Name] = true
			}
			userAlertMap[company.Name] = false
		}
		for company, needAlert := range userAlertMap {
			if needAlert {
				sendAlert(user, company)
			}
		}
	}
}

type newsAPIResponse struct {
	Status   string
	Source   string
	SortBy   string
	Articles []models.Article
}

func refreshTechCrunchArticles(t time.Time) {
	for companyName := range models.CurrentArticlesMap {
		resp, _ := http.Get("https://newsapi.org/v1/articles?source=techcrunch&sortBy=latest&apiKey=bd7079b419d3439ca765e70919837e9d")
		defer resp.Body.Close()
		target := newsAPIResponse{}
		_ = json.NewDecoder(resp.Body).Decode(&target)
		for i, article := range target.Articles {
			if !strings.Contains(article.Description, companyName) {
				target.Articles = append(target.Articles[:i], target.Articles[i+1:]...)
			}
		}
		models.CurrentArticlesMap[companyName] = target.Articles
	}
}

// func refreshTheNextWebArticles(t time.Time) {
// TODO
// }

func main() {
	var connErr error
	session, connErr = mgo.Dial("mongodbstring")
	check(connErr)

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	router := httprouter.New()

	router.POST("/api/addalert", handleAddalert)
	router.POST("/api/authenticate", handleLogin)
	router.POST("/api/register", handleRegister)

	doEvery(250000*time.Millisecond, refreshTechCrunchArticles)
	doEvery(300000*time.Millisecond, alertLoop)

}
