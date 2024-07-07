package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/shivam-bhadani/distributed-cache/store"
)

type Server struct {
	Address  string
	Listener net.Listener
	store    map[string]*store.Store
	quitch   chan struct{}
	mu       sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
		store:   make(map[string]*store.Store),
		quitch:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("error starting sever %w", err)
	}
	fmt.Printf("Server is started at %v\n", s.Address)
	defer listener.Close()
	s.Listener = listener
	go s.acceptConnections()
	<-s.quitch
	return nil
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection:", err)
			continue
		}
		fmt.Printf("New client %v connected\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	conn.Write([]byte("LOGIN or SIGNUP\n"))

	authenticated := false
	var currentUser string

	for !authenticated {
		authType, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection %v Closed\n", conn.RemoteAddr())
			return
		}
		authType = strings.TrimSpace(strings.ToUpper(authType))
		switch authType {
		case "LOGIN":
			conn.Write([]byte("Enter username: "))
			username, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading username:", err)
				return
			}
			username = strings.TrimSpace(username)
			if s.isUserExist(username) {
				conn.Write([]byte("Enter password: "))
				password, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading password:", err)
					return
				}
				password = strings.TrimSpace(password)
				if s.store[username].IsCorrectPassword(password) {
					conn.Write([]byte("Authenticated successfully\n"))
					authenticated = true
					currentUser = username
				} else {
					conn.Write([]byte("Password is wrong\n"))
				}
			} else {
				conn.Write([]byte("Username does not exist\n"))
			}

		case "SIGNUP":
			conn.Write([]byte("Enter username: "))
			username, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading username:", err)
				return
			}
			username = strings.TrimSpace(username)
			if s.isUserExist(username) {
				conn.Write([]byte("This user already exists\n"))
			} else {
				conn.Write([]byte("Enter password: "))
				password, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading password:", err)
					return
				}
				password = strings.TrimSpace(password)
				s.addUser(username, password)
				conn.Write([]byte("User created successfully\n"))
				authenticated = true
				currentUser = username
			}

		default:
			conn.Write([]byte("Invalid command. Please enter LOGIN or SIGNUP\n"))
		}
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection %v Closed\n", conn.RemoteAddr())
			return
		}
		s.handleCommand(currentUser, strings.TrimSpace(message), conn)
	}
}

func (s *Server) isUserExist(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exist := s.store[username]
	return exist
}

func (s *Server) addUser(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[username] = store.NewStore(username, password)
}

func (s *Server) handleCommand(currentUser, command string, conn net.Conn) {
	parts := strings.Split(command, " ")
	if len(parts) < 2 {
		conn.Write([]byte("Invalid Command\n"))
		return
	}
	parts[0] = strings.ToUpper(parts[0])
	switch parts[0] {
	case "GET":
		if len(parts) != 2 {
			conn.Write([]byte("Invalid Command\n"))
			return
		}
		key := parts[1]
		value, found := s.store[currentUser].Cache.Get(key)
		if !found {
			conn.Write([]byte("Key is not found\n"))
			return
		}
		conn.Write([]byte(fmt.Sprintf("%s\n", value)))

	case "SET":
		if len(parts) != 3 {
			conn.Write([]byte("Invalid command\n"))
			return
		}
		key := parts[1]
		value := parts[2]
		err := s.store[currentUser].Cache.Set(key, value)
		if err != nil {
			conn.Write([]byte("Something went wrong\n"))
			return
		}
		conn.Write([]byte("OK\n"))

	case "HAS":
		if len(parts) != 2 {
			conn.Write([]byte("Invalid Command\n"))
			return
		}
		key := parts[1]
		found := s.store[currentUser].Cache.Has(key)
		if found {
			conn.Write([]byte("1\n"))
		} else {
			conn.Write([]byte("0\n"))
		}

	case "DELETE":
		if len(parts) != 2 {
			conn.Write([]byte("Invalid Command\n"))
			return
		}
		key := parts[1]
		err := s.store[currentUser].Cache.Delete(key)
		if err != nil {
			conn.Write([]byte("Something went wrong\n"))
			return
		}
		conn.Write([]byte("OK\n"))

	default:
		conn.Write([]byte("Invalid command\n"))
	}
}
