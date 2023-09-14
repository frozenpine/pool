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
	size int
	pool sync.Pool
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
		size: int(typ.Size()),
		pool: sync.Pool{New: func() any {
			return new(T)
		}},
	}

	log.Printf(
		"Create new pool for struct[%s] with memo size: %d",
		typ.Name(), pool.size,
	)

	return &pool, nil
}

func (p *StructPool[T]) GetData(finalizer bool) *T {
	data := p.pool.Get().(*T)

	if finalizer {
		runtime.SetFinalizer(data, p.pool.Put)
	}

	return data
}

func (p *StructPool[T]) GetEmptyData(finalizer bool) *T {
	v := p.GetData(finalizer)

	underlying := *(*[]byte)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(v)),
			Len:  p.size,
			Cap:  p.size,
		}),
	)

	for idx := range underlying {
		underlying[idx] = 0
	}

	return v
}

func (p *StructPool[T]) PutData(data *T) {
	if data == nil {
		return
	}

	// runtime.SetFinalizer(data, nil)

	p.pool.Put(data)
}
