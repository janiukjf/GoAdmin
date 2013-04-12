package plate

import (
	"../globals"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

type Template struct {
	Layout   string
	Template string
	Bag      map[string]interface{}
	Writer   http.ResponseWriter
	FuncMap  template.FuncMap
}

/* Templating |-- Using html/template library built into golang http://golang.org/pkg/html/template/ --|
   ------------------------------ */

func (t *Template) SetGlobalValues() {
	// Set Bag values
	// example
	t.Bag["CurrentYear"] = time.Now().Year()

	// Set FuncMap Values
	// example:
	if t.FuncMap == nil {
		t.FuncMap = template.FuncMap{}
	}

	t.FuncMap["isLoggedIn"] = func() bool {
		return true
	}

	t.FuncMap["isNotNull"] = func(str string) bool {
		if strings.TrimSpace(str) != "" && len(strings.TrimSpace(str)) > 0 {
			return true
		}
		return false
	}

}

func (this *Server) Template(w http.ResponseWriter) (templ Template, err error) {
	if w == nil {
		log.Printf("Template Error: %v", err.Error())
		return
	}
	templ.Writer = w
	templ.Bag = make(map[string]interface{})
	return
}

func (t Template) SinglePage(file_path string) (err error) {
	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}
	if len(file_path) != 0 {
		t.Template = *globals.Filepath + file_path
	}

	t.SetGlobalValues()

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Template
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	tmpl, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Template)
	if err != nil {
		log.Println(err)
		return err
	}
	err = tmpl.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayTemplate() (err error) {
	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = *globals.Filepath + t.Layout
	t.Template = *globals.Filepath + t.Template

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	t.SetGlobalValues()

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Layout
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	templ, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Layout, t.Template)
	if err != nil {
		log.Println(err)
		return err
	}

	err = templ.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayMultiple(templates []string) (err error) {
	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = *globals.Filepath + t.Layout

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	t.SetGlobalValues()

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Layout
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	templ, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Layout)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, filename := range templates {
		templ.ParseFiles(*globals.Filepath + filename)
	}
	err = templ.Execute(t.Writer, t.Bag)
	if err != nil {
		log.Println(err)
	}

	return
}
