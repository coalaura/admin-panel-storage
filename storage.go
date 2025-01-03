package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct {
	Root string
}

type License [40]byte

type RequestHeader struct {
	Type    uint8
	Server  uint8
	Start   uint32
	End     uint32
	License License
}

const (
	RequestStore   = 1
	RequestReadOne = 2
	RequestReadAll = 3

	MinTimestamp = 946684800 // 2000-01-01
)

func NewStorage(root string) (*Storage, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		os.MkdirAll(abs, 0755)
	}

	return &Storage{
		Root: abs,
	}, nil
}

func (s *Storage) Paths(header *RequestHeader, license string) (string, string) {
	date1 := time.Unix(int64(header.Start), 0)
	date2 := time.Unix(int64(header.End), 0)

	path1 := fmt.Sprintf("c%d/%s", header.Server, date1.Format("2006-01-02"))
	path2 := fmt.Sprintf("c%d/%s", header.Server, date2.Format("2006-01-02"))

	if license != "" {
		path1 = fmt.Sprintf("%s/%s", path1, license)
		path2 = fmt.Sprintf("%s/%s", path2, license)
	}

	return filepath.Join(s.Root, path1), filepath.Join(s.Root, path2)
}

func (s *Storage) HandleRequest(request []byte) ([]byte, error) {
	reader := bytes.NewReader(request)

	var header RequestHeader

	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	err = header.Validate()
	if err != nil {
		return nil, err
	}

	license := header.LicenseString()

	path1, path2 := s.Paths(&header, license)

	switch header.Type {
	case RequestStore:
		return header.Store(license, path1, reader)
	case RequestReadOne:
		return header.ReadOne(license, path1, path2)
	case RequestReadAll:
		return header.ReadAll(license, path1, path2)
	}

	return nil, fmt.Errorf("invalid request type %d", header.Type)
}

func (h *RequestHeader) LicenseString() string {
	var license strings.Builder

	for _, b := range h.License {
		if b == 0 {
			continue
		}

		license.WriteByte(b)
	}

	if license.Len() != 40 {
		return ""
	}

	return license.String()
}

func (h *RequestHeader) Validate() error {
	// Check that the start and end timestamps are valid
	if h.Start <= MinTimestamp || h.End < h.Start {
		return errors.New("invalid timestamps")
	}

	// Check that the server is valid
	if h.Server == 0 {
		return errors.New("invalid server")
	}

	return nil
}

func (h *RequestHeader) Store(license, path string, reader *bytes.Reader) ([]byte, error) {
	if license == "" {
		return nil, fmt.Errorf("storing requires license")
	}

	err := EnsureDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	log.Debugf("Received store request for %s (%d)\n", license, h.Start)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *RequestHeader) ReadOne(license, path1, path2 string) ([]byte, error) {
	if license == "" {
		return nil, fmt.Errorf("read-one requires license")
	}

	log.Debugf("Received read request for %s (%d - %d)\n", license, h.Start, h.End)

	var (
		total  = int(h.End-h.Start) * HistoricEntrySize
		buffer = bytes.NewBuffer(make([]byte, 0, total))
	)

	err := ReadHistoricFiles(buffer, path1, path2, h.Start, h.End)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (h *RequestHeader) ReadAll(license, path1, path2 string) ([]byte, error) {
	if license != "" {
		return nil, fmt.Errorf("read-all requires no license")
	}

	log.Debugf("Received read-all request for c%d (%d - %d)\n", h.Server, h.Start, h.End)

	var buffer bytes.Buffer

	err := ReadAllHistoricFiles(&buffer, path1, h.Start)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func EnsureDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return os.MkdirAll(directory, 0755)
	}

	return nil
}
