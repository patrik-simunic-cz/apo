package header

import (
	"fmt"
)

type Version struct {
	Major int
	Minor int
}

func (version Version) ToString() string {
	return fmt.Sprintf("%d.%d", version.Major, version.Major)
}

func (version Version) ToByte() byte {
	return (byte(version.Major) << 4) | byte(version.Minor)
}

func (version Version) decode(versionByte byte) {
	version.Major = int(versionByte >> 4)
	version.Minor = int((versionByte << 4) >> 4)
}
