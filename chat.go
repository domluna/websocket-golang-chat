package main

import (
	"code.google.com/p/go.net/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
)

const (
	listenAddr = "localhost:4000" // server address
)

var (
	pwd, _         = os.Getwd()
	rootTemp       = template.Must(template.ParseFiles(pwd + "/ex.html"))
	JSON           = websocket.JSON            // codec for JSON
	Message        = websocket.Message         // codec for string, []byte
	active_clients = make(map[client_sock]int) // map containing clients
)

// Initialize handlers and websocket handlers
func init() {
	http.HandleFunc("/", rootHandler)
	http.Handle("/sock", websocket.Handler(SockServer))
}

// 
type client_sock struct {
	websocket *websocket.Conn
	client_ip string
}

func SockServer(ws *websocket.Conn) {
	var err error
	var client_msg string
	// use []byte if websocket binary type is blob or arraybuffer
	// var client_msg []byte

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := ws.Request().RemoteAddr
	log.Println("Client connected:", client)
	sock_cli := client_sock{ws, client}
	active_clients[sock_cli] = 0
	log.Println("Number of clients connected ...", len(active_clients))

	// for loop so the websocket stays open otherwise
	// it'll close after one Receieve and Send
	for {
		if err = Message.Receive(ws, &client_msg); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			// remove the ws client conn from our active clients
			delete(active_clients, sock_cli)
			log.Println("Number of clients still connected ...", len(active_clients))
			return
		}

		client_msg = sock_cli.client_ip + " Said: " + client_msg
		for cs, _ := range active_clients {
			if err = Message.Send(cs.websocket, client_msg); err != nil {
				// we could not send the message to a peer
				log.Println("Could not send message to ", cs.client_ip, err.Error())
			}
		}
	}
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	err := rootTemp.Execute(w, listenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
