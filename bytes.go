package pool

import (
	"sync"
)

const (
	MinBytesSize = 1 << 6
	MaxBytesSize = 1 << 12
)

type BytesPool struct {
	size int
	pool sync.Pool
}

func NewBytesPool(size int) *BytesPool {
	if size <= 0 {
		size = MinBytesSize
	} else if size > MaxBytesSize {
		size = MaxBytesSize
	}

	return &BytesPool{
		size: size,
		pool: sync.Pool{
			New: func() any {
				return make([]byte, size)
			},
		},
	}
}

func (pool *BytesPool) GetSlice() []byte {
	return pool.pool.Get().([]byte)
}

func (pool *BytesPool) GetEmptySlice() []byte {
	bytes := pool.GetSlice()
	for idx := range bytes {
		bytes[idx] = 0
	}
	return bytes
}

func (pool *BytesPool) GetSizedSlice(size int) []byte {
	if size > pool.size || size <= 0 {
		size = pool.size
	}

	return pool.GetSlice()[:size]
}

func (pool *BytesPool) GetEmptySizedSlice(size int) []byte {
	if size > pool.size || size <= 0 {
		size = pool.size
	}

	return pool.GetEmptySlice()[:size]
}

func (pool *BytesPool) PutSlice(data []byte) {
	if cap(data) < pool.size {
		return
	}

	pool.pool.Put(data[:pool.size])
}

var defaulBytestPool = NewBytesPool(MaxBytesSize)

func GetByteSlice() []byte {
	return defaulBytestPool.GetSlice()
}

func GetEmptyByteSlice() []byte {
	return defaulBytestPool.GetEmptySlice()
}

func GetSizedByteSlice(size int) []byte {
	return defaulBytestPool.GetSizedSlice(size)
}

func GetEmptySizedByteSlice(size int) []byte {
	return defaulBytestPool.GetEmptySizedSlice(size)
}

func PutByteSlice(in []byte) {
	defaulBytestPool.PutSlice(in)
}
