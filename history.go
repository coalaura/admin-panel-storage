package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
)

const HistoricEntrySize = 36

func ReadAllHistoricFiles(writer *bytes.Buffer, path string, timestamp uint32) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var (
		total  uint32
		result bytes.Buffer
	)

	for _, file := range files {
		var buffer bytes.Buffer

		err = ReadHistoryFileSection(&buffer, filepath.Join(path, file.Name()), timestamp, timestamp)
		if err != nil {
			return err
		}

		if buffer.Len() == 0 {
			continue
		}

		err = binary.Write(&result, binary.LittleEndian, License([]byte(file.Name())))
		if err != nil {
			return err
		}

		err = binary.Write(&result, binary.LittleEndian, uint32(buffer.Len()))
		if err != nil {
			return err
		}

		result.Write(buffer.Bytes())

		total++
	}

	err = binary.Write(writer, binary.LittleEndian, total)
	if err != nil {
		return err
	}

	writer.Write(result.Bytes())

	return nil
}

func ReadHistoricFiles(writer *bytes.Buffer, path1, path2 string, start uint32, end uint32) error {
	err := ReadHistoryFileSection(writer, path1, start, end)
	if err != nil {
		return err
	}

	if path1 != path2 {
		return ReadHistoryFileSection(writer, path2, start, end)
	}

	return nil
}

func ReadHistoryFileSection(writer *bytes.Buffer, path string, start uint32, end uint32) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	/**
	 * | Timestamp (ui32) | character_id (ui32) | x (f32) | y (f32) | z (f32) | heading (f32) | speed (f32) | character_flags (ui32) | user_flags (ui32) |
	 */
	var (
		timestamp uint32

		entry = make([]byte, HistoricEntrySize)
	)

	for {
		_, err := io.ReadFull(file, entry)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		timestamp = binary.LittleEndian.Uint32(entry[0:4])

		// If we have not yet reached the start of the section, skip it
		if timestamp < start {
			continue
		}

		// Add the entry to the buffer
		writer.Write(entry)

		// If we have reached the end of the section, stop
		if timestamp >= end {
			break
		}
	}

	return nil
}
