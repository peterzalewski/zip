package structure

import (
	"fmt"
	"io"
)

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

func Parse(r io.ReadSeeker) ([]LocalHeader, error) {
	var headers []LocalHeader
	for {
		header, err := readHeader(r)
		if err != nil {
			break
		}
		headers = append(headers, *header)
	}
	return headers, nil
}
