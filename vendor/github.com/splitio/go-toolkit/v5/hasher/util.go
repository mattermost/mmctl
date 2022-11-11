package hasher

import (
	"encoding/base64"
	"errors"
	"strconv"
)

// Encode return the encoding of passed string
func Encode(hasher *Murmur32Hasher, key string) (string, error) {
	if hasher == nil {
		return "", errors.New("Hasher could not be nil")
	}

	hashed := int(hasher.Hash([]byte(key)))
	asStr := strconv.Itoa(hashed)
	return base64.StdEncoding.EncodeToString([]byte(asStr)), nil
}
