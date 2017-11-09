package template

import (
	"html/template"
	"net/http"
	"path"
)

var tmplDir string = "template"
var baseTmplFile string = path.Join(tmplDir, "base.tmpl")
var tmplFiles []string = []string{
	"home.tmpl",
	"chat.tmpl",
}
var tmpls map[string]*template.Template = make(map[string]*template.Template)

func init() {
	for _, file := range tmplFiles {
		tmpls[file] = template.Must(template.ParseFiles(path.Join(tmplDir, file), baseTmplFile))
	}
}

func Render(w http.ResponseWriter, tmpl string, context interface{}) {
	err := tmpls[tmpl].Execute(w, context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
