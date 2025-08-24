package zipfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type ZipFile struct {
	LocalHeaders     []LocalHeader
	CentralDirectory []CentralDirectoryFileHeader
	EndRecord        EndOfCentralDirectoryRecord
}

var ErrDataAfterEndRecord = errors.New("unexpected data at end of file")
var ErrCannotFindEndRecord = errors.New("cannot find end of central directory record")

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

func Parse(r io.ReadSeeker) (*ZipFile, error) {
	fileSize, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("unable to determine file size: %w", err)
	}

	seekBeginning := min(MaxEndOfCentralDirectoryRecordSize, fileSize)
	r.Seek(-1*seekBeginning, io.SeekEnd)
	possibleEndRecordBytes := make([]byte, MaxEndOfCentralDirectoryRecordSize)

	_, err = r.Read(possibleEndRecordBytes)
	if err != nil {
		return nil, err
	}

	recordIndex := int64(bytes.LastIndex(possibleEndRecordBytes, EndOfCentralDirectoryRecordMagicNumber))
	if recordIndex == -1 {
		return nil, ErrCannotFindEndRecord
	}

	r.Seek(fileSize-seekBeginning+recordIndex, io.SeekStart)
	eocdr, err := readEndOfCentralDirectoryRecord(r)
	if err != nil {
		return nil, err
	}

	r.Seek(eocdr.CentralDirectoryOffset, io.SeekStart)
	var cdfhs []CentralDirectoryFileHeader
	for i := 0; i < eocdr.NumberOfRecords; i++ {
		cdfh, err := readCentralDirectoryFileHeader(r)
		if err != nil {
			return nil, err
		}
		cdfhs = append(cdfhs, *cdfh)
	}

	var headers []LocalHeader
	for _, cdfh := range cdfhs {
		r.Seek(cdfh.LocalHeadOffset, io.SeekStart)
		header, err := readHeader(r)
		if err != nil {
			return nil, err
		}
		headers = append(headers, *header)
	}

	return &ZipFile{LocalHeaders: headers, CentralDirectory: cdfhs, EndRecord: *eocdr}, nil
}
