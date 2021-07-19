package envelope

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/index"
)

const (
	ObjectBitmaskArray index.Flag = index.BitmaskA
)

func (envelope *Envelope) parseStringMap(input map[string]interface{}) (objectBlock *ObjectBlock, err error) {
	var (
		itemBlock     block.Block
		itemAddresses []block.BlockAddress
	)

	for itemKey, itemValue := range input {
		if itemBlock, err = envelope.ParseBlock(itemValue); err != nil {
			return
		}

		if err = itemBlock.SetKey(itemKey); err != nil {
			return
		}

		itemAddresses = append(itemAddresses, itemBlock.Address())
	}

	if objectBlock, err = envelope.AddObject(itemAddresses); err != nil {
		return
	}

	return
}

func (envelope *Envelope) parseStruct(input reflect.Value) (objectBlock *ObjectBlock, err error) {
	var (
		fieldType     reflect.StructField
		fieldValue    reflect.Value
		itemKey       string
		itemBlock     block.Block
		hasApoTag     bool
		itemAddresses []block.BlockAddress
	)

	for index := 0; index < input.NumField(); index++ {
		fieldType = input.Type().Field(index)

		if itemKey, hasApoTag = fieldType.Tag.Lookup("apo"); !hasApoTag {
			itemKey = fieldType.Name
		}

		fieldValue = input.Field(index)

		if fieldValue.CanInterface() {
			if itemBlock, err = envelope.ParseBlock(fieldValue.Interface()); err != nil {
				return
			}

			if err = itemBlock.SetKey(itemKey); err != nil {
				return
			}

			itemAddresses = append(itemAddresses, itemBlock.Address())
		}
	}

	if objectBlock, err = envelope.AddObject(itemAddresses); err != nil {
		return
	}

	return
}

func (envelope *Envelope) parseArray(input []interface{}) (objectBlock *ObjectBlock, err error) {
	var (
		itemBlock     block.Block
		itemAddresses []block.BlockAddress
	)

	for itemKey, itemValue := range input {
		if itemBlock, err = envelope.ParseBlock(itemValue); err != nil {
			return
		}

		if err = itemBlock.SetKey(itemKey); err != nil {
			return
		}

		itemAddresses = append(itemAddresses, itemBlock.Address())
	}

	if objectBlock, err = envelope.AddObject(itemAddresses); err != nil {
		return
	}

	return
}

func (envelope *Envelope) parseSliceOrArray(input reflect.Value) (objectBlock *ObjectBlock, err error) {
	var (
		itemBlock     block.Block
		itemAddresses []block.BlockAddress
	)

	switch input.Type().Kind() {
	case reflect.Slice, reflect.Array:
		for itemKey := 0; itemKey < input.Len(); itemKey++ {
			itemValue := input.Index(itemKey)

			if !itemValue.CanInterface() {
				continue
			}

			if itemBlock, err = envelope.ParseBlock(itemValue.Interface()); err != nil {
				return
			}

			if err = itemBlock.SetKey(itemKey); err != nil {
				return
			}

			itemAddresses = append(itemAddresses, itemBlock.Address())
		}

		if objectBlock, err = envelope.AddObject(itemAddresses); err != nil {
			return
		}

		objectBlock.SetIsArray(true)
	default:
		err = fmt.Errorf("input is not slice nor array")
	}

	return
}

