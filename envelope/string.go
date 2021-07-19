package envelope

import (
	"fmt"
	"io"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/index"
)

func (envelope *Envelope) parseString(input interface{}) (stringBlock *StringBlock, err error) {
	return envelope.AddString(input)
}

func (envelope *Envelope) decodeString(address block.BlockAddress, buffer []byte) (stringBlock *StringBlock, err error) {
	stringBlock = &StringBlock{
		envelope: envelope,
		address:  address,
		Value:    buffer,
	}

	return
}

func (envelope *Envelope) AddString(input interface{}) (stringBlock *StringBlock, err error) {
	stringBlock = &StringBlock{
		envelope: envelope,
	}

	switch value := input.(type) {
	case string:
		stringBlock.Value = []byte(value)
	case []byte:
		stringBlock.Value = value
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	err = envelope.allocateBlock(stringBlock)

	return
}

type StringBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Value    []byte
}

func (stringBlock *StringBlock) Type() block.BlockType {
	return block.String
}

func (stringBlock *StringBlock) Address() block.BlockAddress {
	return stringBlock.address
}

func (stringBlock *StringBlock) SetAddress(address block.BlockAddress) block.Block {
	stringBlock.address = address
	return stringBlock
}

func (stringBlock *StringBlock) Key() interface{} {
	return stringBlock.envelope.Index.GetKey(stringBlock.address)
}

func (stringBlock *StringBlock) SetKey(key interface{}) error {
	return stringBlock.envelope.Index.SetKey(stringBlock.address, key)
}

func (stringBlock *StringBlock) IsRequest() bool {
	return stringBlock.envelope.Index.HasFlag(stringBlock.address, index.BitmaskRequest)
}

func (stringBlock *StringBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		stringBlock.envelope.Index.EnableFlag(stringBlock.address, index.BitmaskRequest)
		return stringBlock
	}

	stringBlock.envelope.Index.DisableFlag(stringBlock.address, index.BitmaskRequest)
	return stringBlock
}

func (stringBlock *StringBlock) IsResponse() bool {
	return stringBlock.envelope.Index.HasFlag(stringBlock.address, index.BitmaskResponse)
}

func (stringBlock *StringBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		stringBlock.envelope.Index.EnableFlag(stringBlock.address, index.BitmaskResponse)
		return stringBlock
	}

	stringBlock.envelope.Index.DisableFlag(stringBlock.address, index.BitmaskResponse)
	return stringBlock
}

func (stringBlock *StringBlock) Encode(writer io.Writer) (n int, err error) {
	var addressData []byte

	if addressData, err = stringBlock.address.ToBytes(stringBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	// if stringBlock.envelope.Header.EnableMemoryOptimization {
	// 	// TODO: Compress string
	// }

	return writer.Write(append(addressData, stringBlock.Value...))
}
