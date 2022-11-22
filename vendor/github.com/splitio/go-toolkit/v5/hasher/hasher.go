package hasher

// Hasher interface
type Hasher interface {
	Hash(data []byte) uint32
}

// Murmur32Hasher is a hash function implementing the murmur3 32 bits algorithm
type Murmur32Hasher struct {
	seed uint32
}

// Hash returns the murmur3 (32 bits) hash of a byte slice.
func (h *Murmur32Hasher) Hash(data []byte) uint32 {
	return sum32WithSeed(data, h.seed)
}

// NewMurmur332Hasher returns a new instance of the Murmur32Hasher
func NewMurmur332Hasher(seed uint32) *Murmur32Hasher {
	return &Murmur32Hasher{seed: seed}
}

// Sum32WithSeed returns the MurmurHash3 sum of data
func Sum32WithSeed(data []byte, seed uint32) uint32 {
	return sum32WithSeed(data, seed)
}
