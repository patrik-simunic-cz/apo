package envelope

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"time"

	"apo/block"
	"apo/header"
	"apo/index"
)

type Options struct {
	IsExtension              bool
	EnableMemoryOptimization bool
}

type Envelope struct {
	Header *header.Header
	Index  *index.Index
	Blocks block.Blocks
}

func (envelope Envelope) allocateBlock(block block.Block) (err error) {
	address := envelope.Index.AllocateAddress(block.Type())

	block.SetAddress(address)

	envelope.Blocks.Set(address, block)

	blocksCount := len(envelope.Index.AllocatedAddresses)
	if blocksCount < 256 { // 2^8
		envelope.Header.AddressBytes = 1
	} else if blocksCount < 65536 { // 2^16
		envelope.Header.AddressBytes = 2
	} else if blocksCount < 16777216 { // 2^24
		envelope.Header.AddressBytes = 3
	} else if blocksCount < 4294967296 { // 2^32
		envelope.Header.AddressBytes = 4
	} else if blocksCount < 1099511627776 { // 2^40
		envelope.Header.AddressBytes = 5
	} else if blocksCount < 281474976710656 { // 2^48
		envelope.Header.AddressBytes = 6
	} else if blocksCount < 72057594037927936 { // 2^54
		envelope.Header.AddressBytes = 7
	} else if blocksCount <= 9223372036854775807 { // 2^63 - 1
		envelope.Header.AddressBytes = 8
	} else {
		err = fmt.Errorf("maximum address size exceeded")
	}

	return
}

func NewEnvelope(options ...Options) (envelope *Envelope) {
	envelope = &Envelope{
		Header: header.NewHeader(),
		Index:  index.NewIndex(),
		Blocks: block.Blocks{},
	}

	if len(options) > 0 {
		envelope.Header.IsExtension = options[0].IsExtension
		envelope.Header.EnableMemoryOptimization = options[0].EnableMemoryOptimization
	}

	return
}

func (envelope *Envelope) ParseBlock(input interface{}) (parsedBlock block.Block, err error) {
	switch value := input.(type) {
	case nil:
		parsedBlock, err = envelope.parseNil()
	case block.BlockAddress:
		parsedBlock, err = envelope.parseAddress(value)
	case string, []byte:
		parsedBlock, err = envelope.parseString(value)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		parsedBlock, err = envelope.parseInt(value)
	case float32, float64:
		parsedBlock, err = envelope.parseFloat(value)
	case bool:
		parsedBlock, err = envelope.parseBoolean(value)
	case map[string]interface{}:
		parsedBlock, err = envelope.parseStringMap(value)
	case []interface{}:
		parsedBlock, err = envelope.parseArray(value)
	case json.Number:
		parsedBlock, err = envelope.parseJSONNumber(value)
	case time.Time:
		// TODO: datetime block
		parsedBlock, err = envelope.AddEmpty()
	default:
		reflectValue := reflect.ValueOf(value)

		switch reflectValue.Kind() {
		case reflect.Struct:
			parsedBlock, err = envelope.parseStruct(reflectValue)
		case reflect.Slice, reflect.Array:
			parsedBlock, err = envelope.parseSliceOrArray(reflectValue)
		case reflect.Ptr:
			parsedBlock, err = envelope.parsePointer(value)
		default:
			err = fmt.Errorf("unknown type of %+v", reflectValue.Kind())
		}
	}

	return
}

type TraverseHandler = func(block.Block, *index.BlockIndex)

func (envelope *Envelope) TraverseBlocks(condition func(block.Block, *index.BlockIndex) bool, handler TraverseHandler) {
	for _, allocatedAddress := range envelope.Index.AllocatedAddresses {
		var (
			hasBlockIndex bool
			blockIndex    *index.BlockIndex
			blockAddress  block.BlockAddress = block.BlockAddress(allocatedAddress)
		)

		if blockIndex, hasBlockIndex = envelope.Index.LookupBlockIndex(blockAddress); !hasBlockIndex {
			continue
		}

		block := envelope.Blocks.Get(blockAddress)
		if condition(block, blockIndex) {
			handler(block, blockIndex)
		}
	}
}

func (envelope *Envelope) TraverseAllBlocks(handler TraverseHandler) {
	envelope.TraverseBlocks(func(_ block.Block, _ *index.BlockIndex) bool {
		return true
	}, handler)
}

func (envelope *Envelope) TraverseBlockType(blockType block.BlockType, handler TraverseHandler) {
	envelope.TraverseBlocks(func(_ block.Block, blockIndex *index.BlockIndex) bool {
		return blockIndex.Type == blockType
	}, handler)
}

func (envelope *Envelope) TraverseObjects(handler TraverseHandler) {
	envelope.TraverseBlockType(block.Object, func(block block.Block, blockIndex *index.BlockIndex) {
		handler(block.(*ObjectBlock), blockIndex)
	})
}

func (envelope *Envelope) TraverseBinaries(handler TraverseHandler) {
	envelope.TraverseBlockType(block.Object, func(block block.Block, blockIndex *index.BlockIndex) {
		handler(block.(*BinaryBlock), blockIndex)
	})
}

func (envelope *Envelope) TraverseRequests(handler TraverseHandler) {
	envelope.TraverseBlocks(func(_ block.Block, blockIndex *index.BlockIndex) bool {
		return blockIndex.HasFlag(index.BitmaskRequest)
	}, handler)
}

