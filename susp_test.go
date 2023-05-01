package iso9660

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopReaderAt struct{}

func (ra *noopReaderAt) ReadAt(p []byte, off int64) (int, error) {
	return 0, fmt.Errorf("reading from noop ReaderAt: %w", io.ErrUnexpectedEOF)
}

type fakeReaderAt struct{}

func (ra *fakeReaderAt) ReadAt(p []byte, off int64) (int, error) {
	for i := 0; i < len(p); i++ {
		p[i] = byte(120)
	}
	return len(p), nil
}

func TestEmptySU(t *testing.T) {
	ra := &noopReaderAt{}

	entries, err := splitSystemUseEntries([]byte{}, ra)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)

	_, err = splitSystemUseEntries([]byte{1, 2, 0}, ra)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)

	entries, err = splitSystemUseEntries(nil, ra)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestSUTooShort(t *testing.T) {
	ra := &noopReaderAt{}

	// payload length is declared to be 200 bytes, but there's only 4 bytes after the header
	_, err := splitSystemUseEntries([]byte{1, 2, 200, 12, 0, 0, 0, 0}, ra)
	assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestSUCEGarbledData(t *testing.T) {
	ra := &noopReaderAt{}

	// ContinuationEntry too short
	suArea := []byte{'C', 'E', 7, 1, 0, 0, 0}
	_, err := splitSystemUseEntries(suArea, ra)
	assert.EqualError(t, err, "unmarshaling ContinuationEntry: invalid ContinuationArea record with length 7 instead of 28")

	// ContinuationEntry has garbled block location
	suArea = []byte{'C', 'E', 28, 1, 100, 0, 0, 0, 0, 0, 0, 99, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	_, err = splitSystemUseEntries(suArea, ra)
	assert.EqualError(t, err, "unmarshaling ContinuationEntry: block location: little-endian and big-endian value mismatch: 100 != 99")

	// ContinuationEntry has garbled offset
	suArea = []byte{'C', 'E', 28, 1, 100, 0, 0, 0, 0, 0, 0, 100, 12, 0, 0, 0, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 0}
	_, err = splitSystemUseEntries(suArea, ra)
	assert.EqualError(t, err, "unmarshaling ContinuationEntry: offset: little-endian and big-endian value mismatch: 12 != 11")

	// ContinuationEntry has garbled length
	suArea = []byte{'C', 'E', 28, 1, 100, 0, 0, 0, 0, 0, 0, 100, 12, 0, 0, 0, 0, 0, 0, 12, 64, 0, 0, 0, 0, 0, 0, 32}
	_, err = splitSystemUseEntries(suArea, ra)
	assert.EqualError(t, err, "unmarshaling ContinuationEntry: length: little-endian and big-endian value mismatch: 64 != 32")

	// Continuation Area read error
	suArea = []byte{'C', 'E', 28, 1, 100, 0, 0, 0, 0, 0, 0, 100, 12, 0, 0, 0, 0, 0, 0, 12, 64, 0, 0, 0, 0, 0, 0, 64}
	_, err = splitSystemUseEntries(suArea, ra)
	assert.EqualError(t, err, "reading Continuation Area: reading from noop ReaderAt: unexpected EOF")

	// Continuation Area points to garbled data
	suArea = []byte{'C', 'E', 28, 1, 100, 0, 0, 0, 0, 0, 0, 100, 12, 0, 0, 0, 0, 0, 0, 12, 64, 0, 0, 0, 0, 0, 0, 64}
	fr := &fakeReaderAt{}
	_, err = splitSystemUseEntries(suArea, fr)
	assert.EqualError(t, err, "splitting Continuation Area: splitting System Use entries: unexpected EOF, expected 120 bytes but have only 64")
}

func TestDecodeInvalidSUSPER(t *testing.T) {
	for _, e := range []SystemUseEntry{
		{'S', 'T', 4, 1}, // not ER
		{'E', 'R', 4, 1}, // definitely too short
		{'E', 'R', 8, 1, 3, 0, 0, 0},
		{'E', 'R', 10, 1, 3, 0, 0, 0, 'F', 'O'},
		{'E', 'R', 14, 1, 3, 4, 0, 0, 'F', 'O', 'O', 'D', 'E', 'S'},
		{'E', 'R', 17, 1, 3, 4, 3, 0, 'F', 'O', 'O', 'D', 'E', 'S', 'C', 'S', 'R'},
	} {
		_, err := ExtensionRecordDecode(e)
		assert.Error(t, err)
	}
}

func TestDecodeValidSUSPER(t *testing.T) {
	for _, e := range []SystemUseEntry{
		{'E', 'R', 8, 1, 0, 0, 0, 0},
		{'E', 'R', 11, 1, 3, 0, 0, 0, 'F', 'O', 'O'},
		{'E', 'R', 15, 1, 3, 4, 0, 0, 'F', 'O', 'O', 'D', 'E', 'S', 'C'},
		{'E', 'R', 18, 1, 3, 4, 3, 0, 'F', 'O', 'O', 'D', 'E', 'S', 'C', 'S', 'R', 'C'},
	} {
		_, err := ExtensionRecordDecode(e)
		assert.NoError(t, err)
	}
}

func TestFailToGetSUSPERs(t *testing.T) {
	slice := SystemUseEntrySlice{{'E', 'R', 4, 1}} // definitely too short
	_, err := slice.GetExtensionRecords()
	assert.Error(t, err)
}
