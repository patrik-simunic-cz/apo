package index

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/header"
)

type Flag byte

const (
	BitmaskRequest  Flag = 0x8
	BitmaskResponse Flag = 0x4
	BitmaskIntKey   Flag = 0x2
	BitmaskA        Flag = 0x1
	bitmaskNegative byte = 0x1
)

type BlockIndex struct {
	address   block.BlockAddress
	bitmask   byte
	BlockSize uint32
	Type      block.BlockType
	Key       interface{}
}

func (blockIndex *BlockIndex) Bitmask() byte {
	return blockIndex.bitmask
}

func (blockIndex *BlockIndex) HasFlag(flag Flag) bool {
	return byte(flag) == byte(flag)&blockIndex.bitmask
}

func (blockIndex *BlockIndex) setBitmask(bitmask byte) {
	blockIndex.bitmask = blockIndex.Type.Bitmask() | ((bitmask << 4) >> 4)
}

func (blockIndex *BlockIndex) EnableFlag(flag Flag) {
	blockIndex.setBitmask(blockIndex.bitmask | byte(flag))
}

func (blockIndex *BlockIndex) DisableFlag(flag Flag) {
	blockIndex.setBitmask(blockIndex.bitmask & (^byte(flag)))
}

func (blockIndex *BlockIndex) GetKey() (key interface{}) {
	return blockIndex.Key
}

func (blockIndex *BlockIndex) SetKey(key interface{}) (err error) {
	switch key.(type) {
	case string:
		blockIndex.DisableFlag(BitmaskIntKey)
	case int:
		blockIndex.EnableFlag(BitmaskIntKey)
	default:
		err = fmt.Errorf("invalid key type: %T", key)
		return
	}

	blockIndex.Key = key
	return
}

func (blockIndex *BlockIndex) ToBytes(header *header.Header) (data []byte, err error) {
	var (
		addressBytes    []byte
		blockSizeBuffer []byte
		indexSize       int
		indexSizeBuffer []byte
	)

	data = []byte{blockIndex.bitmask}

	if addressBytes, err = blockIndex.address.ToBytes(header.AddressBytes); err != nil {
		return
	}

	data = append(data, addressBytes...)

	blockSizeBuffer = make([]byte, 4)
	binary.LittleEndian.PutUint32(blockSizeBuffer, blockIndex.BlockSize)
	data = append(data, blockSizeBuffer...)

	if blockIndex.HasFlag(BitmaskIntKey) {
		var (
			keyBuffer   []byte
			unsignedKey uint
			key         int  = blockIndex.Key.(int)
			isNegative  bool = key < 0
		)

		if isNegative {
			unsignedKey = uint(key * -1)
		} else {
			unsignedKey = uint(key)
		}

		keyBuffer = make([]byte, reflect.TypeOf(unsignedKey).Size())

		switch len(keyBuffer) {
		case 4:
			binary.LittleEndian.PutUint32(keyBuffer, uint32(unsignedKey))
		case 8:
			binary.LittleEndian.PutUint64(keyBuffer, uint64(unsignedKey))
		default:
			err = fmt.Errorf("invalid int key size")
			return
		}

		if isNegative {
			keyBuffer[len(keyBuffer)-1] = keyBuffer[len(keyBuffer)-1] | bitmaskNegative
		}

		data = append(data, keyBuffer...)
	} else if blockIndex.Key != nil {
		data = append(data, []byte(blockIndex.Key.(string))...)
	}

	indexSize = len(data)

	if indexSize >= 65536 { // 2^16
		err = fmt.Errorf("blockIndex key exceeded max size")
		return
	}

	indexSizeBuffer = make([]byte, 2)
	binary.LittleEndian.PutUint16(indexSizeBuffer, uint16(indexSize))
	data = append(indexSizeBuffer, data...)

	return
}