func (envelope *Envelope) TraverseResponses(handler TraverseHandler) {
	envelope.TraverseBlocks(func(_ block.Block, blockIndex *index.BlockIndex) bool {
		return blockIndex.HasFlag(index.BitmaskResponse)
	}, handler)
}

func (envelope *Envelope) Encode(writer io.Writer) (err error) {
	var (
		indexBufferSize       int
		indexBufferSizeBuffer []byte
		indexBuffer           *bytes.Buffer = &bytes.Buffer{}
		blocksBuffer          *bytes.Buffer = &bytes.Buffer{}
	)

	for allocatedAddress := range envelope.Index.AllocatedAddresses {
		var (
			blockAddress     block.BlockAddress = block.BlockAddress(allocatedAddress)
			hasBlockIndex    bool
			blockIndex       *index.BlockIndex
			blockIndexBuffer []byte
			hasBlock         bool
			block            block.Block
			blockSize        int
		)

		if blockIndex, hasBlockIndex = envelope.Index.LookupBlockIndex(blockAddress); !hasBlockIndex {
			continue
		}

		if block, hasBlock = envelope.Blocks[blockAddress]; !hasBlock {
			continue
		}

		if blockSize, err = block.Encode(blocksBuffer); err != nil {
			return
		}

		if blockSize >= 4294967296 {
			err = fmt.Errorf("block exceeded maximum size of 4 GiB")
			return
		}

		blockIndex.BlockSize = uint32(blockSize)
		if blockIndexBuffer, err = blockIndex.ToBytes(envelope.Header); err != nil {
			return
		}

		if _, err = indexBuffer.Write(blockIndexBuffer); err != nil {
			return
		}
	}

	indexBufferSize = indexBuffer.Len()
	if indexBufferSize >= 4294967296 {
		err = fmt.Errorf("APO index exceeded maximum size of 4 GiB")
		return
	}

	indexBufferSizeBuffer = make([]byte, 4)
	binary.LittleEndian.PutUint32(indexBufferSizeBuffer, uint32(indexBufferSize))

	envelope.Header.IndexChecksum = header.CalcuateChecksum(indexBuffer)
	envelope.Header.BlocksChecksum = header.CalcuateChecksum(blocksBuffer)

	if _, err = envelope.Header.Encode(writer); err != nil {
		return
	}

	if _, err = writer.Write(indexBufferSizeBuffer); err != nil {
		return
	}

	if _, err = indexBuffer.WriteTo(writer); err != nil {
		return
	}

	if _, err = blocksBuffer.WriteTo(writer); err != nil {
		return
	}

	return
}

func (envelope *Envelope) Decode(reader io.Reader) (err error) {
	var (
		data         []byte
		dataSize     int
		blocksOffset uint32
		cursor       int
	)

	if data, err = ioutil.ReadAll(reader); err != nil {
		return
	}

	dataSize = len(data)

	if err = envelope.Header.Decode(data); err != nil {
		return
	}

	if blocksOffset, err = envelope.Index.Decode(envelope.Header, data); err != nil {
		return
	}

	cursor = int(blocksOffset)

	for cursor < dataSize {
		var (
			addressBuffer []byte
			address       block.BlockAddress
			hasBlockIndex bool
			blockIndex    *index.BlockIndex
			blockBuffer   []byte
			decodedBlock  block.Block
		)

		switch envelope.Header.AddressBytes {
		case 1:
			addressBuffer = []byte{data[cursor], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
		case 2:
			addressBuffer = []byte{data[cursor], data[cursor+1]}
			address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
		case 3:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
		case 4:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], data[cursor+3]}
			address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
		case 5:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], data[cursor+3], data[cursor+4], 0x0, 0x0, 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 6:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], data[cursor+3], data[cursor+4], data[cursor+5], 0x0, 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 7:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], data[cursor+3], data[cursor+4], data[cursor+5], data[cursor+6], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 8:
			addressBuffer = []byte{data[cursor], data[cursor+1], data[cursor+2], data[cursor+3], data[cursor+4], data[cursor+5], data[cursor+6], data[cursor+7]}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		}

		if blockIndex, hasBlockIndex = envelope.Index.LookupBlockIndex(address); !hasBlockIndex {
			err = fmt.Errorf("block index with address %d does not exist", address)
			return
		}

		blockBuffer = data[cursor+envelope.Header.AddressBytes : cursor+int(blockIndex.BlockSize)]

		switch blockIndex.Type {
		case block.Address:
			decodedBlock, err = envelope.decodeAddress(address, blockBuffer)
		case block.Empty:
			decodedBlock, err = envelope.decodeEmpty(address, blockBuffer)
		case block.Object:
			decodedBlock, err = envelope.decodeObject(address, blockBuffer)
		case block.Binary:
			decodedBlock, err = envelope.decodeBinary(address, blockBuffer)
		case block.Boolean:
			decodedBlock, err = envelope.decodeBoolean(address, blockBuffer)
		case block.String:
			decodedBlock, err = envelope.decodeString(address, blockBuffer)
		case block.Int:
			decodedBlock, err = envelope.decodeInt(address, blockBuffer)
		case block.Float:
			decodedBlock, err = envelope.decodeFloat(address, blockBuffer)
		case block.DateTime:
			// TODO: datetime block
			decodedBlock, err = envelope.decodeEmpty(address, blockBuffer)
		default:
			err = fmt.Errorf("invalid block type: %d", blockIndex.Type)
			return
		}

		if err != nil {
			return
		}

		envelope.Blocks.Set(address, decodedBlock)

		cursor += int(blockIndex.BlockSize)
	}

	return
}
