package smb

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	address string
	guid    [16]byte
}

const (
	CommandNegotiate      uint16 = 0x0000
	CommandSessionSetup   uint16 = 0x0001
	CommandLogoff         uint16 = 0x0002
	CommandTreeConnect    uint16 = 0x0003
	CommandTreeDisconnect uint16 = 0x0004
	CommandCreate         uint16 = 0x0005
	CommandClose          uint16 = 0x0006
	CommandFlush          uint16 = 0x0007
	CommandRead           uint16 = 0x0008
	CommandWrite          uint16 = 0x0009
	CommandLock           uint16 = 0x000A
	CommandIoctl          uint16 = 0x000B
	CommandCancel         uint16 = 0x000C
	CommandEcho           uint16 = 0x000D
	CommandQueryDirectory uint16 = 0x000E
	CommandChangeNotify   uint16 = 0x000F
	CommandQueryInfo      uint16 = 0x0010
	CommandSetInfo        uint16 = 0x0011
	CommandOplockBreak    uint16 = 0x0012
)

func NewServer(address string) *Server {
	s := &Server{address: address}
	rand.Read(s.guid[:])
	return s
}

func (s *Server) ListenAndServe() error {

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Println("[ListenAndServe] Listening on " + s.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[ListenAndServe] Failed to accept connection: %v", err)
			continue
		}
		log.Println("[ListenAndServe] Accepted connection from ", conn.RemoteAddr())
		go s.handleConnection(conn)

	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		msg, err := readSMB2Message(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("[handleConnection] client %s disconnected", conn.RemoteAddr())
			} else {
				log.Printf("[handleConnection] client %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		if err := s.handleMessage(conn, msg); err != nil {
			log.Printf("[handleConnection] failed to handle message for client %s : %v", conn.RemoteAddr(), err)
			return
		}
	}

}

func (s *Server) handleMessage(conn net.Conn, msg []byte) error {
	header, err := parseHeader(msg)
	if err != nil {
		return fmt.Errorf("[handleMessage] failed to parse header: %v", err)
	}
	switch header.Command {
	case CommandNegotiate:
		return s.handleNegotiate(conn, header, msg)
	default:
		return fmt.Errorf("[handleMessage] unsupported command %#x", header.Command)
	}

}
