package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	localHeaderSize = 30
	zip64HeadSize   = 20
)

var (
	zipMagicNumber = []byte{80, 75, 3, 4}
	zip64Marker    = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

type LocalHeader struct {
	Version          int
	Flags            []byte
	IsZIP64          bool
	Name             string
	CompressedSize   int
	UncompressedSize int
	Content          []byte
}

func readExact(r io.Reader, destination []byte) error {
	n, err := io.ReadFull(r, destination)
	if err != nil {
		return fmt.Errorf("failed to read %d bytes: %w", len(destination), err)
	}
	if n != len(destination) {
		return fmt.Errorf("expected %d bytes, got %d", len(destination), n)
	}
	return nil
}

func readHeader(f *os.File) (*LocalHeader, error) {
	headerBytes := make([]byte, localHeaderSize)
	err := readExact(f, headerBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read header: %w", err)
	}

	if !bytes.Equal(zipMagicNumber, headerBytes[0:4]) {
		return nil, errors.New("Magic number missing")
	}

	header := LocalHeader{}
	header.Version = int(binary.LittleEndian.Uint16(headerBytes[4:6]))
	header.Flags = headerBytes[6:8]

	if bytes.Equal(headerBytes[18:26], zip64Marker) {
		header.IsZIP64 = true
	} else {
		header.IsZIP64 = false
		header.CompressedSize = int(binary.LittleEndian.Uint32(headerBytes[18:22]))
		header.UncompressedSize = int(binary.LittleEndian.Uint32(headerBytes[22:26]))
	}

	filenameLength := int(binary.LittleEndian.Uint16(headerBytes[26:30]))
	filenameBytes := make([]byte, filenameLength)
	err = readExact(f, filenameBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read filename: %w", err)
	}

	header.Name = string(filenameBytes[:])

	if header.IsZIP64 {
		zip64FieldHeader := make([]byte, 20)
		err = readExact(f, zip64FieldHeader)
		if err != nil {
			return nil, fmt.Errorf("unable to read ZIP64 header: %w", err)
		}
		header.UncompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[4:12]))
		header.CompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[12:20]))
	}

	header.Content = make([]byte, header.CompressedSize)
	err = readExact(f, header.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to read content: %w", err)
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

	fmt.Printf("%+v\n", header)
}
