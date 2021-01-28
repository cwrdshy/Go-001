package main

import (
	"fmt"
	"net"
)

// 定义客户端结构体
type Client struct {
	socket net.Conn
	data   chan []byte
}

// 池连接，能够将收到的任何消息广播到连接池中的所有连接
// 当一个客户端连接或断开连接时，能够通知现有的客户端
type ClientPool struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func (p *ClientPool) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			p.unregister <- client
			_ = client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("received message: " + string(message))
			p.broadcast <- message
		}
	}
}

func (p *ClientPool) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			_, _ = client.socket.Write(message)
		}
	}
}

func (p *ClientPool) start() {
	for {
		select {
		case connection := <-p.register:
			p.clients[connection] = true
			fmt.Println("Added new connection!")
		case connection := <-p.unregister:
			if _, ok := p.clients[connection]; ok {
				close(connection.data)
				delete(p.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		case message := <-p.broadcast:
			for connection := range p.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(p.clients, connection)
				}
			}
		}
	}
}

func startServer() {
	fmt.Println("starting server...")
	listener, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		fmt.Println(err)
	}
	pool := ClientPool{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go pool.start()
	for {
		connection, err := listener.Accept()
		if err != nil {
			print(err)
		}
		client := &Client{socket: connection, data: make(chan []byte)}
		pool.register <- client
		go pool.receive(client)
		go pool.send(client)
	}
}

func main() {
	startServer()
}
