package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

var clients = make(map[(*websocket.Conn)]bool)
var clientMap = make(map[(*websocket.Conn)]string)
var broadcast = make(chan Message)

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws/", handleConnections)
	http.HandleFunc("/connected-clients", func(w http.ResponseWriter, r *http.Request) {
		connectedClients := getConnectedClients()
		fmt.Fprintf(w, "Connected clients: %v", connectedClients)
	})

	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Chat Room!")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Sender username is passed after /ws/ in the URL
	sender := r.URL.Path[len("/ws/"):]
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	clientMap[conn] = sender

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}

		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			if msg.Recipient != "" && msg.Recipient == clientMap[client] {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func getConnectedClients() []string {
	var connectedClients []string
	for _, client := range clientMap {
		connectedClients = append(connectedClients, client)
	}
	return connectedClients
}
