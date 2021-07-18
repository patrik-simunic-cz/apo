package envelope

import (
	"fmt"
	"io"

	"apo/block"
	"apo/index"
)

func (envelope *Envelope) parseBoolean(input interface{}) (booleanBlock *BooleanBlock, err error) {
	return envelope.AddBoolean(input)
}

func (envelope *Envelope) decodeBoolean(address block.BlockAddress, buffer []byte) (booleanBlock *BooleanBlock, err error) {
	booleanBlock = &BooleanBlock{
		envelope: envelope,
		address:  address,
	}

	return
}

func (envelope *Envelope) AddBoolean(input interface{}) (booleanBlock *BooleanBlock, err error) {
	booleanBlock = &BooleanBlock{
		envelope: envelope,
	}

	switch value := input.(type) {
	case bool:
		if value {
			booleanBlock.Value = 0x1
			break
		}

		booleanBlock.Value = 0x0
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	err = envelope.allocateBlock(booleanBlock)

	return
}

type BooleanBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Value    byte
}

func (booleanBlock *BooleanBlock) Type() block.BlockType {
	return block.Boolean
}

func (booleanBlock *BooleanBlock) Address() block.BlockAddress {
	return booleanBlock.address
}

func (booleanBlock *BooleanBlock) SetAddress(address block.BlockAddress) block.Block {
	booleanBlock.address = address
	return booleanBlock
}

func (booleanBlock *BooleanBlock) Key() interface{} {
	return booleanBlock.envelope.Index.GetKey(booleanBlock.address)
}

func (booleanBlock *BooleanBlock) SetKey(key interface{}) error {
	return booleanBlock.envelope.Index.SetKey(booleanBlock.address, key)
}

func (booleanBlock *BooleanBlock) IsRequest() bool {
	return booleanBlock.envelope.Index.HasFlag(booleanBlock.address, index.BitmaskRequest)
}

func (booleanBlock *BooleanBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		booleanBlock.envelope.Index.EnableFlag(booleanBlock.address, index.BitmaskRequest)
		return booleanBlock
	}

	booleanBlock.envelope.Index.DisableFlag(booleanBlock.address, index.BitmaskRequest)
	return booleanBlock
}

func (booleanBlock *BooleanBlock) IsResponse() bool {
	return booleanBlock.envelope.Index.HasFlag(booleanBlock.address, index.BitmaskResponse)
}

func (booleanBlock *BooleanBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		booleanBlock.envelope.Index.EnableFlag(booleanBlock.address, index.BitmaskResponse)
		return booleanBlock
	}

	booleanBlock.envelope.Index.DisableFlag(booleanBlock.address, index.BitmaskResponse)
	return booleanBlock
}

func (booleanBlock *BooleanBlock) Encode(writer io.Writer) (n int, err error) {
	var addressData []byte

	if addressData, err = booleanBlock.address.ToBytes(booleanBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	return writer.Write(append(addressData, booleanBlock.Value))
}
