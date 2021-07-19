package envelope

import (
	"io"

	"github.com/deitas/apo/block"
	"github.com/deitas/apo/index"
)

func (envelope *Envelope) parseNil() (emptyBlock *EmptyBlock, err error) {
	return envelope.AddEmpty()
}

func (envelope *Envelope) decodeEmpty(address block.BlockAddress, buffer []byte) (emptyBlock *EmptyBlock, err error) {
	emptyBlock = &EmptyBlock{
		envelope: envelope,
		address:  address,
	}

	return
}

func (envelope *Envelope) AddEmpty() (emptyBlock *EmptyBlock, err error) {
	emptyBlock = &EmptyBlock{
		envelope: envelope,
	}

	err = envelope.allocateBlock(emptyBlock)

	return emptyBlock, err
}

type EmptyBlock struct {
	envelope *Envelope
	address  block.BlockAddress
}

func (emptyBlock *EmptyBlock) Type() block.BlockType {
	return block.Empty
}

func (emptyBlock *EmptyBlock) Address() block.BlockAddress {
	return emptyBlock.address
}

func (emptyBlock *EmptyBlock) SetAddress(address block.BlockAddress) block.Block {
	emptyBlock.address = address
	return emptyBlock
}

func (emptyBlock *EmptyBlock) Key() interface{} {
	return emptyBlock.envelope.Index.GetKey(emptyBlock.address)
}

func (emptyBlock *EmptyBlock) SetKey(key interface{}) error {
	return emptyBlock.envelope.Index.SetKey(emptyBlock.address, key)
}

func (emptyBlock *EmptyBlock) IsRequest() bool {
	return emptyBlock.envelope.Index.HasFlag(emptyBlock.address, index.BitmaskRequest)
}

func (emptyBlock *EmptyBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		emptyBlock.envelope.Index.EnableFlag(emptyBlock.address, index.BitmaskRequest)
		return emptyBlock
	}

	emptyBlock.envelope.Index.DisableFlag(emptyBlock.address, index.BitmaskRequest)
	return emptyBlock
}

func (emptyBlock *EmptyBlock) IsResponse() bool {
	return emptyBlock.envelope.Index.HasFlag(emptyBlock.address, index.BitmaskResponse)
}

func (emptyBlock *EmptyBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		emptyBlock.envelope.Index.EnableFlag(emptyBlock.address, index.BitmaskResponse)
		return emptyBlock
	}

	emptyBlock.envelope.Index.DisableFlag(emptyBlock.address, index.BitmaskResponse)
	return emptyBlock
}

func (emptyBlock *EmptyBlock) Encode(writer io.Writer) (n int, err error) {
	var addressData []byte

	if addressData, err = emptyBlock.address.ToBytes(emptyBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	return writer.Write(append(addressData, 0x0))
}
