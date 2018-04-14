package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"chat/db"
	"chat/protocol"
)

const (
	Port uint16 = 4242
	Network = "tcp"
)

// Simple Server struct
type Server struct {
	User *db.User
	Listener *net.TCPListener
	Sessions []*Session
}

// Setup listener for server
func setupServer(address string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr(Network, address)
	if err != nil {
		return nil, err
	}
	return net.ListenTCP(Network, tcpAddr)
}

// Handle receiving messages from a TCPConn
func handleConnection(s *Server, conn *net.TCPConn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		log.Panicf("handleConnection: %s", err.Error())
	}

	if msg.StartProto != nil {
		log.Printf("They want to start a new protocol of type %s", msg.StartProto)
	}
	if s.User.IP != msg.DestIP {
		log.Panicln("User received a message that was not meant for them")
	}

	sess := s.CreateOrGetSession(msg)
	dec, err := sess.Proto.Decrypt(msg.Text)
	switch errorType := err.(type) {
	case protocol.OTRHandshakeStep:
		// If it's part of the OTR handshake, send each part of the message back directly to the source,
		// and immediately return
		for _, stepMessage := range dec {
			s.Send(msg.SourceIP, stepMessage)
		}
		return
	default:
		log.Panicf("ReceiveMessage: %s, Error Type: %s", err.Error(), errorType)
	}
	// Print the decoded MAC
	fmt.Printf("%s: %s\n", msg.SourceMAC, dec[0])
}

// Function that continuously polls for new messages being sent to the server
func receive(s *Server) {
	for {
		if conn, err := (*(*s).Listener).AcceptTCP(); err == nil {
			go handleConnection(s, conn)
		}
	}
}

func initDialer(address string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr(Network, address)
	if err != nil {
		return nil, err
	}
	return net.DialTCP(Network, nil, tcpAddr)
}

// Start up server
func (s *Server) Start(username string, mac string, ip string) error {
	var err error
	log.Println("Launching Server...")
	(*s).User = &db.User{username, mac, ip}
	ipAddr := fmt.Sprintf("%s:%d", ip, Port)
	if (*s).Listener, err = setupServer(ipAddr); err != nil {
		return err
	}
	go receive(s)
	log.Printf("Listening on: '%s:%d'", ip, Port)
	return nil
}

// End server connection
func (s *Server) Shutdown() error {
	log.Println("Shutting Down Server...")
	return (*s).Listener.Close()
}

// Private-helper method that sends a formatted message object with the server
func (s *Server) sendMessage(msg *Message) error {
	dialer, err := initDialer(fmt.Sprintf("%s:%d", msg.SourceIP, Port))
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(dialer)
	if err := encoder.Encode(msg); err != nil {
		return err
	}
	return nil
}

// Send a message to another Server
func (s *Server) Send(destIp string, message []byte) error  {
	return s.sendMessage(NewMessage(s.User, destIp, message))
}

// Returns a session based on the message received
func (s *Server) CreateOrGetSession(msg Message) (*Session) {
	for _, sess := range s.Sessions {
		if (*sess).ConverseWith(msg.SourceIP) {
			return sess
		}
	}
	// TODO: If we have to create a new session, do we update the protocol?
	friend := new(User)
	friend.IP = msg.SourceIP
	friend.MAC = msg.SourceMAC
	// The From field of a session is always the server's user
	sess := NewSession(s.User, friend, msg.StartProto)
	s.Sessions = append(s.Sessions, sess)
	return sess
}

// High-level function when you want to enable a session with a protocol with another user
func (s *Server) StartSession(destIp string, proto protocol.Protocol) (error) {
	// When you first start a session, you don't know the SourceMAC, so just don't create a session for now, create it
	// when the user gets a response
	firstMessage, err := proto.NewSession()
	if err != nil {
		log.Panicf("StartSession: Error starting new session: %s", err)
		return err
	}

	if len(firstMessage) > 0 {
		return err
	}
	msg := NewMessage(s.User, destIp, []byte(firstMessage))
	msg.StartProto = proto
	return s.sendMessage(msg)
}
