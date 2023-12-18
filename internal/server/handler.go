package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"wantsome.ro/messagingapp/pkg/models"
)

var (
	m               sync.Mutex
	userConnections = make(map[*websocket.Conn]string)
	broadcast       = make(chan models.Message)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world from my server!")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("got error upgrading connection %s\n", err)
		return
	}
	defer conn.Close()

	m.Lock()
	userConnections[conn] = ""
	m.Unlock()
	fmt.Printf("connected client!")

	for {
		log.Printf("starting connection loop: %v", userConnections)
		var msg models.Message = models.Message{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("got error reading message %s\n", err)
			m.Lock()
			delete(userConnections, conn)
			m.Unlock()
			return
		}
		if msg.Type == models.MessageTypeConnect {
			fmt.Printf("Server got connect request from client: %s", msg.UserName)
			m.Lock()
			userConnections[conn] = msg.UserName
			m.Unlock()
			continue
		} else if msg.Type == models.MessageTypeDisconnect {
			fmt.Printf("Server got disconnect request from client: %s", msg.UserName)
			m.Lock()
			delete(userConnections, conn)
			m.Unlock()
			continue
		} else if msg.Type == models.MessageTypeListUsers {
			fmt.Printf("Server got list users request from client: %s", msg.UserName)
			m.Lock()
			usersList := []string{}
			for _, username := range userConnections {
				usersList = append(usersList, username)
			}
			m.Unlock()
			reply := models.Message{
				UserName: msg.UserName,
				Target:   msg.UserName,
				Message:  fmt.Sprintf("User list: %v", usersList),
				Type:     models.MessageTypeListUsers,
			}
			err := forwardMessageToTarget(reply)
			if err != nil {
				log.Printf("Error sending reply with users: %q", err)
			}
			continue
		}
		log.Printf("Server got following msg: %v", msg)
		if msg.Target != "" {
			err := forwardMessageToTarget(msg)
			if err != nil {
				log.Printf("User target does not exist: %q", msg.Target)
				//fmt.Fprintf(w, "User target does not exist: %q", msg.Target)
				reply := models.Message{
					UserName: msg.UserName,
					Target:   msg.Target,
					Message:  fmt.Sprintf("Error sending wisper to: %q : %s", msg.Target, err),
				}
				err := conn.WriteJSON(reply)
				if err != nil {
					fmt.Printf("Error sending error reply: %s", err)
				}
			}
			continue
		}
		broadcast <- msg
	}
}

func handleMsg() {
	for {
		msg := <-broadcast

		m.Lock()
		for client, username := range userConnections {
			if username != msg.UserName {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Printf("got error broadcating message to client %s", err)
					client.Close()
					delete(userConnections, client)
				}
			}
		}
		m.Unlock()
	}
}

func forwardMessageToTarget(msg models.Message) error {
	userfound := false
	m.Lock()
	for client, username := range userConnections {
		if username == msg.Target {
			userfound = true
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Printf("got error sending wisper to client %s", err)
				return err
			}
		}
	}
	m.Unlock()
	if !userfound {
		return fmt.Errorf("user not found: %q", msg.Target)
	}
	return nil
}
