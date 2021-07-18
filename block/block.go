package block

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BlockType int

const (
	Address BlockType = iota
	Empty
	Object
	Binary
	Boolean
	String
	Int
	Float
	DateTime
)

func (blockType BlockType) Bitmask() byte {
	return byte(blockType) << 4
}

func ParseBlockTypeBitmask(bitmask byte) BlockType {
	return BlockType(bitmask >> 4)
}

type BlockAddress uint

func (blockAddress BlockAddress) ToBytes(addressBytes int) (address []byte, err error) {
	var buffer []byte

	switch addressBytes {
	case 1:
		buffer = make([]byte, 2)
		binary.LittleEndian.PutUint16(buffer, uint16(blockAddress))

		address = []byte{buffer[0]}
	case 2:
		address = make([]byte, 2)
		binary.LittleEndian.PutUint16(address, uint16(blockAddress))
	case 3:
		buffer = make([]byte, 4)
		binary.LittleEndian.PutUint32(buffer, uint32(blockAddress))

		address = []byte{buffer[0], buffer[1], buffer[2]}
	case 4:
		address = make([]byte, 4)
		binary.LittleEndian.PutUint32(address, uint32(blockAddress))
	case 5:
		buffer = make([]byte, 8)
		binary.LittleEndian.PutUint64(buffer, uint64(blockAddress))

		address = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4]}
	case 6:
		buffer = make([]byte, 8)
		binary.LittleEndian.PutUint64(buffer, uint64(blockAddress))

		address = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5]}
	case 7:
		buffer = make([]byte, 8)
		binary.LittleEndian.PutUint64(buffer, uint64(blockAddress))

		address = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5], buffer[6]}
	case 8:
		address = make([]byte, 8)
		binary.LittleEndian.PutUint64(address, uint64(blockAddress))
	default:
		err = fmt.Errorf("maximum address size exceeded")
	}

	return
}

type Block interface {
	Type() BlockType

	Address() BlockAddress
	SetAddress(BlockAddress) Block

	Key() interface{}
	SetKey(interface{}) error

	IsRequest() bool
	SetIsRequest(bool) Block

	IsResponse() bool
	SetIsResponse(bool) Block

	Encode(io.Writer) (int, error)
}

type Blocks map[BlockAddress]Block

func (blocks Blocks) Lookup(address BlockAddress) (block Block, hasBlock bool) {
	block, hasBlock = blocks[address]
	return
}

func (blocks Blocks) Get(address BlockAddress) (block Block) {
	var hasBlock bool

	if block, hasBlock = blocks.Lookup(address); !hasBlock {
		return nil
	}

	return
}

func (blocks Blocks) Set(address BlockAddress, block Block) {
	blocks[address] = block
}
