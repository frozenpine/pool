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
	dataSize      int
	finalizerFlag sync.Map
	pool          sync.Pool
}

func NewStructPool[T any]() (*StructPool[T], error) {
	data := new(T)
	typ := reflect.TypeOf(data).Elem()

	switch typ.Kind() {
	case reflect.Struct:
	default:
		return nil, errors.New("generic type must be a struct")
	}

	pool := StructPool[T]{
		dataSize: int(typ.Size()),
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

func (pool *StructPool[T]) GetData(finalizer bool) *T {
	data := pool.pool.Get().(*T)

	if finalizer {
		runtime.SetFinalizer(data, pool.pool.Put)
		pool.finalizerFlag.Store(unsafe.Pointer(data), true)
	}

	return data
}

func (pool *StructPool[T]) GetEmptyData(finalizer bool) *T {
	v := pool.GetData(finalizer)

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
	if data == nil {
		return
	}

	if pool.finalizerFlag.CompareAndDelete(unsafe.Pointer(data), true) {
		runtime.SetFinalizer(data, nil)
	}

	pool.pool.Put(data)
}
