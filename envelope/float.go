package envelope

import (
	"fmt"
	"io"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/index"
)

func (envelope *Envelope) parseFloat(input interface{}) (floatBlock *FloatBlock, err error) {
	return envelope.AddFloat(input)
}

func (envelope *Envelope) decodeFloat(address block.BlockAddress, buffer []byte) (floatBlock *FloatBlock, err error) {
	floatBlock = &FloatBlock{
		envelope: envelope,
		address:  address,
		Integer:  []byte{0x0},
		Decimal:  []byte{0x0},
	}

	// TODO: do someting with the buffer

	return
}

func (envelope *Envelope) AddFloat(input interface{}) (floatBlock *FloatBlock, err error) {
	floatBlock = &FloatBlock{
		envelope: envelope,
	}

	switch value := input.(type) {
	case float32:
		floatBlock.Integer = []byte(fmt.Sprintf("%f", value))
		floatBlock.Decimal = []byte(fmt.Sprintf("%f", value))
	case float64:
		floatBlock.Integer = []byte(fmt.Sprintf("%f", value))
		floatBlock.Decimal = []byte(fmt.Sprintf("%f", value))
	default:
		err = fmt.Errorf("invalid value type: %T", value)
		return
	}

	err = envelope.allocateBlock(floatBlock)

	return
}

type FloatBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	Integer  []byte
	Decimal  []byte
}

func (floatBlock *FloatBlock) Type() block.BlockType {
	return block.Float
}

func (floatBlock *FloatBlock) Address() block.BlockAddress {
	return floatBlock.address
}

func (floatBlock *FloatBlock) SetAddress(address block.BlockAddress) block.Block {
	floatBlock.address = address
	return floatBlock
}

func (floatBlock *FloatBlock) Key() interface{} {
	return floatBlock.envelope.Index.GetKey(floatBlock.address)
}

func (floatBlock *FloatBlock) SetKey(key interface{}) error {
	return floatBlock.envelope.Index.SetKey(floatBlock.address, key)
}

func (floatBlock *FloatBlock) IsRequest() bool {
	return floatBlock.envelope.Index.HasFlag(floatBlock.address, index.BitmaskRequest)
}

func (floatBlock *FloatBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		floatBlock.envelope.Index.EnableFlag(floatBlock.address, index.BitmaskRequest)
		return floatBlock
	}

	floatBlock.envelope.Index.DisableFlag(floatBlock.address, index.BitmaskRequest)
	return floatBlock
}

func (floatBlock *FloatBlock) IsResponse() bool {
	return floatBlock.envelope.Index.HasFlag(floatBlock.address, index.BitmaskResponse)
}

func (floatBlock *FloatBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		floatBlock.envelope.Index.EnableFlag(floatBlock.address, index.BitmaskResponse)
		return floatBlock
	}

	floatBlock.envelope.Index.DisableFlag(floatBlock.address, index.BitmaskResponse)
	return floatBlock
}

func (floatBlock *FloatBlock) Encode(writer io.Writer) (n int, err error) {
	var data []byte

	if data, err = floatBlock.address.ToBytes(floatBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	data = append(data, floatBlock.Integer...)
	data = append(data, floatBlock.Decimal...)

	return writer.Write(data)
}
