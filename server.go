package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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
	Username  string
	Companies []string
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
	// bsonify takes in JSON and returns BSON?
	// M is a convenient alias (type) for a map[string]interface{} map,
	//	useful for dealing with BSON in a native way.
	body := bson.M{}
	//check just makes sure there isn't an error with whatever code is passed in
	//is the argument within Decode (body) where the decoded info is stored?
	check(json.NewDecoder(r.Body).Decode(&body))
	return body
}

func handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := session.DB("company-alerts").C("users")
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
	c := session.DB("company-alerts").C("users")
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
	// set up headers
	w.Header().Set("Content-Type", "application/json")
	// connect to DB?
	c := session.DB("company-alerts").C("users")
	// create new, empty instance of addAlertRequest struct
	body := addAlertRequest{}
	// decode the request Body and store the decoded bson format in local body var
	_ = json.NewDecoder(r.Body).Decode(&body)
	// create new, empty instance of User struct
	user := models.User{}
	// look to make sure username exists in DB
	readSuccess := dao.ReadOne(c, bson.M{"username": body.Username}, &user)
	fmt.Println(readSuccess)
	if readSuccess == true {
		// if the user is found in the DB, call the AddWatchedCompany method
		models.AddWatchedCompanies(body.Companies, user)
		dao.UpdateOne(c, bson.M{"username": body.Username}, bson.M{"$set": bson.M{"companies": body.Companies}})
		createResponse(bson.M{"success": true}, &w)
		return
	}
	createFailureResponse("Something went wrong", &w)
	return
}

func handleIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.FileServer(http.Dir("./views"))
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

func evaluateSentiment(text string) int64 {
	fileHandle, _ := os.Create("article_desc.txt")
	defer fileHandle.Close()
	writer := bufio.NewWriter(fileHandle)
	fmt.Fprintln(writer, text)
	writer.Flush()
	cmd := exec.Command("node", "sentiscript.js", "article_desc.txt")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	score, _ := strconv.ParseInt(strings.TrimSpace(out.String()), 10, 64)
	return score
}

func alertLoop(t time.Time) {
	for user, companies := range models.CurrentUsers {
		var userAlertMap = make(map[string]bool)
		for _, company := range companies {
			articles := company.GetCompanyArticles()
			var text = ""
			for _, article := range articles {
				text += article.Description
			}
			if evaluateSentiment(text) < 0 {
				userAlertMap[company.Name] = true
			} else {
				userAlertMap[company.Name] = false
			}
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

//TODO NewsAPI appears to return the same cats for each source (Status,
//	Source, SortBy, Articles) and same API key, so if we abstract out
//	the source portion of the url, we can use the same refreshArticles func
//	for whichever sources we wish to include.
func refreshTheNextWebArticles(t time.Time) {
	for companyName := range models.CurrentArticlesMap {
		resp, _ := http.Get("https://newsapi.org/v1/articles?source=the-next-web&sortBy=latest&apiKey=bd7079b419d3439ca765e70919837e9d")
		defer resp.Body.Close()
		target := newsAPIResponse{} //initialize an empty newsAPIResponse object/struct?
		_ = json.NewDecoder(resp.Body).Decode(&target)
		for i, article := range target.Articles {
			if !strings.Contains(article.Description, companyName) {
				target.Articles = append(target.Articles[:i], target.Articles[i+1:]...)
			}
		}
		models.CurrentArticlesMap[companyName] = target.Articles
	}
}

func main() {
	// evaluateSentiment("Hello World")
	evaluateSentiment("The Go language is great. It's fantastic when it works the way you want")
	var connErr error
	session, connErr = mgo.Dial("mongodb://terry:terrypass@ds059284.mlab.com:59284/company-alerts") //TODO store this in env
	check(connErr)

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	router := httprouter.New()

	router.POST("/api/addalert", handleAddalert)
	router.POST("/api/authenticate", handleLogin)
	router.POST("/api/register", handleRegister)
	router.GET("/", handleIndex)
	//
	// doEvery(250000*time.Millisecond, refreshTechCrunchArticles)
	// doEvery(300000*time.Millisecond, alertLoop)
	http.ListenAndServe(":8080", router)

}
