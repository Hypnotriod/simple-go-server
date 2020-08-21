package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"./simplsrvr"
)

func main() {
	var server simplsrvr.SimpleServer
	go lauchServer(&server)
	handleSeverInput(&server)
	fmt.Println("Bye, have a beautiful time!")
}

func handleSeverInput(server *simplsrvr.SimpleServer) {
	reader := bufio.NewReader(os.Stdin)
	for {
		if str, err := reader.ReadString('\n'); err == nil {
			msg := strings.Trim(str, "\n\r\t")
			if msg == "exit" {
				server.Stop()
				break
			}
			server.SendToAll(msg + "\r\n")
		}
	}
}

func lauchServer(server *simplsrvr.SimpleServer) {
	onServerEvent := func(event simplsrvr.SimpleServerEvent) {
		switch event {
		case simplsrvr.Started:
			fmt.Println("Server Started")
		case simplsrvr.Stopped:
			fmt.Println("Server Stopped")
		case simplsrvr.ConnAccepted:
			fmt.Println("Connection Accepted")
		case simplsrvr.ConnClosed:
			fmt.Println("Connection Closed")
		}
	}

	onServerError := func(err error) {
		fmt.Println("Server Error:", err)
	}

	onMessage := func(id int, msg string) {
		fmt.Printf("Message received form %d: %s\r\n", id, msg)
	}

	server.OnEvent(onServerEvent)
	server.OnError(onServerError)
	server.OnMessage(onMessage)
	server.Start("tcp", ":9876")
}
