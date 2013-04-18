package users

import (
	"../../helpers/plate"
	"log"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	server := plate.NewServer()

	tmpl, err := plate.GetTemplate()
	if err != nil {
		tmpl, err = server.Template(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tmpl.ParseFile("templates/users/index.html", false)

	err = tmpl.Display(w)
	if err != nil {
		log.Println(err)
	}
}
