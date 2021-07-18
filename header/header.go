package header

import (
	"fmt"
	"io"
)

const (
	fileSignature                string = "\x89\x41\x50\x4f\x0d\x0a\x1a\x0a"
	isExtensionFlag              byte   = 0x8
	enableMemoryOptimizationFlag byte   = 0x4
)

type Header struct {
	Version                  Version
	IndexChecksum            Checksum
	BlocksChecksum           Checksum
	IsExtension              bool
	EnableMemoryOptimization bool
	AddressBytes             int
}

func NewHeader() *Header {
	return &Header{}
}

// Version:						8 bits		1 byte

// AddressBytes:				3 bits		|
// IsExtension:					1 bit		| 1 byte
// EnableMemoryOptimization:	1 bit		|

// IndexChecksum:				64 bits		8 bytes
// BlocksChecksum:				64 bits		8 bytes

func (header *Header) Encode(writer io.Writer) (int, error) {
	var (
		flags byte
		data  []byte = []byte(fileSignature)
	)

	data = append(data, header.Version.ToByte())

	flags = byte(header.AddressBytes-1) << 4

	if header.IsExtension {
		flags = flags | isExtensionFlag
	}

	if header.EnableMemoryOptimization {
		flags = flags | enableMemoryOptimizationFlag
	}

	data = append(data, flags)

	data = append(data, header.IndexChecksum.Value...)
	data = append(data, header.BlocksChecksum.Value...)

	return writer.Write(data)
}

func (header *Header) Decode(data []byte) (err error) {
	if len(data) < 26 {
		err = fmt.Errorf("APO header is too short")
	}

	if string(data[0:8]) != fileSignature {
		err = fmt.Errorf("not APO file")
		return
	}

	header.Version.decode(data[8])

	flags := data[9]

	header.AddressBytes = int(flags>>4) + 1

	header.IsExtension = (flags & isExtensionFlag) == isExtensionFlag
	header.EnableMemoryOptimization = (flags & enableMemoryOptimizationFlag) == enableMemoryOptimizationFlag

	header.IndexChecksum = Checksum{Value: data[10:18]}
	header.BlocksChecksum = Checksum{Value: data[18:26]}

	return
}
