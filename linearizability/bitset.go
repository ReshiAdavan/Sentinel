package linearizability

// bitset is a type representing a set of bits.
type bitset []uint64

// newBitset creates a new bitset with a specified number of bits.
func newBitset(bits uint) bitset {
	extra := uint(0)
	// Determine if an extra chunk is needed for the remaining bits.
	if bits%64 != 0 {
		extra = 1
	}
	// Calculate total number of 64-bit chunks needed.
	chunks := bits/64 + extra
	return bitset(make([]uint64, chunks))
}

// clone creates a copy of the bitset.
func (b bitset) clone() bitset {
	dataCopy := make([]uint64, len(b))
	copy(dataCopy, b)
	return bitset(dataCopy)
}

// bitsetIndex calculates the index of the 64-bit chunk and the bit position within that chunk.
func bitsetIndex(pos uint) (uint, uint) {
	return pos / 64, pos % 64
}

// set sets the bit at the specified position to 1.
func (b bitset) set(pos uint) bitset {
	major, minor := bitsetIndex(pos)
	b[major] |= (1 << minor)
	return b
}

// clear sets the bit at the specified position to 0.
func (b bitset) clear(pos uint) bitset {
	major, minor := bitsetIndex(pos)
	b[major] &^= (1 << minor)
	return b
}

// get returns true if the bit at the specified position is 1.
func (b bitset) get(pos uint) bool {
	major, minor := bitsetIndex(pos)
	return b[major]&(1<<minor) != 0
}

// popcnt returns the total number of bits set to 1 in the bitset.
func (b bitset) popcnt() uint {
	total := uint(0)
	for _, v := range b {
		// Hamming weight algorithm to count set bits efficiently.
		v = (v & 0x5555555555555555) + ((v & 0xAAAAAAAAAAAAAAAA) >> 1)
		v = (v & 0x3333333333333333) + ((v & 0xCCCCCCCCCCCCCCCC) >> 2)
		v = (v & 0x0F0F0F0F0F0F0F0F) + ((v & 0xF0F0F0F0F0F0F0F0) >> 4)
		v *= 0x0101010101010101
		total += uint((v >> 56) & 0xFF)
	}
	return total
}

// hash computes a hash value for the bitset, useful in hash-based collections.
func (b bitset) hash() uint64 {
	hash := uint64(b.popcnt())
	for _, v := range b {
		hash ^= v
	}
	return hash
}

// equals checks if two bitsets are equal.
func (b bitset) equals(b2 bitset) bool {
	if len(b) != len(b2) {
		return false
	}
	for i := range b {
		if b[i] != b2[i] {
			return false
		}
	}	
	return true
}
