# Lots going on here (i may have gone overboard, it's late and not enough sleep is killer)

## setup
  run go install before running go build

    
# YOUR TASK
  your mission should you choose to take it, is to ...
  1. complete what we talked about last night
  2. understand what is happening (by reading the rest of this readme) and hunt down any bugs that you might find
  3. complete the unfinished function called refreshTheNextWebArticles() in a same fashion as refreshTechCrunchArticles() however change the database source to be the next web (don't worry about how it will overwrite from techcrunch, lets just focus on one source. however i'd like you to complete this function to better understand what is happening)

# 1. DAO package
  * so I created a dao (data access object) package that you can use to communicate with mongodb
  * it works by first taking in a collection, then a query / new object to insert, and a resulting taget object just like when we decode objects
  ```go
              // the collection    // bson.M is just a way to initialize abson object 
                                                        // target is the struct we are stuffing the result into
  func ReadOne(collection *mgo.Collection, query bson.M, target interface{}) bool {
    err := collection.Find(query).One(target)
    return check(err)
  }
  ```
# 2. Model Package
  * in here we got our structs that we will be using throughout our app, mostly just definitions but from outside of the models package you 
    can access a struct and their functions through models.Structname
  * you can also access the currentarticlesmap and other things through models.CurrentArticlesMap (it is capitilized and hence you can access publicly)
  * models.CurrentUsers maps userIDs to companies
  * models.CurrentArticlesMap maps companyNames to Articles

# 3. How the routing works
  lets look at handel login
```go
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
```

this bit of code says that our collection "user" is in the company alerts database don't worry about mongo connections yet
```
	c := session.DB("company-alerts").C("user")
```

here we read in the request using the bsonify function which is in server.go

```
  request := bsonify(r)
	doc := models.User{}
	readSuccess := dao.ReadOne(c, request, &doc)
```

finally we create and send a response back to the client using createResponse which sends a bson document back to the user

```
if doc.Username == request["username"] && doc.Password == request["password"] {
  createResponse(doc, &w)
  return
}

```

  4. time triggered fetch of new articles and checek for alert

    the t time.Time parameter is something somewhat complicated all you need to know is that it is implicitly used to call the function every x seconds

    see alertloop and refreshTechCrunchArticles functions towards bottom of file, work through line by line and understnad the algorithm that is happening
    ###track it step by step and write it down as you go, we'll talk about it for sure :)

