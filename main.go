package main

import (
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"

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

			cmdstr := string(p)
			if cmdstr == "" {
				continue
			}
			cmds := strings.Split(strings.TrimSpace(cmdstr), " ")

			cmd := cmds[0]
			args := cmds[1:]

			result := []byte{}
			c := exec.Command(cmd, args...)
			out, err := c.CombinedOutput()
			if err != nil {
				result = []byte("error: " + err.Error())
			} else {
				result = out
			}

			err = conn.WriteMessage(websocket.TextMessage, result)
			if err != nil {
				log.Print(err)
				return
			}
		}
	})
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
