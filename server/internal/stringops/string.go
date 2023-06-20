package stringops

import (
	"errors"

	"github.com/eknkc/basex"
)

// EncodeFunctionType specifies all the valid encoding schemes supported in stringops to
// encode any binary input
type EncodeFunctionType uint8

const (
	// EncoderBase36 implements a encoder for binary data resulting in a base36 encoded string.
	// See implementation example at https://play.golang.com/p/KBhdaKW-pl1
	EncoderBase36 EncodeFunctionType = iota + 1
	// Our base36 alphabet will start with letters rather than digits
	base36Encoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Encode will apply over the in byte slice the encoding operation specified by
// encodeType
// It will return an error in case the EncodeFunctionType is not supported, or the
// backing encoder finds an error
func Encode(in []byte, encodeType EncodeFunctionType) (string, error) {
	data := ""
	switch encodeType {
	case EncoderBase36:
		encoder, err := basex.NewEncoding(base36Encoding)
		if err != nil {
			return "", err
		}
		data = encoder.Encode(in)
	default:
		return data, errors.New("undefined encode function")
	}
	return data, nil
}
