package envelope

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/index"
)

func (envelope *Envelope) parseAddress(input interface{}) (addressBlock *AddressBlock, err error) {
	return envelope.AddAddress(input)
}

func (envelope *Envelope) decodeAddress(address block.BlockAddress, buffer []byte) (addressBlock *AddressBlock, err error) {
	addressBlock = &AddressBlock{
		envelope: envelope,
		address:  address,
	}

	if len(buffer) != envelope.Header.AddressBytes {
		err = fmt.Errorf("invalid address size")
		return
	}

	var addressBuffer []byte
	switch envelope.Header.AddressBytes {
	case 1:
		addressBuffer = []byte{buffer[0], 0x0}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
	case 2:
		addressBuffer = []byte{buffer[0], buffer[1]}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
	case 3:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], 0x0}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
	case 4:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], buffer[3]}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
	case 5:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], 0x0, 0x0, 0x0}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 6:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5], 0x0, 0x0}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 7:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5], buffer[6], 0x0}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	case 8:
		addressBuffer = []byte{buffer[0], buffer[1], buffer[2], buffer[3], buffer[4], buffer[5], buffer[6], buffer[7]}
		addressBlock.Value = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
	}

	return
}

func (envelope *Envelope) AddAddress(input interface{}) (addressBlock *AddressBlock, err error) {
	addressBlock = &AddressBlock{
		envelope: envelope,
	}

	switch value := input.(type) {
	case block.BlockAddress:
		addressBlock.Value = value
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	err = envelope.allocateBlock(addressBlock)

	return
}

type AddressBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Value    block.BlockAddress
}

func (addressBlock *AddressBlock) Type() block.BlockType {
	return block.Boolean
}

func (addressBlock *AddressBlock) Address() block.BlockAddress {
	return addressBlock.address
}

func (addressBlock *AddressBlock) SetAddress(address block.BlockAddress) block.Block {
	addressBlock.address = address
	return addressBlock
}

func (addressBlock *AddressBlock) Key() interface{} {
	return addressBlock.envelope.Index.GetKey(addressBlock.address)
}

func (addressBlock *AddressBlock) SetKey(key interface{}) error {
	return addressBlock.envelope.Index.SetKey(addressBlock.address, key)
}

func (addressBlock *AddressBlock) IsRequest() bool {
	return addressBlock.envelope.Index.HasFlag(addressBlock.address, index.BitmaskRequest)
}

func (addressBlock *AddressBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		addressBlock.envelope.Index.EnableFlag(addressBlock.address, index.BitmaskRequest)
		return addressBlock
	}

	addressBlock.envelope.Index.DisableFlag(addressBlock.address, index.BitmaskRequest)
	return addressBlock
}

func (addressBlock *AddressBlock) IsResponse() bool {
	return addressBlock.envelope.Index.HasFlag(addressBlock.address, index.BitmaskResponse)
}

func (addressBlock *AddressBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		addressBlock.envelope.Index.EnableFlag(addressBlock.address, index.BitmaskResponse)
		return addressBlock
	}

	addressBlock.envelope.Index.DisableFlag(addressBlock.address, index.BitmaskResponse)
	return addressBlock
}

func (addressBlock *AddressBlock) Encode(writer io.Writer) (n int, err error) {
	var (
		addressData []byte
		valueData   []byte
	)

	if addressData, err = addressBlock.address.ToBytes(addressBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	if valueData, err = addressBlock.Value.ToBytes(addressBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	return writer.Write(append(addressData, valueData...))
}
