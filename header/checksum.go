package header

import (
	"bytes"
	"encoding/binary"
	"hash/crc64"
)

type Checksum struct {
	Value   []byte
	IsValid bool
}

func CalcuateChecksum(buffer *bytes.Buffer) Checksum {
	var checksumBuffer []byte = make([]byte, 8)

	binary.LittleEndian.PutUint64(checksumBuffer, crc64.Checksum(buffer.Bytes(), crc64.MakeTable(crc64.ECMA)))

	return Checksum{
		Value:   checksumBuffer,
		IsValid: true,
	}
}
