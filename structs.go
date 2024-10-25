package pool

import (
	"log/slog"
	"reflect"
	"runtime"
	"sync"
)

type StructPool[T any] struct {
	size int
	pool sync.Pool

	converter func(*T) []byte
}

func NewStructPool[T any]() (*StructPool[T], error) {
	data := new(T)
	typ := reflect.TypeOf(data).Elem()

	switch typ.Kind() {
	case reflect.Struct:
	default:
		return nil, ErrInvalidType
	}

	converter, err := Struct2Slice[T]()
	if err != nil {
		return nil, err
	}

	pool := StructPool[T]{
		size: int(typ.Size()),
		pool: sync.Pool{New: func() any {
			return new(T)
		}},
		converter: converter,
	}

	slog.Info(
		"create new pool for struct",
		slog.String("struct", typ.Name()),
		slog.String("pkg_path", typ.PkgPath()),
		slog.Int("size", pool.size),
	)

	return &pool, nil
}

func (p *StructPool[T]) GetData(finalizer bool) *T {
	var data *T

	// 确保获取到结构体
	for {
		v := p.pool.Get()
		if v == nil {
			continue
		}

		var ok bool

		if data, ok = v.(*T); ok {
			break
		}
	}

	if finalizer {
		runtime.SetFinalizer(data, nil)
		runtime.SetFinalizer(data, p.pool.Put)
	}

	return data
}

func (p *StructPool[T]) GetDataWithInit(finalizer bool, fn func(*T)) *T {
	data := p.GetData(finalizer)

	if fn != nil {
		fn(data)
	}

	return data
}

func (p *StructPool[T]) GetEmptyData(finalizer bool) *T {
	v := p.GetData(finalizer)

	data := p.converter(v)

	for idx := range data {
		data[idx] = 0
	}

	return v
}

func (p *StructPool[T]) GetEmptyDataWithInit(finalizer bool, fn func(*T)) *T {
	data := p.GetEmptyData(finalizer)

	if fn != nil {
		fn(data)
	}

	return data
}

func (p *StructPool[T]) PutData(data *T) {
	if data == nil {
		return
	}

	// 确保已入池的对象没有finalizer
	runtime.SetFinalizer(data, nil)
	p.pool.Put(data)
}
