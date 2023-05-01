package iso9660

import (
	"fmt"
	"io"
)

// SUSP-112 4.1
type SystemUseEntry []byte

func (e SystemUseEntry) Length() int {
	return int(e[2])
}

func (e SystemUseEntry) Data() []byte {
	return e[4:]
}

func (e SystemUseEntry) Type() string {
	return string(e[:2])
}

// SUSP-112 5.1
type ContinuationEntry struct {
	blockLocation uint32
	offset        uint32
	lengthOfArea  uint32
}

func umarshalContinuationEntry(e SystemUseEntry) (*ContinuationEntry, error) {
	if e.Length() != 28 {
		return nil, fmt.Errorf("invalid ContinuationArea record with length %d instead of 28", e.Length())
	}

	location, err := UnmarshalUint32LSBMSB(e.Data()[0:8])
	if err != nil {
		return nil, fmt.Errorf("block location: %w", err)
	}
	offset, err := UnmarshalUint32LSBMSB(e.Data()[8:16])
	if err != nil {
		return nil, fmt.Errorf("offset: %w", err)
	}
	length, err := UnmarshalUint32LSBMSB(e.Data()[16:24])
	if err != nil {
		return nil, fmt.Errorf("length: %w", err)
	}

	return &ContinuationEntry{
		blockLocation: location,
		offset:        offset,
		lengthOfArea:  length,
	}, nil
}

const (
	SUEType_ContinuationArea          = "CE"
	SUEType_PaddingField              = "PD"
	SUEType_SharingProtocolIndicator  = "SP"
	SUEType_SharingProtocolTerminator = "ST"
	SUEType_ExtensionsReference       = "ER"
	SUEType_ExtensionSelector         = "ES"
)

func splitSystemUseEntries(data []byte, ra io.ReaderAt) ([]SystemUseEntry, error) {
	output := make([]SystemUseEntry, 0)

	for len(data) > 0 {
		if len(data) < 4 {
			// SUSP-112 4
			// If the remaining allocated space /.../ is less than four bytes long /.../ shall be ignored.
			break
		}

		entryLen := int(data[2])
		if len(data) < entryLen {
			return nil, fmt.Errorf("splitting System Use entries: %w, expected %d bytes but have only %d", io.ErrUnexpectedEOF, entryLen, len(data))
		}

		entry := SystemUseEntry(data[:entryLen])

		if entry.Type() == SUEType_ContinuationArea {
			ce, err := umarshalContinuationEntry(entry)
			if err != nil {
				return output, fmt.Errorf("unmarshaling ContinuationEntry: %w", err)
			}
			continuation := make([]byte, ce.lengthOfArea)
			finalOffset := (ce.blockLocation * sectorSize) + ce.offset
			if _, err := ra.ReadAt(continuation, int64(finalOffset)); err != nil {
				return output, fmt.Errorf("reading Continuation Area: %w", err)
			}

			continuedEntries, err := splitSystemUseEntries(continuation, ra)
			if err != nil {
				return output, fmt.Errorf("splitting Continuation Area: %w", err)
			}
			output = append(output, continuedEntries...)
		} else {
			output = append(output, entry)
		}

		data = data[entryLen:]
	}

	return output, nil
}