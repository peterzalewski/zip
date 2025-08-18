package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type LocalHeader struct {
	Version int
	Flags []byte
	IsZIP64 bool
	Name string
	CompressedSize int
	UncompressedSize int
	Content []byte
}

func readHeader(f *os.File) (*LocalHeader, error) {
	headerBytes := make([]byte, 30)
	num_read, err := f.Read(headerBytes)
	if err != nil {
		return nil, errors.New("Unable to read header")
	}

	if num_read != 30 {
		return nil, errors.New("Did not read initial header info")
	}

	if !bytes.Equal([]byte{80, 75, 3, 4}, headerBytes[0:4]) {
		return nil, errors.New("Magic number missing")
	}

	header := LocalHeader{}
	header.Version = int(binary.LittleEndian.Uint16(headerBytes[4:6]))
	header.Flags = headerBytes[6:8]

	if bytes.Equal(headerBytes[18:26], []byte{255, 255, 255, 255, 255, 255, 255, 255}) {
		header.IsZIP64 = true
	} else {
		header.IsZIP64 = false
		header.CompressedSize = int(binary.LittleEndian.Uint32(headerBytes[18:22]))
		header.UncompressedSize = int(binary.LittleEndian.Uint32(headerBytes[22:26]))
	}

	filenameLength := int(binary.LittleEndian.Uint16(headerBytes[26:30]))
	filenameBytes := make([]byte, filenameLength)
	nameBytesRead, err := f.Read(filenameBytes)
	if err != nil {
		return nil, errors.New("Unable to read filename")
	}

	if nameBytesRead != filenameLength {
		return nil, fmt.Errorf("Expected %d bytes, got %d", filenameLength, nameBytesRead)
	}

	header.Name = string(filenameBytes[:])

	if header.IsZIP64 {
		zip64FieldHeader := make([]byte, 20)
		z64BytesRead, err := f.Read(zip64FieldHeader)

		if err != nil {
			return nil, err
		}

		if z64BytesRead != 20 {
			return nil, fmt.Errorf("Expected %d bytes from ZIP64 header, got %d", 20, z64BytesRead)
		}

		header.UncompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[4:12]))
		header.CompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[12:20]))
	}

	header.Content = make([]byte, header.CompressedSize)
	contentRead, err := f.Read(header.Content)
	if err != nil {
			return nil, err
	}

	if contentRead != header.CompressedSize {
		return nil, fmt.Errorf("Expected %d bytes from content, got %d", header.CompressedSize, contentRead)
	}

	return &header, nil
}

func main() {
	abspath, err := filepath.Abs("./test.zip")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	f, err := os.Open(abspath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer f.Close()
	header, err := readHeader(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(header.Content[:]))
}
