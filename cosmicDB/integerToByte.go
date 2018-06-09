package cosmicDB

import "encoding/binary"

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func uitob(v uint64) []byte {
	y := int(v)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(y))
	return b
}
