package main

import (
	"fmt"
	"io"
	"net"
)

type TCPConnection struct {
	key        *EncryptionKey
	connection net.Conn
}

func StartTCPServer() error {
	hostname := fmt.Sprintf("%s:%d", config.Hostname, config.Port)

	listener, err := net.Listen("tcp", hostname)
	if err != nil {
		return err
	}

	defer listener.Close()

	log.Infof("Listening at %s...\n", hostname)

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Warningf("Failed to accept connection: %s\n", err)

			continue
		}

		ip := connection.RemoteAddr().String()

		if !config.IsIPAllowed(ip) {
			connection.Close()

			log.Warningf("Rejected connection from %s\n", ip)

			continue
		}

		secure, err := NewSecureConnection(connection)
		if err != nil {
			connection.Close()

			log.Warningf("Failed to create secure connection with %s: %s\n", ip, err)

			continue
		}

		go HandleConnection(secure)
	}
}

func HandleConnection(secure *SecureConnection) {
	defer secure.Close()

	for {
		data, err := secure.WaitForData()
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Warningf("Failed to receive data from %s: %s\n", secure.IP, err)

			return
		}

		response, err := storage.HandleRequest(data)
		if err != nil {
			log.Warningf("Failed to handle request from %s: %s\n", secure.IP, err)

			secure.Error()

			continue
		}

		if response == nil {
			secure.Acknowledge()
		} else {
			secure.Send(response)
		}
	}
}
