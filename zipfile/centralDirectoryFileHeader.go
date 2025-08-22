package zipfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	centralDirectoryFileHeaderSize = 46
)

var (
	centralDirectoryFileHeaderMagicNumber = []byte{0x50, 0x4B, 0x01, 0x02}
)

var ErrNoMoreCDFHs = errors.New("no more directory headers")

type CentralDirectoryFileHeader struct {
	FileName   string
	ExtraField []byte
	Comment    string
}

func (cdfh CentralDirectoryFileHeader) String() string {
	return fmt.Sprintf("CentralDirectoryFileHeader{Name:%q, Comment:%q}", cdfh.FileName, cdfh.Comment)
}

func readCentralDirectoryFileHeader(r io.ReadSeeker) (*CentralDirectoryFileHeader, error) {
	headerBytes := make([]byte, centralDirectoryFileHeaderSize)
	err := readExact(r, headerBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read CDFH: %w", err)
	}

	if !bytes.Equal(centralDirectoryFileHeaderMagicNumber, headerBytes[0:4]) {
		_, err := r.Seek(-1*centralDirectoryFileHeaderSize, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("unable to rewind local cdfh after reaching end: %w", err)
		}
		return nil, ErrNoMoreCDFHs
	}

	filenameLength := int(binary.LittleEndian.Uint16(headerBytes[28:30]))
	extraFieldLength := int(binary.LittleEndian.Uint16(headerBytes[30:32]))
	commentLength := int(binary.LittleEndian.Uint16(headerBytes[32:34]))

	cdfh := CentralDirectoryFileHeader{}

	filenameBytes := make([]byte, filenameLength)
	err = readExact(r, filenameBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read filename: %w", err)
	}
	cdfh.FileName = string(filenameBytes[:])

	cdfh.ExtraField = make([]byte, extraFieldLength)
	err = readExact(r, cdfh.ExtraField)
	if err != nil {
		return nil, fmt.Errorf("unable to read extra field: %w", err)
	}

	commentBytes := make([]byte, commentLength)
	err = readExact(r, commentBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read comment: %w", err)
	}
	cdfh.Comment = string(commentBytes[:])

	return &cdfh, nil
}
