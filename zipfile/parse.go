package zipfile

import (
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
	var headers []LocalHeader
	for {
		header, err := readHeader(r)
		if err != nil {
			if errors.Is(err, ErrNoMoreHeaders) {
				break
			} else {
				return nil, err
			}
		}
		headers = append(headers, *header)
	}

	var cdfhs []CentralDirectoryFileHeader
	for {
		cdfh, err := readCentralDirectoryFileHeader(r)
		if err != nil {
			if errors.Is(err, ErrNoMoreCDFHs) {
				break
			} else {
				return nil, err
			}
		}
		cdfhs = append(cdfhs, *cdfh)
	}

	eocdr, err := readEndOfCentralDirectoryRecord(r)
	if err != nil {
		return nil, err
	}

	_, err = r.Read(make([]byte, 1))
	if !errors.Is(err, io.EOF) {
		return nil, ErrDataAfterEndRecord
	}

	return &ZipFile{LocalHeaders: headers, CentralDirectory: cdfhs, EndRecord: *eocdr}, nil
}
