package zipfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	endOfCentralDirectoryRecordSize    = 22
	MaxEndOfCentralDirectoryRecordSize = endOfCentralDirectoryRecordSize + 1<<16
)

var (
	EndOfCentralDirectoryRecordMagicNumber = []byte{0x50, 0x4B, 0x05, 0x06}
)

var ErrUnexpectedEndOfZipFile = errors.New("reached end of zip file before eocdr")

type EndOfCentralDirectoryRecord struct {
	NumberOfRecords int
	Comment         string
}

func (endRecord EndOfCentralDirectoryRecord) String() string {
	return fmt.Sprintf("EndOfCentralDirectoryRecord{NumberOfRecords:%d, Comment:%q}", endRecord.NumberOfRecords, endRecord.Comment)
}

func readEndOfCentralDirectoryRecord(r io.Reader) (*EndOfCentralDirectoryRecord, error) {
	recordBytes := make([]byte, endOfCentralDirectoryRecordSize)
	err := readExact(r, recordBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read end record: %w", err)
	}

	if !bytes.Equal(EndOfCentralDirectoryRecordMagicNumber, recordBytes[0:4]) {
		return nil, ErrUnexpectedEndOfZipFile
	}

	record := EndOfCentralDirectoryRecord{}
	record.NumberOfRecords = int(binary.LittleEndian.Uint16(recordBytes[10:12]))
	commentLength := int(binary.LittleEndian.Uint16(recordBytes[20:22]))
	commentBytes := make([]byte, commentLength)
	err = readExact(r, commentBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to read comment: %w", err)
	}
	record.Comment = string(commentBytes[:])

	return &record, nil
}
