package telemetrylog

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"math"
	"os"
	"time"

	"github.com/kvoloboi/telemetry/internal/domain"
)

const (
	magicValue = 0x544C5942 // "TLYB"
	formatVer  = 1

	// field sizes
	magicLen     = 4
	versionLen   = 1
	flagsLen     = 1
	reservedLen  = 2
	timestampLen = 8
	seqLen       = 8
	payloadLen   = 4

	headerLen = magicLen +
		versionLen +
		flagsLen +
		reservedLen +
		timestampLen +
		payloadLen +
		seqLen

	crcLen = 4
)

// header field offsets (little endian)
const (
	offMagic      = 0
	offVersion    = offMagic + magicLen
	offFlags      = offVersion + versionLen
	offReserved   = offFlags + flagsLen
	offTimestamp  = offReserved + reservedLen
	offSeq        = offTimestamp + timestampLen
	offPayloadLen = offSeq + seqLen
)

var (
	ErrPartialBatch = errors.New("partial batch detected")
	ErrLogClosed    = errors.New("telemetry log closed")
	ErrTooLarge     = errors.New("batch too large")
	ErrCorruptLog   = errors.New("log corruption detected")
)

// TelemetryBatch represents a batch of telemetry events
type TelemetryBatch struct {
	Events []domain.Telemetry
}

// TelemetryLog writes batches to disk.
// It is NOT safe for concurrent use.
// All writes must be serialized by the caller.
type TelemetryLog struct {
	f      *os.File
	seq    uint64
	closed bool
}

// Open opens or creates a WAL-style telemetry log and recovers partial batches.
func Open(path string) (*TelemetryLog, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	tl := &TelemetryLog{f: f}

	// Recover partial batches and set seq to last batch + 1
	if err := tl.recover(); err != nil && err != ErrPartialBatch {
		f.Close()
		return nil, err
	}

	// Seek to end for appends
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return nil, err
	}

	return tl, nil
}

// Append writes a batch: header + payload + CRC32
func (tl *TelemetryLog) Append(events []domain.Telemetry) error {
	if tl.closed {
		return ErrLogClosed
	}

	payload, err := marshal(events)
	if err != nil {
		return err
	}

	if len(payload) > math.MaxUint32 {
		return ErrTooLarge
	}

	header := recordHeader{
		magic:      magicValue,
		version:    formatVer,
		flags:      0,
		reserved:   [2]byte{0, 0},
		timestamp:  time.Now().UnixNano(),
		payloadLen: uint32(len(payload)),
		seq:        tl.seq,
	}

	// single buffer allocation for header + payload + CRC
	record := make([]byte, headerLen+len(payload)+crcLen)
	header.encode(record[:headerLen])
	copy(record[headerLen:], payload)

	crc := crc32.ChecksumIEEE(record[:headerLen+len(payload)])
	binary.LittleEndian.PutUint32(record[headerLen+len(payload):], crc)

	if _, err := tl.f.Write(record); err != nil {
		return err
	}

	if err := tl.f.Sync(); err != nil {
		return err
	}

	tl.seq++
	return nil
}

// Close the log
func (tl *TelemetryLog) Close() error {
	if tl.closed {
		return nil
	}
	tl.closed = true
	return tl.f.Close()
}

// recover scans WAL and truncates partial or corrupted batches
func (tl *TelemetryLog) recover() error {
	info, err := tl.f.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	offset := int64(0)

	var (
		hdrBuf [headerLen]byte
		crcBuf [crcLen]byte
	)
	for offset+headerLen+crcLen <= size {
		if _, err := tl.f.ReadAt(hdrBuf[:], offset); err != nil {
			return tl.truncate(offset)
		}
		hdr, err := decodeHeader(hdrBuf[:])

		if err != nil {
			return tl.truncate(offset)
		}
		recordLen := int64(headerLen) + int64(hdr.payloadLen) + crcLen
		if offset+recordLen > size {
			return tl.truncate(offset)
		}

		payload := make([]byte, hdr.payloadLen)
		if _, err := tl.f.ReadAt(payload, offset+headerLen); err != nil {
			return tl.truncate(offset)
		}
		if _, err := tl.f.ReadAt(crcBuf[:], offset+headerLen+int64(hdr.payloadLen)); err != nil {
			return tl.truncate(offset)
		}
		storedCRC := binary.LittleEndian.Uint32(crcBuf[:])

		// streaming CRC check
		crc := crc32.NewIEEE()
		if _, err := crc.Write(hdrBuf[:]); err != nil {
			return err
		}
		if _, err := crc.Write(payload); err != nil {
			return err
		}
		if crc.Sum32() != storedCRC {
			return tl.truncate(offset)
		}

		offset += recordLen
		tl.seq++
	}

	return nil
}

func (tl *TelemetryLog) truncate(offset int64) error {
	return tl.f.Truncate(offset)
}
