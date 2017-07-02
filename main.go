package main

import (
	"html/template"
	"log"
	"net/http"
)

type shell struct {
	Dir string
}

var shellTmpl = template.Must(template.ParseFiles("index.html"))

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sh := shell{Dir: "/home/kybin"}
		err := shellTmpl.Execute(w, sh)
		if err != nil {
			log.Fatal(err)
		}
	})
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
