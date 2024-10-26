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

	Initialize func(*T)
	ToSlice    func(*T) []byte
}

func NewStructPool[T any](initializer func(*T) *T) (*StructPool[T], error) {
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
			v := new(T)

			if initializer != nil {
				initializer(v)
			}

			return v
		}},

		ToSlice: converter,
	}

	if initializer != nil {
		initializer(data)
	}
	pool.ReleaseData(data)

	slog.Debug(
		"created new pool for struct",
		slog.String("struct", typ.Name()),
		slog.String("pkg_path", typ.PkgPath()),
		slog.Int("size", pool.size),
	)

	return &pool, nil
}

func (p *StructPool[T]) GetData(setFinal bool) *T {
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

	if setFinal {
		p.RetainData(data)
	}

	return data
}

func (p *StructPool[T]) GetDataWithInit(setFinal bool, fn func(*T)) *T {
	data := p.GetData(setFinal)

	if fn != nil {
		fn(data)
	}

	return data
}

func (p *StructPool[T]) ClearData(data *T) *T {
	if data == nil {
		return nil
	}

	under := p.ToSlice(data)
	for idx := 0; idx < len(under); idx++ {
		under[idx] = 0
	}

	return data
}

func (p *StructPool[T]) ClearDataWithInit(data *T, fn func(*T)) *T {
	p.ClearData(data)

	if fn != nil {
		fn(data)
	}

	return data
}

func (p *StructPool[T]) GetEmptyData(setFinal bool) *T {
	return p.ClearData(p.GetData(setFinal))
}

func (p *StructPool[T]) GetEmptyDataWithInit(setFinal bool, fn func(*T)) *T {
	return p.ClearDataWithInit(p.GetEmptyData(setFinal), fn)
}

func (p *StructPool[T]) ReleaseData(data *T) {
	// sync.Pool 已进行空值判断
	p.pool.Put(data)
}

func (p *StructPool[T]) RetainData(data *T) {
	if data == nil {
		return
	}

	// 避免对象上已设置Finalizer，先置空
	runtime.SetFinalizer(data, nil)
	runtime.SetFinalizer(data, p.pool.Put)
}

func (p *StructPool[T]) Copy(data *T, setFinal bool) *T {
	if data == nil {
		return nil
	}

	result := p.GetData(false)

	src := p.ToSlice(data)
	dst := p.ToSlice(result)

	copy(dst, src)

	if setFinal {
		p.RetainData(result)
	}

	return result
}

func (p *StructPool[T]) CopyWithInit(data *T, setFinal bool, fn func(*T)) *T {
	result := p.Copy(data, setFinal)
	if result == nil {
		return nil
	}

	if fn != nil {
		fn(result)
	}

	return result
}
