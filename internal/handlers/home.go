package handlers

import (
	"html/template"
	"net/http"
)

func BrokerLanding() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("internal/templates/broker.html"))
		tmpl.Execute(w, nil)
	})
}

func AdminLanding() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("internal/templates/admin.html"))
		tmpl.Execute(w, nil)
	})
}