func (blockIndex *BlockIndex) decode(header *header.Header, data []byte) (err error) {
	blockIndex.bitmask = data[0]
	blockIndex.Type = block.ParseBlockTypeBitmask(blockIndex.bitmask)

	var addressBuffer []byte
	switch header.AddressBytes {
	case 1:
		addressBuffer = []byte{data[1], 0x0}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
	case 2:
		addressBuffer = []byte{data[1], data[2]}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
	case 3:
		addressBuffer = []byte{data[1], data[2], data[3], 0x0}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
	case 4:
		addressBuffer = []byte{data[1], data[2], data[3], data[4]}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
	case 5:
		addressBuffer = []byte{data[1], data[2], data[3], data[4], data[5], 0x0, 0x0, 0x0}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 6:
		addressBuffer = []byte{data[1], data[2], data[3], data[4], data[5], data[6], 0x0, 0x0}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 7:
		addressBuffer = []byte{data[1], data[2], data[3], data[4], data[5], data[6], data[7], 0x0}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 8:
		addressBuffer = []byte{data[1], data[2], data[3], data[4], data[5], data[6], data[7], data[8]}
		blockIndex.address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	}

	blockIndex.BlockSize = binary.LittleEndian.Uint32(data[header.AddressBytes+1 : header.AddressBytes+5])

	if len(data) > header.AddressBytes+5 {
		keyBuffer := data[header.AddressBytes+5:]

		if blockIndex.HasFlag(BitmaskIntKey) {
			var isNegative bool = (keyBuffer[len(keyBuffer)-1] & bitmaskNegative) == bitmaskNegative

			keyBuffer[len(keyBuffer)-1] = keyBuffer[len(keyBuffer)-1] & ^bitmaskNegative

			switch len(keyBuffer) {
			case 4:
				blockIndex.Key = int(binary.LittleEndian.Uint32(keyBuffer))
			case 8:
				blockIndex.Key = int(binary.LittleEndian.Uint64(keyBuffer))
			default:
				err = fmt.Errorf("invalid int key size")
				return
			}

			if isNegative {
				blockIndex.Key = blockIndex.Key.(int) * -1
			}

			return
		}

		blockIndex.Key = string(keyBuffer)
	}

	return
}

type Index struct {
	AllocatedAddresses []block.BlockAddress
	Blocks             map[block.BlockAddress]*BlockIndex
}

func NewIndex() *Index {
	return &Index{
		Blocks: map[block.BlockAddress]*BlockIndex{},
	}
}

func (index *Index) AllocateAddress(blockType block.BlockType) (address block.BlockAddress) {
	address = block.BlockAddress(len(index.AllocatedAddresses) + 1)

	index.AllocatedAddresses = append(index.AllocatedAddresses, address)
	index.Blocks[address] = &BlockIndex{
		address: address,
		bitmask: blockType.Bitmask(),
		Type:    blockType,
	}

	return
}

func (index *Index) LookupBlockIndex(address block.BlockAddress) (blockIndex *BlockIndex, hasBlock bool) {
	blockIndex, hasBlock = index.Blocks[address]

	if hasBlock && blockIndex == nil {
		hasBlock = false
		return
	}

	return
}

func (index *Index) GetKey(address block.BlockAddress) (key interface{}) {
	var (
		blockIndex    *BlockIndex
		hasBlockIndex bool
	)

	if blockIndex, hasBlockIndex = index.LookupBlockIndex(address); hasBlockIndex {
		return blockIndex.GetKey()
	}

	return
}

func (index *Index) SetKey(address block.BlockAddress, key interface{}) (err error) {
	var (
		blockIndex    *BlockIndex
		hasBlockIndex bool
	)

	if blockIndex, hasBlockIndex = index.LookupBlockIndex(address); hasBlockIndex {
		return blockIndex.SetKey(key)
	}

	err = fmt.Errorf("blockIndex with address not found: %X", address)
	return
}

func (index *Index) HasFlag(address block.BlockAddress, flag Flag) bool {
	var (
		blockIndex    *BlockIndex
		hasBlockIndex bool
	)

	if blockIndex, hasBlockIndex = index.LookupBlockIndex(address); hasBlockIndex {
		return blockIndex.HasFlag(flag)
	}

	return false
}

func (index *Index) EnableFlag(address block.BlockAddress, flag Flag) {
	var (
		blockIndex    *BlockIndex
		hasBlockIndex bool
	)

	if blockIndex, hasBlockIndex = index.LookupBlockIndex(address); hasBlockIndex {
		blockIndex.EnableFlag(flag)
	}
}

func (index *Index) DisableFlag(address block.BlockAddress, flag Flag) {
	var (
		blockIndex    *BlockIndex
		hasBlockIndex bool
	)

	if blockIndex, hasBlockIndex = index.LookupBlockIndex(address); hasBlockIndex {
		blockIndex.DisableFlag(flag)
	}
}

func (index *Index) Decode(header *header.Header, data []byte) (cursor uint32, err error) {
	indexSize := binary.LittleEndian.Uint32(data[26:30])

	cursor = 30
	for cursor < 30+indexSize {
		blockIndexSize := uint32(binary.LittleEndian.Uint16(data[cursor : cursor+2]))

		blockIndex := &BlockIndex{}
		if err = blockIndex.decode(header, data[cursor+2:cursor+blockIndexSize+2]); err != nil {
			return
		}

		index.AllocatedAddresses = append(index.AllocatedAddresses, blockIndex.address)
		index.Blocks[blockIndex.address] = blockIndex

		cursor += blockIndexSize + 2
	}

	return
}
