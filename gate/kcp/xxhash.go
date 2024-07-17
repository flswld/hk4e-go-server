//go:build !xxhash
// +build !xxhash

package kcp

func byte_check_hash(data []byte) uint32 {
	switch byteCheckMode {
	case 1:
		return 0
	default:
		return 0
	}
}
