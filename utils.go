package pool

import (
	"reflect"
	"unsafe"
)

var (
	dummySlice = []byte{}
)

func Struct2Slice[T any]() (func(ptr *T) []byte, error) {
	data := new(T)
	typ := reflect.TypeOf(data).Elem()

	switch typ.Kind() {
	case reflect.Struct:
	default:
		return nil, ErrInvalidType
	}

	return func(ptr *T) []byte {
		if ptr == nil {
			return dummySlice
		}

		data := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(ptr)),
			Len:  int(typ.Size()),
			Cap:  int(typ.Size()),
		}))

		return data
	}, nil
}
