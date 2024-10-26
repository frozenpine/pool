package pool

import (
	"sync"
)

const (
	minBits         = 6
	defaultBits     = 9
	maxBits         = 12
	MinBytesSize    = 1 << minBits     // 64 bytes
	defaultByteSize = 1 << defaultBits // 512 bytes
	MaxBytesSize    = 1 << maxBits     // 4096 bytes
)

type BytesPool struct {
	size int
	pool sync.Pool
}

func judgeSize(size int) int {
	if size < MinBytesSize {
		return MinBytesSize
	}

	if size > MaxBytesSize {
		return MaxBytesSize
	}

	return size
}

func NewBytesPool(size int) *BytesPool {
	p := BytesPool{
		size: calcSize(judgeSize(size)),
	}

	p.pool.New = func() any { return make([]byte, p.size) }
	return &p
}

func (p *BytesPool) GetSlice() []byte {
	return p.pool.Get().([]byte)
}

func (p *BytesPool) GetEmptySlice() []byte {
	bytes := p.GetSlice()
	for idx := range bytes {
		bytes[idx] = 0
	}
	return bytes
}

func (p *BytesPool) GetSizedSlice(size int) []byte {
	if size > p.size || size <= 0 {
		size = p.size
	}

	return p.GetSlice()[:size]
}

func (p *BytesPool) GetEmptySizedSlice(size int) []byte {
	if size > p.size || size <= 0 {
		size = p.size
	}

	return p.GetEmptySlice()[:size]
}

func (p *BytesPool) PutSlice(data []byte) {
	capSize := cap(data)
	if capSize < p.size || capSize > MaxBytesSize {
		return
	}

	p.pool.Put(data[:p.size])
}

func calcSize(size int) int {
	if size <= 0 {
		return 1
	}

	size = size - 1
	size |= size >> 1
	size |= size >> 2
	size |= size >> 4
	size |= size >> 8
	size |= size >> 16

	return size + 1
}

var defaulBytestPool = map[int]*BytesPool{}

func init() {
	for i := minBits; i <= maxBits; i++ {
		size := 1 << i
		defaulBytestPool[size] = NewBytesPool(size)
	}
}

func GetByteSlice() []byte {
	return defaulBytestPool[defaultByteSize].GetSlice()
}

func GetEmptyByteSlice() []byte {
	return defaulBytestPool[defaultByteSize].GetEmptySlice()
}

func GetSizedByteSlice(size int) []byte {
	return defaulBytestPool[calcSize(judgeSize(size))].GetSlice()
}

func GetEmptySizedByteSlice(size int) []byte {
	return defaulBytestPool[calcSize(judgeSize(size))].GetEmptySlice()
}

func PutByteSlice(in []byte) {
	capSize := cap(in)

	if capSize < MinBytesSize || capSize > MaxBytesSize {
		return
	}

	defaulBytestPool[calcSize(capSize)>>1].PutSlice(in)
}
