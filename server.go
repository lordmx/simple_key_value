package main

import (
	"net"
	"bufio"
	"time"
	"log"
)

type Client struct {
	ID int
	Server *Server
	conn net.Conn
}

type Server struct {
	clients map[int]*Client
	clientsCount int
	joins chan net.Conn
	income chan *message
	onConnectedCallback func(client *Client)
	onDisconnectedCallback func(client *Client, err error)
	onMessageCallback func(client *Client, data []byte)
}

type message struct {
	client *Client
	data []byte
}

func NewServer() *Server {
	return &Server{
		clients: make(map[int]*Client),
		joins: make(chan net.Conn),
		income: make(chan *message),
	}
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) read() {
	reader := bufio.NewReader(client.conn)
	server := client.Server

	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			server.onDisconnectedCallback(client, err)
			delete(server.clients, client.ID)
			server.clientsCount--

			break
		}

		server.income <- &message{client, data}
	}

	client.Close()
}

func (server *Server) Listen(host string) error {
	go func() {
		for {
			select {
			case conn := <- server.joins:
				server.clientConnected(conn)

			case message := <- server.income:
				if client, ok := server.clients[message.client.ID]; ok {
					server.onMessageCallback(client, message.data)
				}
			}
		}
	}()

	listener, err := net.Listen("tcp", host)

	if err != nil {
		return err
	}

	defer listener.Close()

	log.Printf("[%v] Begin listen on %s", time.Now(), host)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("[%v] %s", time.Now(), err)
			continue
		}

		server.joins <- conn
	}
}

func (server *Server) OnConnect(callback func (client *Client)) {
	server.onConnectedCallback = callback
}

func (server *Server) OnDisconnect(callback func (client *Client, err error)) {
	server.onDisconnectedCallback = callback
}

func (server *Server) OnMessage(callback func (client *Client, data []byte)) {
	server.onMessageCallback = callback
}

func (server *Server) sendToClient(client *Client, data []byte) {
	_, err := client.conn.Write(data)

	if err != nil {
		log.Printf("[%v] %s", time.Now(), err)
	}
}

func (server *Server) clientConnected(conn net.Conn) {
	server.clientsCount++

	client := &Client{
		ID: server.clientsCount,
		conn: conn,
		Server: server,
	}

	server.clients[client.ID] = client
	server.onConnectedCallback(client)

	go client.read()
}