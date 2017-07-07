package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type shell struct {
	mu  sync.Mutex
	Dir string
}

func (s *shell) changeDir(d string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, err := os.Stat(d)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%v is not a directory", d)
	}
	s.Dir = d
	fmt.Println(s)
	return nil
}

func (s *shell) runCommand(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cmd == "" {
		return []byte{}, nil
	}
	cmds := strings.Split(strings.TrimSpace(cmd), " ")

	name := cmds[0]
	args := cmds[1:]

	c := exec.Command(name, args...)
	c.Dir = s.Dir
	return c.CombinedOutput()
}

var shellTmpl = template.Must(template.ParseFiles("index.html"))

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	dir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	sh := shell{Dir: dir}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = shellTmpl.Execute(w, sh)
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
			msg := strings.Split(string(p), ":")
			if len(msg) != 2 {
				log.Printf("invalid message: %v", msg)
				return
			}
			typ := msg[0]
			body := msg[1]

			result := []byte{}
			switch typ {
			case "cmd":
				out, err := sh.runCommand(body)
				if err != nil {
					result = []byte("error:" + string(out) + err.Error())
				} else {
					result = append([]byte("ok:"), out...)
				}
			case "dir":
				err := sh.changeDir(body)
				if err != nil {
					result = []byte("error:" + err.Error())
				} else {
					result = []byte("ok:")
				}
			default:
				log.Printf("unknown message type: %v", typ)
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, result)
			if err != nil {
				log.Print(err)
				return
			}
		}
	})
	err = http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
