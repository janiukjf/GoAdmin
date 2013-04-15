package plate

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	tmpl *Template
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

	t.FuncMap["isNotNull"] = func(str string) bool {
		if strings.TrimSpace(str) != "" && len(strings.TrimSpace(str)) > 0 {
			return true
		}
		return false
	}

}

func (this *Server) Template(w http.ResponseWriter) (templ *Template, err error) {
	if w == nil {
		log.Printf("Template Error: %v", err.Error())
		return
	}
	templ = &Template{
		Writer: w,
		Bag:    make(map[string]interface{}),
	}

	return
}

func (t Template) SinglePage(file_path string) (err error) {

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}
	if len(file_path) != 0 {
		t.Template = dir + "/" + file_path
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

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = dir + "/" + t.Layout
	t.Template = dir + "/" + t.Template

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

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = dir + "/" + t.Layout

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
		templ.ParseFiles(dir + "/" + filename)
	}
	err = templ.Execute(t.Writer, t.Bag)
	if err != nil {
		log.Println(err)
	}

	return
}
func SetTemplate(t Template) {
	tmpl = &t
}

func GetTemplate() (*Template, error) {
	if tmpl != nil {
		return tmpl, nil
	}

	return nil, errors.New("No template defined")
}
