package envelope

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"apo/block"
	"apo/index"
)

const (
	IntBitmaskNegative index.Flag = index.BitmaskA
)

func (envelope *Envelope) parseInt(input interface{}) (intBlock *IntBlock, err error) {
	return envelope.AddInt(input)
}

func (envelope *Envelope) decodeInt(address block.BlockAddress, buffer []byte) (intBlock *IntBlock, err error) {
	intBlock = &IntBlock{
		envelope: envelope,
		address:  address,
		Value:    buffer,
	}

	return
}

func (envelope *Envelope) AddInt(input interface{}) (intBlock *IntBlock, err error) {
	var isNegative bool

	intBlock = &IntBlock{
		envelope: envelope,
	}

	/*
		Size		Type		Function
		---			---			---
		2 bytes		uint16		binary.LittleEndian.PutUint16
		4 bytes		uint32		binary.LittleEndian.PutUint32
		8 bytes		uint64		binary.LittleEndian.PutUint64
	*/

	// if envelope.Header.EnableMemoryOptimization {
	// 	/*
	// 		If memory optimization is enabled, then besides choosing bytes
	// 		length soley by variable type, calculate minimum required length
	// 		by value of such variable using:
	// 			bitsCount = math.Ceil(math.Log2(value + 1))
	// 	*/
	// }

	switch value := input.(type) {
	case int:
		isNegative = value < 0

		var unsignedValue uint
		if isNegative {
			unsignedValue = uint(value * -1)
		} else {
			unsignedValue = uint(value)
		}

		intBlock.Value = make([]byte, reflect.TypeOf(unsignedValue).Size())

		if len(intBlock.Value) == 4 {
			binary.LittleEndian.PutUint32(intBlock.Value, uint32(unsignedValue))
			break
		}

		binary.LittleEndian.PutUint64(intBlock.Value, uint64(unsignedValue))
	case int8:
		var unsignedValue uint16

		isNegative = value < 0
		intBlock.Value = make([]byte, 2)

		if isNegative {
			unsignedValue = uint16(value * -1)
		} else {
			unsignedValue = uint16(value)
		}

		binary.LittleEndian.PutUint16(intBlock.Value, unsignedValue)
	case int16:
		var unsignedValue uint16

		isNegative = value < 0
		intBlock.Value = make([]byte, 2)

		if isNegative {
			unsignedValue = uint16(value * -1)
		} else {
			unsignedValue = uint16(value)
		}

		binary.LittleEndian.PutUint16(intBlock.Value, unsignedValue)
	case int32:
		var unsignedValue uint32

		isNegative = value < 0
		intBlock.Value = make([]byte, 4)

		if isNegative {
			unsignedValue = uint32(value * -1)
		} else {
			unsignedValue = uint32(value)
		}

		binary.LittleEndian.PutUint32(intBlock.Value, unsignedValue)
	case int64:
		var unsignedValue uint64

		isNegative = value < 0
		intBlock.Value = make([]byte, 8)

		if isNegative {
			unsignedValue = uint64(value * -1)
		} else {
			unsignedValue = uint64(value)
		}

		binary.LittleEndian.PutUint64(intBlock.Value, unsignedValue)
	case uint:
		intBlock.Value = make([]byte, reflect.TypeOf(value).Size())

		if len(intBlock.Value) == 4 {
			binary.LittleEndian.PutUint32(intBlock.Value, uint32(value))
			break
		}

		binary.LittleEndian.PutUint64(intBlock.Value, uint64(value))
	case uint8:
		intBlock.Value = make([]byte, 2)
		binary.LittleEndian.PutUint16(intBlock.Value, uint16(value))
	case uint16:
		intBlock.Value = make([]byte, 2)
		binary.LittleEndian.PutUint16(intBlock.Value, value)
	case uint32:
		intBlock.Value = make([]byte, 4)
		binary.LittleEndian.PutUint32(intBlock.Value, value)
	case uint64:
		intBlock.Value = make([]byte, 8)
		binary.LittleEndian.PutUint64(intBlock.Value, value)
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	if err = envelope.allocateBlock(intBlock); err != nil {
		return
	}

	intBlock.SetIsNegative(isNegative)

	return
}

type IntBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Value    []byte
}

func (intBlock *IntBlock) Type() block.BlockType {
	return block.Int
}

func (intBlock *IntBlock) Address() block.BlockAddress {
	return intBlock.address
}

func (intBlock *IntBlock) SetAddress(address block.BlockAddress) block.Block {
	intBlock.address = address
	return intBlock
}

func (intBlock *IntBlock) Key() interface{} {
	return intBlock.envelope.Index.GetKey(intBlock.address)
}

func (intBlock *IntBlock) SetKey(key interface{}) error {
	return intBlock.envelope.Index.SetKey(intBlock.address, key)
}

func (intBlock *IntBlock) IsRequest() bool {
	return intBlock.envelope.Index.HasFlag(intBlock.address, index.BitmaskRequest)
}

func (intBlock *IntBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		intBlock.envelope.Index.EnableFlag(intBlock.address, index.BitmaskRequest)
		return intBlock
	}

	intBlock.envelope.Index.DisableFlag(intBlock.address, index.BitmaskRequest)
	return intBlock
}

func (intBlock *IntBlock) IsResponse() bool {
	return intBlock.envelope.Index.HasFlag(intBlock.address, index.BitmaskResponse)
}

func (intBlock *IntBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		intBlock.envelope.Index.EnableFlag(intBlock.address, index.BitmaskResponse)
		return intBlock
	}

	intBlock.envelope.Index.DisableFlag(intBlock.address, index.BitmaskResponse)
	return intBlock
}

func (intBlock *IntBlock) IsNegative() bool {
	return intBlock.envelope.Index.HasFlag(intBlock.address, IntBitmaskNegative)
}

func (intBlock *IntBlock) SetIsNegative(isArray bool) block.Block {
	if isArray {
		intBlock.envelope.Index.EnableFlag(intBlock.address, IntBitmaskNegative)
		return intBlock
	}

	intBlock.envelope.Index.DisableFlag(intBlock.address, IntBitmaskNegative)
	return intBlock
}

func (intBlock *IntBlock) Encode(writer io.Writer) (n int, err error) {
	var addressData []byte

	if addressData, err = intBlock.address.ToBytes(intBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	return writer.Write(append(addressData, intBlock.Value...))
}
