package envelope

import (
	"reflect"

	"github.com/deitas/apo/block"
)

func (envelope *Envelope) parsePointer(input interface{}) (block block.Block, err error) {
	var inputValue reflect.Value = reflect.ValueOf(input)

	if inputValue.IsNil() {
		block, err = envelope.ParseBlock(nil)
		return
	}

	inputValue = inputValue.Elem()

	if inputValue.CanInterface() {
		block, err = envelope.ParseBlock(inputValue.Interface())
	}

	return
}
