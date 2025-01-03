package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type SecureConnection struct {
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

	log.Debugf("Accepted connection from %s\n", secure.IP)

	connection.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read their public key first
	var length uint32

	err := binary.Read(connection, binary.LittleEndian, &length)
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

	_, err = connection.Write(secure.CreatePacket(data))
	if err != nil {
		return nil, fmt.Errorf("Failed to write session key: %s", err)
	}

	return secure, nil
}

func (s *SecureConnection) Close() {
	s.connection.Close()

	log.Debugf("Closed connection from %s\n", s.IP)
}

func (s *SecureConnection) CreatePacket(data []byte) []byte {
	packet := make([]byte, len(data)+4)

	binary.LittleEndian.PutUint32(packet, uint32(len(data)))

	copy(packet[4:], data)

	return packet
}

func (s *SecureConnection) Send(data []byte) error {
	encrypted, err := s.key.Encrypt(data)
	if err != nil {
		return err
	}

	_, err = s.connection.Write(s.CreatePacket(encrypted))

	return err
}

func (s *SecureConnection) WaitForData() ([]byte, error) {
	s.connection.SetReadDeadline(time.Now().Add(10 * time.Minute))

	// First read the length of the data
	var length uint32

	err := binary.Read(s.connection, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	// Check that the packet is not too large
	if length > MaxPacketSize {
		return nil, fmt.Errorf("packet too large")
	}

	// Then read the data
	buf := make([]byte, length)

	_, err = io.ReadFull(s.connection, buf)
	if err != nil {
		return nil, err
	}

	return s.key.Decrypt(buf)
}

func (s *SecureConnection) Acknowledge() {
	s.Send([]byte("ACK"))
}

func (s *SecureConnection) Error() {
	s.Send([]byte("ERR"))
}
