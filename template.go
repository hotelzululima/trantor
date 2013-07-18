package main

import (
	"html/template"
	"net/http"
)

type Status struct {
	Search string
	User   string
	Notif  []Notification
	Home   bool
	About  bool
	News   bool
	Upload bool
	Stats  bool
	Help   bool
}

func GetStatus(w http.ResponseWriter, r *http.Request) Status {
	var s Status
	sess := GetSession(r)
	sess.Save(w, r)
	s.User = sess.User
	s.Notif = sess.Notif
	return s
}

var templates = template.Must(template.ParseFiles(TEMPLATE_PATH+"header.html",
	TEMPLATE_PATH+"footer.html",
	TEMPLATE_PATH+"404.html",
	TEMPLATE_PATH+"index.html",
	TEMPLATE_PATH+"about.html",
	TEMPLATE_PATH+"news.html",
	TEMPLATE_PATH+"edit_news.html",
	TEMPLATE_PATH+"book.html",
	TEMPLATE_PATH+"search.html",
	TEMPLATE_PATH+"upload.html",
	TEMPLATE_PATH+"new.html",
	TEMPLATE_PATH+"read.html",
	TEMPLATE_PATH+"edit.html",
	TEMPLATE_PATH+"settings.html",
	TEMPLATE_PATH+"stats.html",
	TEMPLATE_PATH+"help.html",
))

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
