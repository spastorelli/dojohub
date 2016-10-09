package handlers

import (
  "net/http"
  "html/template"
)

var templatesLoc = []string{
  "template/index.tmpl",
  "template/chat.tmpl",
  "template/content.tmpl",
  "template/nav_drawer.tmpl",
  "template/nav.tmpl",
}
var templates = template.Must(template.ParseFiles(templatesLoc...))

func renderTemplate(w http.ResponseWriter, tmpl string, context interface{}) {
    err := templates.ExecuteTemplate(w, tmpl+".tmpl", context)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
