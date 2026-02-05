package telemetrylog

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"

	"github.com/kvoloboi/telemetry/internal/domain"
)

// BatchReader reads a snapshot of the log at open time.
// Appends after creation are not visible.
type BatchReader struct {
	f      *os.File
	offset int64
	size   int64
}

func NewBatchReader(path string) (*BatchReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &BatchReader{f: f, size: info.Size(), offset: 0}, nil
}

func (r *BatchReader) Next() ([]domain.Telemetry, error) {
	if r.offset >= r.size {
		return nil, io.EOF
	}

	var headerBuf [headerLen]byte
	if _, err := r.f.ReadAt(headerBuf[:], r.offset); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, err
	}

	hdr, err := decodeHeader(headerBuf[:])
	if err != nil {
		return nil, err
	}

	recordLen := int64(headerLen) + int64(hdr.payloadLen) + crcLen
	if r.offset+recordLen > r.size {
		return nil, ErrPartialBatch
	}

	payload := make([]byte, hdr.payloadLen)
	if _, err := r.f.ReadAt(payload, r.offset+headerLen); err != nil {
		return nil, err
	}

	var crcBuf [crcLen]byte
	if _, err := r.f.ReadAt(crcBuf[:], r.offset+headerLen+int64(hdr.payloadLen)); err != nil {
		return nil, err
	}
	storedCRC := binary.LittleEndian.Uint32(crcBuf[:])

	crc := crc32.NewIEEE()
	if _, err := crc.Write(headerBuf[:]); err != nil {
		return nil, err
	}
	if _, err := crc.Write(payload); err != nil {
		return nil, err
	}

	if crc.Sum32() != storedCRC {
		return nil, ErrCorruptLog
	}

	r.offset += recordLen
	return unmarshal(payload)
}

func (r *BatchReader) Close() error {
	return r.f.Close()
}
