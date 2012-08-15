package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

const (
	IP         = "127.0.0.1"
	DB_NAME    = "trantor"
	BOOKS_COLL = "books"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	loadTemplate(w, "about", nil)
}

func bookHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var book Book
		coll.Find(bson.M{"title": r.URL.Path[len("/book/"):]}).One(&book)
		loadTemplate(w, "book", book)
	}
}


func fileHandler(path string) {
	h := http.FileServer(http.Dir(path[1:]))
	http.Handle(path, http.StripPrefix(path, h))
}

func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(BOOKS_COLL)
	num, _ := coll.Count()

	http.HandleFunc("/book/", bookHandler(coll))
	http.HandleFunc("/search/", searchHandler(coll))
	http.HandleFunc("/upload/", uploadHandler(coll))
	http.HandleFunc("/about/", aboutHandler)
	fileHandler("/img/")
	fileHandler("/cover/")
	fileHandler("/books/")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { loadTemplate(w, "index", num) })
	http.ListenAndServe(":8080", nil)
}
