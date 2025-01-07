package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type SecureConnection struct {
	mx sync.Mutex

	key        *SessionKey
	connection net.Conn

	IP string
}

const MaxPacketSize = 64 * 1024

func NewSecureConnection(connection net.Conn) (*SecureConnection, error) {
	secure := &SecureConnection{
		connection: connection,

		IP: connection.RemoteAddr().String(),
	}

	log.Notef("Accepted connection from %s\n", secure.IP)

	connection.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Request ID is always 0
	var (
		requestId uint32
		length    uint32
	)

	err := binary.Read(connection, binary.LittleEndian, &requestId)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request ID: %s", err)
	}

	if requestId != 0 {
		return nil, fmt.Errorf("Invalid request ID for handshake: %d", requestId)
	}

	// Read their public key first
	err = binary.Read(connection, binary.LittleEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("Failed to read public key length: %s", err)
	}

	buf := make([]byte, length)

	_, err = io.ReadFull(connection, buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to read public key: %s", err)
	}

	key, err := NewEncryptionKey(buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse public key: %s", err)
	}

	session, err := NewSessionKey()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate session key: %s", err)
	}

	secure.key = session

	// Then we respond with the session key, encrypted using their key
	data, err := key.Encrypt(session.key)
	if err != nil {
		return nil, fmt.Errorf("Failed to encrypt session key: %s", err)
	}

	_, err = connection.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to write session key: %s", err)
	}

	return secure, nil
}

func (s *SecureConnection) Close() {
	s.connection.Close()

	log.Notef("Closed connection from %s\n", s.IP)
}

func (s *SecureConnection) CreatePacket(requestId uint32, data []byte) []byte {
	packet := make([]byte, len(data)+4+4)

	// Write the request id
	binary.LittleEndian.PutUint32(packet[0:4], requestId)

	// Write the data length
	binary.LittleEndian.PutUint32(packet[4:8], uint32(len(data)))

	// Write the data
	copy(packet[8:], data)

	return packet
}

func (s *SecureConnection) Send(requestId uint32, data []byte) error {
	encrypted, err := s.key.Encrypt(data)
	if err != nil {
		return err
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err = s.connection.Write(s.CreatePacket(requestId, encrypted))

	return err
}

func (s *SecureConnection) WaitForData() (uint32, []byte, error) {
	s.connection.SetReadDeadline(time.Now().Add(10 * time.Minute))

	var (
		requestId uint32
		length    uint32
	)

	// First read the request id
	err := binary.Read(s.connection, binary.LittleEndian, &requestId)
	if err != nil {
		return 0, nil, err
	}

	// Then read the length
	err = binary.Read(s.connection, binary.LittleEndian, &length)
	if err != nil {
		return 0, nil, err
	}

	// Check that the packet is not too large
	if length > MaxPacketSize {
		return 0, nil, fmt.Errorf("packet too large")
	}

	// Then read the data
	buf := make([]byte, length)

	_, err = io.ReadFull(s.connection, buf)
	if err != nil {
		return 0, nil, err
	}

	data, err := s.key.Decrypt(buf)
	if err != nil {
		return 0, nil, err
	}

	return requestId, data, nil
}

func (s *SecureConnection) Acknowledge(requestId uint32) {
	s.Send(requestId, []byte("ACK"))
}

func (s *SecureConnection) Error(requestId uint32) {
	s.Send(requestId, []byte("ERR"))
}
