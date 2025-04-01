package hashs

import (
	"hash"
	"encoding/hex"
	"hash/fnv"
)


func stringHasher(algorithm hash.Hash, text string) string {
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func FNV64a(s string) string {
	h := fnv.New64a()
	return stringHasher(h, s)
}
