package simplsrvr

import (
	"bufio"
	"io"
	"net"
	"strings"
	"sync"
)

// SimpleServerEvent simple server event
type SimpleServerEvent int

const (
	// Started server started event
	Started SimpleServerEvent = iota
	// Stopped server stopped event
	Stopped
	// ConnAccepted new connection accepted
	ConnAccepted
	// ConnClosed connection closed
	ConnClosed
)

// SimpleServer simple server wrapper
type SimpleServer struct {
	sync.RWMutex
	errorHandler func(error)
	eventHandler func(SimpleServerEvent)
	msgHandler   func(int, string)
	connections  map[int]net.Conn
	connCounter  int
	listener     net.Listener
	running      bool
}

func (s *SimpleServer) onError(err error) {
	if s.errorHandler != nil && err != io.EOF {
		s.errorHandler(err)
	}
}

func (s *SimpleServer) onEvent(event SimpleServerEvent) {
	if s.eventHandler != nil {
		s.eventHandler(event)
	}
}

func (s *SimpleServer) onMessage(id int, msg string) {
	if s.msgHandler != nil {
		s.msgHandler(id, msg)
	}
}

func (s *SimpleServer) registerConnection(conn net.Conn) int {
	s.Lock()
	id := s.connCounter
	s.connections[id] = conn
	s.connCounter++
	s.Unlock()
	return id
}

func (s *SimpleServer) removeConnection(id int) {
	s.Lock()
	delete(s.connections, id)
	s.Unlock()
}

func (s *SimpleServer) isRunning() bool {
	s.RLock()
	defer s.RUnlock()
	return s.running
}

// OnError error callback setter
func (s *SimpleServer) OnError(callback func(error)) {
	s.errorHandler = callback
}

// OnEvent event callback setter
func (s *SimpleServer) OnEvent(callback func(SimpleServerEvent)) {
	s.eventHandler = callback
}

// OnMessage message callback setter
func (s *SimpleServer) OnMessage(callback func(int, string)) {
	s.msgHandler = callback
}

// SendToAll sends message to all connections
func (s *SimpleServer) SendToAll(msg string) {
	s.RLock()
	for _, conn := range s.connections {
		conn.Write([]byte(msg))
	}
	s.RUnlock()
}

// Stop stops the server
func (s *SimpleServer) Stop() {
	s.Lock()
	s.running = false
	for _, conn := range s.connections {
		conn.Close()
	}
	s.connCounter = 0
	s.connections = make(map[int]net.Conn)
	s.listener.Close()
	s.onEvent(Stopped)
	s.Unlock()
}

// Start starts the server
func (s *SimpleServer) Start(network string, address string) {
	listener, err := net.Listen(network, address)
	if err != nil {
		s.onError(err)
		return
	}
	s.onEvent(Started)
	s.running = true
	s.listener = listener
	s.connections = make(map[int]net.Conn)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if !s.isRunning() {
				break
			} else {
				s.errorHandler(err)
				conn.Close()
				continue
			}
		}
		go handleConnection(s, conn)
	}
}

func handleConnection(s *SimpleServer, conn net.Conn) {
	id := s.registerConnection(conn)
	s.eventHandler(ConnAccepted)
	defer conn.Close()
	defer s.eventHandler(ConnClosed)

	reader := bufio.NewReader(conn)

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			if s.isRunning() {
				s.onError(err)
				s.removeConnection(id)
			}
			break
		}

		s.onMessage(id, strings.Trim(str, "\n\r\t"))
	}
}
