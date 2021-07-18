package envelope

import (
	"io"
	"io/ioutil"
	"os"

	"apo/block"
	"apo/index"
)

func (envelope *Envelope) decodeBinary(address block.BlockAddress, buffer []byte) (binaryBlock *BinaryBlock, err error) {
	binaryBlock = &BinaryBlock{
		envelope: envelope,
		address:  address,
	}

	binaryBlock.Data = buffer

	return
}

func (envelope *Envelope) AddFile(name string) (binaryBlock *BinaryBlock, err error) {
	var (
		file     *os.File
		fileStat os.FileInfo
	)

	if file, err = os.Open(name); err != nil {
		return
	}

	if fileStat, err = file.Stat(); err != nil {
		return
	}

	binaryBlock = &BinaryBlock{
		envelope: envelope,
		Name:     fileStat.Name(),
		Data:     []byte{},
	}

	if binaryBlock.Data, err = ioutil.ReadAll(file); err != nil {
		return
	}

	err = envelope.allocateBlock(binaryBlock)

	return
}

type BinaryBlock struct {
	envelope *Envelope
	address  block.BlockAddress
	MIME     string
	Name     string
	Data     []byte
}

func (binaryBlock *BinaryBlock) Type() block.BlockType {
	return block.Binary
}

func (binaryBlock *BinaryBlock) Address() block.BlockAddress {
	return binaryBlock.address
}

func (binaryBlock *BinaryBlock) SetAddress(address block.BlockAddress) block.Block {
	binaryBlock.address = address
	return binaryBlock
}

func (binaryBlock *BinaryBlock) Key() interface{} {
	return binaryBlock.envelope.Index.GetKey(binaryBlock.address)
}

func (binaryBlock *BinaryBlock) SetKey(key interface{}) error {
	return binaryBlock.envelope.Index.SetKey(binaryBlock.address, key)
}

func (binaryBlock *BinaryBlock) IsRequest() bool {
	return binaryBlock.envelope.Index.HasFlag(binaryBlock.address, index.BitmaskRequest)
}

func (binaryBlock *BinaryBlock) SetIsRequest(isRequest bool) block.Block {
	if isRequest {
		binaryBlock.envelope.Index.EnableFlag(binaryBlock.address, index.BitmaskRequest)
		return binaryBlock
	}

	binaryBlock.envelope.Index.DisableFlag(binaryBlock.address, index.BitmaskRequest)
	return binaryBlock
}

func (binaryBlock *BinaryBlock) IsResponse() bool {
	return binaryBlock.envelope.Index.HasFlag(binaryBlock.address, index.BitmaskResponse)
}

func (binaryBlock *BinaryBlock) SetIsResponse(isResponse bool) block.Block {
	if isResponse {
		binaryBlock.envelope.Index.EnableFlag(binaryBlock.address, index.BitmaskResponse)
		return binaryBlock
	}

	binaryBlock.envelope.Index.DisableFlag(binaryBlock.address, index.BitmaskResponse)
	return binaryBlock
}

func (binaryBlock *BinaryBlock) Encode(writer io.Writer) (n int, err error) {
	var addressData []byte

	if addressData, err = binaryBlock.address.ToBytes(binaryBlock.envelope.Header.AddressBytes); err != nil {
		return
	}

	return writer.Write(append(addressData, binaryBlock.Data...))
}
