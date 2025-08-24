package zipfile

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

type CompressionMethod int

const (
	NoCompression CompressionMethod = iota
	Shrunk
	Reduced1
	Reduced2
	Reduced3
	Reduced4
	Imploded
	_
	Deflated
	EnhancedDeflated
	PKWare
	_
	Bzip2
	_
	LZMA
)

const (
	localHeaderSize = 30
	zip64HeadSize   = 20
)

var (
	zipMagicNumber = []byte{80, 75, 3, 4}
	zip64Marker    = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

var ErrNoMoreHeaders = errors.New("no more headers")
var ErrCannotDecompress = errors.New("cannot decompress file")

type LocalHeader struct {
	Version           int
	Flags             []byte
	IsZIP64           bool
	Name              string
	LastModified      time.Time
	CompressionMethod CompressionMethod
	CompressedSize    int
	UncompressedSize  int
	ExtraField        []byte
	Content           []byte
}

func (header LocalHeader) GetContent() (string, error) {
	switch header.CompressionMethod {
	case NoCompression:
		return string(header.Content[:]), nil
	case Deflated:
		r := flate.NewReader(bytes.NewReader(header.Content))
		contentBytes := make([]byte, header.UncompressedSize)
		_, err := r.Read(contentBytes)
		if err != nil {
			return "", err
		}
		return string(contentBytes[:]), nil
	}
	return "", ErrCannotDecompress
}

func (header LocalHeader) String() string {
	return fmt.Sprintf("LocalHeader{Name:%q, Flags:%q, Extra:%q, Size:%d->%d, Compression:%d}", header.Name, header.Flags, header.ExtraField, header.UncompressedSize, header.CompressedSize, header.CompressionMethod)
}

func readHeader(r io.ReadSeeker) (*LocalHeader, error) {
	headerBytes := make([]byte, localHeaderSize)
	err := readExact(r, headerBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read header: %w", err)
	}

	if !bytes.Equal(zipMagicNumber, headerBytes[0:4]) {
		_, err := r.Seek(-1*localHeaderSize, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("unable to rewind local header after reaching end: %w", err)
		}
		return nil, ErrNoMoreHeaders
	}

	header := LocalHeader{}
	header.Version = int(binary.LittleEndian.Uint16(headerBytes[4:6]))
	header.Flags = headerBytes[6:8]
	header.CompressionMethod = CompressionMethod(binary.LittleEndian.Uint16(headerBytes[8:10]))

	// Last Modified date time
	dosTime := binary.LittleEndian.Uint16(headerBytes[10:12])
	second := int((dosTime & 0x1F) << 2)
	minute := int((dosTime >> 5) & 0x3F)
	hour := int((dosTime >> 11) & 0x1F)
	dosDate := binary.LittleEndian.Uint16(headerBytes[12:14])
	day := int(dosDate & 0x1F)
	month := int((dosDate >> 5) & 0x0F)
	year := int((dosDate>>9)&0x7F) + 1980
	header.LastModified = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	if bytes.Equal(headerBytes[18:26], zip64Marker) {
		header.IsZIP64 = true
	} else {
		header.IsZIP64 = false
		header.CompressedSize = int(binary.LittleEndian.Uint32(headerBytes[18:22]))
		header.UncompressedSize = int(binary.LittleEndian.Uint32(headerBytes[22:26]))
	}

	// Filename
	filenameLength := int(binary.LittleEndian.Uint16(headerBytes[26:28]))
	filenameBytes := make([]byte, filenameLength)
	err = readExact(r, filenameBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read filename: %w", err)
	}
	header.Name = string(filenameBytes[:])

	// Extra field
	extraFieldLength := int(binary.LittleEndian.Uint16(headerBytes[28:30]))
	header.ExtraField = make([]byte, extraFieldLength)
	err = readExact(r, header.ExtraField)
	if err != nil {
		return nil, fmt.Errorf("unable to read extra field: %w", err)
	}

	if header.IsZIP64 {
		zip64FieldHeader := make([]byte, 20)
		err = readExact(r, zip64FieldHeader)
		if err != nil {
			return nil, fmt.Errorf("unable to read ZIP64 header: %w", err)
		}
		header.UncompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[4:12]))
		header.CompressedSize = int(binary.LittleEndian.Uint64(zip64FieldHeader[12:20]))
	}

	header.Content = make([]byte, header.CompressedSize)
	err = readExact(r, header.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to read content: %w", err)
	}

	return &header, nil
}
