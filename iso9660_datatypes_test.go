package iso9660

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPadding(t *testing.T) {
	out := MarshalString("foobar", 16)
	assert.Equal(t, []byte("foobar          "), out)

	out = MarshalString("foobar", 4)
	assert.Equal(t, []byte("foob"), out)
}

func TestUnmarshallInt32LSBMSB(t *testing.T) {
	// data OK
	number, err := UnmarshalInt32LSBMSB([]byte{0x00, 0x2D, 0x31, 0x01, 0x01, 0x31, 0x2D, 0x00})
	assert.NoError(t, err)
	assert.Equal(t, int32(20000000), number)

	// data too short
	_, err = UnmarshalInt32LSBMSB([]byte{0x01, 0x31, 0x2D, 0x00, 0x00, 0x2D, 0x31})
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	// endian versions mismatch
	_, err = UnmarshalInt32LSBMSB([]byte{0x01, 0x31, 0x2D, 0x00, 0x00, 0x00, 0x00, 0x00})
	assert.Error(t, err)
}

func TestUnmarshallInt16LSBMSB(t *testing.T) {
	// data OK
	number, err := UnmarshalInt16LSBMSB([]byte{0x20, 0x4E, 0x4E, 0x20})
	assert.NoError(t, err)
	assert.Equal(t, int16(20000), number)

	// data too short
	_, err = UnmarshalInt16LSBMSB([]byte{0x20, 0x4E, 0x4E})
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	// endian versions mismatch
	_, err = UnmarshalInt16LSBMSB([]byte{0x20, 0x4E, 0x00, 0x00})
	assert.Error(t, err)
}
