package telemetrylog

import "encoding/binary"

type recordHeader struct {
	magic      uint32
	version    uint8
	flags      uint8
	reserved   [2]byte
	timestamp  int64
	seq        uint64
	payloadLen uint32
}

func (h recordHeader) encode(buf []byte) {
	binary.LittleEndian.PutUint32(buf[offMagic:], h.magic)

	buf[offVersion] = h.version
	buf[offFlags] = h.flags
	copy(buf[offReserved:], h.reserved[:])

	binary.LittleEndian.PutUint64(buf[offTimestamp:], uint64(h.timestamp))
	binary.LittleEndian.PutUint64(buf[offSeq:], h.seq)
	binary.LittleEndian.PutUint32(buf[offPayloadLen:], h.payloadLen)
}

func decodeHeader(buf []byte) (recordHeader, error) {
	if len(buf) != headerLen {
		return recordHeader{}, ErrPartialBatch
	}

	if binary.LittleEndian.Uint32(buf[offMagic:]) != magicValue {
		return recordHeader{}, ErrCorruptLog
	}

	h := recordHeader{
		magic:      magicValue,
		version:    buf[offVersion],
		flags:      buf[offFlags],
		timestamp:  int64(binary.LittleEndian.Uint64(buf[offTimestamp:])),
		seq:        binary.LittleEndian.Uint64(buf[offSeq:]),
		payloadLen: binary.LittleEndian.Uint32(buf[offPayloadLen:]),
	}

	copy(h.reserved[:], buf[offReserved:offReserved+reservedLen])

	if h.version != formatVer {
		return recordHeader{}, ErrCorruptLog
	}

	return h, nil
}
