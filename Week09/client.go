package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Client struct {
	socket net.Conn
	data   chan []byte
}

//read message from conn
func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			_ = client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("client received message: " + string(message))
		}
	}
}

func startClientMode() {
	fmt.Println("Starting client...")
	connection, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		fmt.Println(err)
	}
	client := &Client{socket: connection}
	go client.receive()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		_, _ = connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
}

func main() {
	startClientMode()
}
