package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type shell struct {
	Dir string
}

var shellTmpl = template.Must(template.ParseFiles("index.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sh := shell{Dir: "/home/kybin"}
		err := shellTmpl.Execute(w, sh)
		if err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				// if session is closed, it will raise 'going away' error.
				// so this function will return.
				log.Print(err)
				return
			}
			fmt.Println(string(p))
			err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
			if err != nil {
				log.Print(err)
				return
			}
		}
	})
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
