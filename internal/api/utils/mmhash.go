package utils

import (
	"github.com/spaolacci/murmur3"
)

func MMHash(buf []byte) uint64 {
	return murmur3.Sum64WithSeed(buf, 0)
}
