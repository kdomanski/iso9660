// +build !integration

package iso9660

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteVolumeDescriptorHeader(t *testing.T) {
	testHeader := volumeDescriptorHeader{
		Type:       volumeTypePrimary,
		Identifier: standardIdentifierBytes,
		Version:    1,
	}

	data, err := testHeader.MarshalBinary()
	assert.NoError(t, err)
	assert.Len(t, data, 7)

	var vdh volumeDescriptorHeader
	err = vdh.UnmarshalBinary(data)
	assert.NoError(t, err)

	assert.Equal(t, testHeader, vdh)
}

func TestReadWriteVolumeDateTime(t *testing.T) {
	exampleTimeDate := VolumeDescriptorTimestamp{
		Year:      2018,
		Month:     06,
		Day:       01,
		Hour:      03,
		Minute:    12,
		Second:    50,
		Hundredth: 7,
		Offset:    8,
	}

	data, err := exampleTimeDate.MarshalBinary()
	assert.NoError(t, err)
	assert.Len(t, data, 17)

	newTS := VolumeDescriptorTimestamp{}
	err = newTS.UnmarshalBinary(data)
	assert.NoError(t, err)

	assert.Equal(t, exampleTimeDate, newTS)
}

func TestReadWriteRecordingTimestamp(t *testing.T) {
	currentTime := time.Now()

	buffer := make([]byte, 7)

	RecordingTimestamp(currentTime).MarshalBinary(buffer)

	var newTimestamp RecordingTimestamp
	err := newTimestamp.UnmarshalBinary(buffer)
	assert.NoError(t, err)

	assert.Equal(t, currentTime.Year(), time.Time(newTimestamp).Year())
	assert.Equal(t, currentTime.Month(), time.Time(newTimestamp).Month())
	assert.Equal(t, currentTime.Day(), time.Time(newTimestamp).Day())
	assert.Equal(t, currentTime.Hour(), time.Time(newTimestamp).Hour())
	assert.Equal(t, currentTime.Minute(), time.Time(newTimestamp).Minute())
	assert.Equal(t, currentTime.Second(), time.Time(newTimestamp).Second())
	_, currentTimeOffset := currentTime.Zone()
	_, newTimestampOffset := time.Time(newTimestamp).Zone()
	assert.Equal(t, currentTimeOffset, newTimestampOffset)
}

func TestReadWriteDirectoryEntry(t *testing.T) {
	f, err := os.Open("fixtures/test.iso")
	assert.NoError(t, err)
	defer f.Close() // nolint:errcheck

	buffer1 := make([]byte, 34)

	_, err = f.ReadAt(buffer1, int64((sectorSize*16)+156))
	assert.NoError(t, err)

	var de DirectoryEntry
	err = de.UnmarshalBinary(buffer1)
	assert.NoError(t, err)

	buffer2, err := de.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, buffer1, buffer2)
}

func TestReadWriteVolumeDescriptorPrimary(t *testing.T) {
	f, err := os.Open("fixtures/test.iso")
	assert.NoError(t, err)
	defer f.Close() // nolint:errcheck

	buffer1 := make([]byte, sectorSize)

	_, err = f.ReadAt(buffer1, int64(sectorSize*16))
	assert.NoError(t, err)

	var vd volumeDescriptor
	err = vd.UnmarshalBinary(buffer1)
	assert.NoError(t, err)

	buffer2, err := vd.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, buffer1, buffer2)
}
