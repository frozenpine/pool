package pool

import (
	"errors"
	"log"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type StructPool[T any] struct {
	dataSize  int
	finalizer bool
	pool      sync.Pool
}

func NewStructPool[T any](finalizer bool) (*StructPool[T], error) {
	data := new(T)
	typ := reflect.TypeOf(data).Elem()

	switch typ.Kind() {
	case reflect.Struct:
	default:
		return nil, errors.New("generic type must be a struct")
	}

	pool := StructPool[T]{
		dataSize:  int(typ.Size()),
		finalizer: finalizer,
		pool: sync.Pool{New: func() any {
			return new(T)
		}},
	}

	log.Printf(
		"Create new pool for struct[%s] with memo size: %d",
		typ.Name(), pool.dataSize,
	)

	return &pool, nil
}

func (pool *StructPool[T]) GetData() *T {
	data := pool.pool.Get().(*T)

	if pool.finalizer {
		runtime.SetFinalizer(data, pool.pool.Put)
	}

	return data
}

func (pool *StructPool[T]) GetEmptyData() *T {
	v := pool.GetData()

	underlying := *(*[]byte)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(v)),
			Len:  pool.dataSize,
			Cap:  pool.dataSize,
		}),
	)

	for idx := range underlying {
		underlying[idx] = 0
	}

	return v
}

func (pool *StructPool[T]) PutData(data *T) {
	if data == nil || pool.finalizer {
		return
	}

	pool.pool.Put(data)
}
