package client

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"wantsome.ro/messagingapp/pkg/models"
)

func RunClient() {
	config, err := LoadConfig("config.client.toml")
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	url := fmt.Sprintf("ws://%s:%d%s", config.Client.Server, config.Client.Port, config.Client.WsEndpoint)

	//url := "ws://localhost:8080/ws"
	randId := rand.Intn(10)
	username := fmt.Sprintf("Client_%d", randId)
	log.Printf("Username will be: %s", username)
	//message := models.Message{Message: fmt.Sprintf("Hello world from my client %d !", randId), UserName: fmt.Sprintf("Client %d", randId)}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("error dialing %s\n", err)
	}
	defer c.Close()

	err = c.WriteJSON(models.Message{
		Type:     models.MessageTypeConnect,
		UserName: username,
	})
	if err != nil {
		log.Printf("Failed to send connection info to server: %q", err)
	}

	done := make(chan bool)
	// reading server messages
	go func() {
		//defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("error reading: %s\n", err)
				return
			}
			fmt.Printf("Got message: %s\n", message)
		}
	}()

	// writing messages to server
	go func() {
		//for {
		//	err := c.WriteJSON(message)
		//	if err != nil {
		//		log.Printf("error writing %s\n", err)
		//		return
		//	}
		//	time.Sleep(3 * time.Second)
		//}
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			// TODO : see why we get error on quit
			if strings.HasPrefix(text, "/quit") {
				err := c.WriteJSON(models.Message{
					Type:     models.MessageTypeDisconnect,
					UserName: username,
				})
				if err != nil {
					log.Printf("Failed to send disconnect info to server: %q", err)
				}
				done <- true
				break
			}
			if strings.HasPrefix(text, "/list_users") {
				err := c.WriteJSON(models.Message{
					Type:     models.MessageTypeListUsers,
					UserName: username,
				})
				if err != nil {
					log.Printf("Failed to send list user info to server: %q", err)
				}
				continue
			}
			target := ""
			if strings.HasPrefix(text, "/w") {
				// TODO, use regex maybe
				_, err := fmt.Sscanf(text, "/w %s %s", &target, &text)
				if err != nil {
					log.Printf("Failed parsing wisper command: %s\n", err)
				}
			}
			//fmt.Fprintln(conn, text)
			message := models.Message{Message: text, UserName: username, Target: target}
			err := c.WriteJSON(message)
			if err != nil {
				log.Printf("error writing %s\n", err)
				return
			}

		}
	}()

	<-done
}
