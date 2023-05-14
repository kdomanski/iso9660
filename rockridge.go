package iso9660

/* The following types of Rock Ridge records are being handled in some way:
 * - [ ] PX (RR 4.1.1: POSIX file attributes)
 * - [ ] PN (RR 4.1.2: POSIX device number)
 * - [ ] SL (RR 4.1.3: symbolic link)
 * - [ ] NM (RR 4.1.4: alternate name)
 * - [ ] CL (RR 4.1.5.1: child link)
 * - [ ] PL (RR 4.1.5.2: parent link)
 * - [ ] RE (RR 4.1.5.3: relocated directory)
 * - [ ] TF (RR 4.1.6: time stamp(s) for a file)
 * - [ ] SF (RR 4.1.7: file data in sparse file format)
 */

const (
	RockRidgeIdentifier = "RRIP_1991A"
	RockRidgeVersion    = 1
)

type RockRidgeNameEntry struct {
	Flags byte
	Name  string
}

func suspHasRockRidge(se SystemUseEntrySlice) (bool, error) {
	extensions, err := se.GetExtensionRecords()
	if err != nil {
		return false, err
	}

	for _, entry := range extensions {
		if entry.Identifier == RockRidgeIdentifier && entry.Version == RockRidgeVersion {
			return true, nil
		}
	}

	return false, nil
}

func (s SystemUseEntrySlice) GetRockRidgeName() string {
	var name string

	for _, entry := range s {
		if entry.Type() == "NM" {
			nm := umarshalRockRidgeNameEntry(entry)
			name += nm.Name
		}
	}

	return name
}

func umarshalRockRidgeNameEntry(e SystemUseEntry) *RockRidgeNameEntry {
	// There is a continuation flag in byte 0, but we determine continuation
	// by simply reading all entries.

	return &RockRidgeNameEntry{
		Flags: e.Data()[0],
		Name:  string(e.Data()[1:]),
	}
}
