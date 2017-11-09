package handlers

import (
	"github.com/spastorelli/dojohub/template"
	"net/http"
)

func ExampleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Render(w, "chat.tmpl", r.Host)
}