func (envelope *Envelope) decodeObject(address block.BlockAddress, buffer []byte) (objectBlock *ObjectBlock, err error) {
	objectBlock = &ObjectBlock{
		envelope: envelope,
		address:  address,
		Values:   []block.BlockAddress{},
	}

	if len(buffer)%envelope.Header.AddressBytes != 0 {
		err = fmt.Errorf("invalid address sizes")
		return
	}

	for cursor := 0; cursor < len(buffer); cursor += envelope.Header.AddressBytes {
		var (
			addressBuffer []byte
			address       block.BlockAddress
		)

		switch envelope.Header.AddressBytes {
		case 1:
			addressBuffer = []byte{buffer[cursor], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
		case 2:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1]}
			address = block.BlockAddress(binary.LittleEndian.Uint16(addressBuffer))
		case 3:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
		case 4:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], buffer[cursor+3]}
			address = block.BlockAddress(binary.LittleEndian.Uint32(addressBuffer))
		case 5:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], buffer[cursor+3], buffer[cursor+4], 0x0, 0x0, 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 6:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], buffer[cursor+3], buffer[cursor+4], buffer[cursor+5], 0x0, 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 7:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], buffer[cursor+3], buffer[cursor+4], buffer[cursor+5], buffer[cursor+6], 0x0}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		case 8:
			addressBuffer = []byte{buffer[cursor], buffer[cursor+1], buffer[cursor+2], buffer[cursor+3], buffer[cursor+4], buffer[cursor+5], buffer[cursor+6], buffer[cursor+7]}
			address = block.BlockAddress(binary.LittleEndian.Uint64(addressBuffer))
		}

		objectBlock.Values = append(objectBlock.Values, address)
	}

	return
}

func (envelope *Envelope) AddObject(input interface{}) (objectBlock *ObjectBlock, err error) {
	objectBlock = &ObjectBlock{
		envelope: envelope,
		Values:   []block.BlockAddress{},
	}

	switch value := input.(type) {
	case block.BlockAddress:
		objectBlock.Values = append(objectBlock.Values, value)
	case []block.BlockAddress:
		objectBlock.Values = append(objectBlock.Values, value...)
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	err = envelope.allocateBlock(objectBlock)

	return
}

type ObjectBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Values   []block.BlockAddress
}

func (objectBlock *ObjectBlock) Type() block.BlockType {
	return block.Object
}

func (objectBlock *ObjectBlock) Address() block.BlockAddress {
	return objectBlock.address
}

func (objectBlock *ObjectBlock) SetAddress(address block.BlockAddress) block.Block {
	objectBlock.address = address
	return objectBlock
}

func (objectBlock *ObjectBlock) Key() interface{} {
	return objectBlock.envelope.Index.GetKey(objectBlock.address)
}

func (objectBlock *ObjectBlock) SetKey(key interface{}) error {
	return objectBlock.envelope.Index.SetKey(objectBlock.address, key)
}

func (objectBlock *ObjectBlock) AppendBlock(block block.Block) block.Block {
	objectBlock.Values = append(objectBlock.Values, block.Address())
	return objectBlock
}

func (objectBlock *ObjectBlock) IsRequest() bool {
	return objectBlock.envelope.Index.HasFlag(objectBlock.address, index.BitmaskRequest)
}

func (objectBlock *ObjectBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		objectBlock.envelope.Index.EnableFlag(objectBlock.address, index.BitmaskRequest)
		return objectBlock
	}

	objectBlock.envelope.Index.DisableFlag(objectBlock.address, index.BitmaskRequest)
	return objectBlock
}

func (objectBlock *ObjectBlock) IsResponse() bool {
	return objectBlock.envelope.Index.HasFlag(objectBlock.address, index.BitmaskResponse)
}

func (objectBlock *ObjectBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		objectBlock.envelope.Index.EnableFlag(objectBlock.address, index.BitmaskResponse)
		return objectBlock
	}

	objectBlock.envelope.Index.DisableFlag(objectBlock.address, index.BitmaskResponse)
	return objectBlock
}

func (objectBlock *ObjectBlock) IsArray() bool {
	return objectBlock.envelope.Index.HasFlag(objectBlock.address, ObjectBitmaskArray)
}

func (objectBlock *ObjectBlock) SetIsArray(isArray bool) block.Block {
	if isArray {
		objectBlock.envelope.Index.EnableFlag(objectBlock.address, ObjectBitmaskArray)
		return objectBlock
	}

	objectBlock.envelope.Index.DisableFlag(objectBlock.address, ObjectBitmaskArray)
	return objectBlock
}

func (objectBlock *ObjectBlock) Encode(writer io.Writer) (n int, err error) {
	var (
		data        []byte
		addressData []byte
	)

	if data, err = objectBlock.address.ToBytes(objectBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	for _, address := range objectBlock.Values {
		if addressData, err = address.ToBytes(objectBlock.envelope.Header.AddressBytes); err != nil {
			return
		}

		data = append(data, addressData...)
	}

	return writer.Write(data)
}
