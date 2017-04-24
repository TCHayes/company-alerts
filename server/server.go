package main

import (
	"net/http"
  "fmt"
  "encoding/json"
)

func main() {
  http.HandleFunc("/api/getlatest", getLatestHandler)
  http.ListenAndServe(":8080", nil)
}

type Article struct {
  Author string
  Title string
  Description string
  Url string
  UrlToImage string
  PublishedAt string
}

type NewsApiResponse struct {
  Status string
  Source string
  SortBy string
  Articles []Article
}

func getLatestHandler(w http.ResponseWriter, r *http.Request){
  resp, _ := http.Get("https://newsapi.org/v1/articles?source=techcrunch&sortBy=latest&apiKey=bd7079b419d3439ca765e70919837e9d")
  defer resp.Body.Close()
  target := NewsApiResponse{}
  _ = json.NewDecoder(resp.Body).Decode(&target)
  fmt.Println(target.Articles)
}
