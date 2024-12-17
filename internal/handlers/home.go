package handlers

import (
	"html/template"
	"net/http"
)

func BrokerLanding() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		data := struct {
			FirstName string
		}{
			FirstName: user.FirstName,
		}
		println(data.FirstName)
		tmpl := template.Must(template.ParseFiles("internal/templates/broker.html"))
		tmpl.Execute(w, data)
	})
}

func AdminLanding() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		data := struct {
			FirstName string
		}{
			FirstName: user.FirstName,
		}
		println(data.FirstName)
		tmpl := template.Must(template.ParseFiles("internal/templates/admin.html"))
		tmpl.Execute(w, data)
	})
}

func SignUpSuccessPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("internal/templates/signup_success.html"))
		tmpl.Execute(w, nil)
	}
}
