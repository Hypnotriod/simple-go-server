package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Hypnotriod/simple-go-server/simplsrvr"
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
			server.SendMessageToAll("Message from Server: " + msg + "\r\n")
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
		msg = fmt.Sprintf("Message form %d: %s\r\n", id, msg)
		fmt.Print(msg)
		server.SendMessageToAllExcept(msg, id)
	}

	server.OnEvent(onServerEvent)
	server.OnError(onServerError)
	server.OnMessage(onMessage)
	server.Start("tcp", ":9876")
}
