package simplsrvr

import (
	"io"
	"net"
	"strings"
	"sync"
)

// SimpleServerEvent simple server event
type SimpleServerEvent int

// BufferSizeDefault default size of buffer for conn.Read()
const BufferSizeDefault uint = 1024

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
	connections  map[int]*net.Conn
	connCounter  int
	listener     net.Listener
	running      bool
	bufferSize   uint
}

// OnError error callback setter
func (s *SimpleServer) OnError(callback func(error)) {
	s.Lock()
	s.errorHandler = callback
	s.Unlock()
}

// OnEvent event callback setter
func (s *SimpleServer) OnEvent(callback func(SimpleServerEvent)) {
	s.Lock()
	s.eventHandler = callback
	s.Unlock()
}

// OnMessage message callback setter
func (s *SimpleServer) OnMessage(callback func(int, string)) {
	s.Lock()
	s.msgHandler = callback
	s.Unlock()
}

// SendMessageToAll sends message to all connections
func (s *SimpleServer) SendMessageToAll(msg string) {
	s.RLock()
	for _, conn := range s.connections {
		(*conn).Write([]byte(msg))
	}
	s.RUnlock()
}

// SendMessageToAllExcept sends message to all connections except id provided
func (s *SimpleServer) SendMessageToAllExcept(msg string, id int) {
	s.RLock()
	for i, conn := range s.connections {
		if i != id {
			(*conn).Write([]byte(msg))
		}
	}
	s.RUnlock()
}

// GetBufferSize returns the buffer size for conn.Read()
func (s *SimpleServer) GetBufferSize() uint {
	s.RLock()
	defer s.RUnlock()
	return s.bufferSize
}

// SetBufferSize sets the buffer size for conn.Read()
func (s *SimpleServer) SetBufferSize(value uint) {
	s.Lock()
	s.bufferSize = value
	s.Unlock()
}

// Stop stops the server
func (s *SimpleServer) Stop() {
	s.setRunning(false)
	s.closeAllConnections()

	s.RLock()
	s.listener.Close()
	s.RUnlock()

	s.sendEvent(Stopped)
}

// Start starts the server
func (s *SimpleServer) Start(network string, address string) {
	listener, err := net.Listen(network, address)
	if err != nil {
		s.sendError(err)
		s.sendEvent(ConnClosed)
		return
	}
	s.sendEvent(Started)
	s.running = true
	s.listener = listener
	if s.bufferSize == 0 {
		s.bufferSize = BufferSizeDefault
	}
	s.connections = make(map[int]*net.Conn)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if !s.isRunning() {
				break
			} else {
				s.sendError(err)
				conn.Close()
				continue
			}
		}
		go handleConnection(s, &conn)
	}
}

func handleConnection(s *SimpleServer, conn *net.Conn) {
	id := s.registerConnection(conn)
	s.sendEvent(ConnAccepted)
	buff := make([]byte, s.GetBufferSize())

	for {
		n, err := (*conn).Read(buff)
		if err != nil {
			if s.isRunning() {
				s.sendError(err)
				s.unregisterConnection(id)
				(*conn).Close()
			}
			s.sendEvent(ConnClosed)
			break
		}

		msg := strings.Trim(string(buff[0:n]), "\n\r\t")
		if len(msg) > 0 {
			s.sendMessage(id, msg)
		}
	}
}

func (s *SimpleServer) sendError(err error) {
	s.RLock()
	if s.errorHandler != nil && err != io.EOF {
		s.errorHandler(err)
	}
	s.RUnlock()
}

func (s *SimpleServer) sendEvent(event SimpleServerEvent) {
	s.RLock()
	if s.eventHandler != nil {
		s.eventHandler(event)
	}
	s.RUnlock()
}

func (s *SimpleServer) sendMessage(id int, msg string) {
	s.RLock()
	if s.msgHandler != nil {
		s.msgHandler(id, msg)
	}
	s.RUnlock()
}

func (s *SimpleServer) registerConnection(conn *net.Conn) int {
	s.Lock()
	id := s.connCounter
	s.connections[id] = conn
	s.connCounter++
	s.Unlock()
	return id
}

func (s *SimpleServer) unregisterConnection(id int) {
	s.Lock()
	delete(s.connections, id)
	s.Unlock()
}

func (s *SimpleServer) isRunning() bool {
	s.RLock()
	defer s.RUnlock()
	return s.running
}

func (s *SimpleServer) setRunning(value bool) {
	s.Lock()
	s.running = value
	s.Unlock()
}

func (s *SimpleServer) closeAllConnections() {
	s.RLock()
	for _, conn := range s.connections {
		(*conn).Close()
	}
	s.RUnlock()

	s.Lock()
	s.connCounter = 0
	s.connections = make(map[int]*net.Conn)
	s.Unlock()
}
