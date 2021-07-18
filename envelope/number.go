package envelope

import (
	"encoding/json"

	"apo/block"
)

func (envelope *Envelope) parseJSONNumber(input json.Number) (numberBlock block.Block, err error) {
	var (
		valueFloat float64
		valueInt   int64
	)

	if valueFloat, err = input.Float64(); err != nil {
		return
	}

	valueInt = int64(valueFloat)

	if valueFloat == float64(valueInt) {
		if numberBlock, err = envelope.AddInt(valueInt); err != nil {
			return
		}

		return
	}

	if numberBlock, err = envelope.AddFloat(valueFloat); err != nil {
		return
	}

	return
}
