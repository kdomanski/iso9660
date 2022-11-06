//go:build !integration
// +build !integration

package iso9660

import (
	"io"
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

func TestIncorrectData(t *testing.T) {
	t.Run("volumeDescriptorHeader data too short", func(tt *testing.T) {
		vdh := &volumeDescriptorHeader{}
		err := vdh.UnmarshalBinary([]byte{0, 0, 0, 0, 0})
		assert.ErrorIs(tt, err, io.ErrUnexpectedEOF)
	})

	t.Run("volumeDescriptor data too short", func(tt *testing.T) {
		vd := &volumeDescriptor{}
		err := vd.UnmarshalBinary([]byte{0, 0, 0, 0, 0})
		assert.ErrorIs(tt, err, io.ErrUnexpectedEOF)
	})

	t.Run("volumeDescriptor has invalid volume type", func(tt *testing.T) {
		vd := &volumeDescriptor{}
		data := make([]byte, sectorSize)
		copy(data[0:6], "XCD001")
		err := vd.UnmarshalBinary(data)
		assert.EqualError(tt, err, "unknown volume type 0x58")
	})

	t.Run("volumeDescriptor has invalid identifier", func(tt *testing.T) {
		vd := &volumeDescriptor{}
		data := make([]byte, sectorSize)
		copy(data[0:6], "\x01ABCDE")
		err := vd.UnmarshalBinary(data)
		assert.EqualError(tt, err, "volume descriptor \"ABCDE\" != \"CD001\"")
	})
}

func TestUnmarshalInvalidTimestamp(t *testing.T) {
	t.Run("VolumeDescriptorTimestamp data too short", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte{0, 0, 0, 0, 0})
		assert.ErrorIs(tt, err, io.ErrUnexpectedEOF)
	})

	t.Run("VolumeDescriptorTimestamp year invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("YYYYMMDDHHmmsshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"YYYY\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp month invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("2000MMDDHHmmsshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"MM\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp day invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("200001DDHHmmsshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"DD\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp hour invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("20000102HHmmsshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"HH\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp minutes invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("2000010212mmsshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"mm\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp seconds invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("200001021230sshhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"ss\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp hundredths invalid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("20000102123040hhX"))
		assert.EqualError(tt, err, "strconv.Atoi: parsing \"hh\": invalid syntax")
	})

	t.Run("VolumeDescriptorTimestamp valid", func(tt *testing.T) {
		vdt := &VolumeDescriptorTimestamp{}
		err := vdt.UnmarshalBinary([]byte("2000010212304099\x02"))
		assert.NoError(tt, err)
		assert.Equal(t, VolumeDescriptorTimestamp{
			Year:      2000,
			Month:     01,
			Day:       02,
			Hour:      12,
			Minute:    30,
			Second:    40,
			Hundredth: 99,
			Offset:    2,
		}, *vdt)
	})
}
