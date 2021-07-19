package apo

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/deitas/apo/envelope"
)

func Parse(input interface{}, options ...envelope.Options) (*envelope.Envelope, error) {
	var (
		err       error
		_envelope *envelope.Envelope = envelope.NewEnvelope(options...)
	)

	if _, err = _envelope.ParseBlock(input); err != nil {
		return _envelope, err
	}

	return _envelope, err
}

func ParseJSON(input []byte, options ...envelope.Options) (*envelope.Envelope, error) {
	var (
		err       error
		decoder   *json.Decoder
		value     interface{}
		_envelope *envelope.Envelope = envelope.NewEnvelope(options...)
	)

	decoder = json.NewDecoder(bytes.NewBuffer(input))
	decoder.UseNumber()

	if err = decoder.Decode(&value); err != nil {
		return _envelope, err
	}

	if _, err = _envelope.ParseBlock(value); err != nil {
		return _envelope, err
	}

	return _envelope, err
}

func Marshal(input interface{}, options ...envelope.Options) ([]byte, error) {
	var (
		err       error
		data      []byte
		_envelope *envelope.Envelope
		buffer    *bytes.Buffer
	)

	if _envelope, err = Parse(input, options...); err != nil {
		return data, err
	}

	buffer = &bytes.Buffer{}
	if err = _envelope.Encode(buffer); err != nil {
		return data, err
	}

	return buffer.Bytes(), err
}

func Read(reader io.Reader, options ...envelope.Options) (*envelope.Envelope, error) {
	var (
		err       error
		_envelope *envelope.Envelope = envelope.NewEnvelope()
	)

	if err = _envelope.Decode(reader); err != nil {
		return nil, err
	}

	return _envelope, err
}

func ReadFile(name string) (*envelope.Envelope, error) {
	var (
		err       error
		file      *os.File
		_envelope *envelope.Envelope = envelope.NewEnvelope()
	)

	if file, err = os.Open(name); err != nil {
		return nil, err
	}

	if err = _envelope.Decode(file); err != nil {
		return nil, err
	}

	return _envelope, err
}
