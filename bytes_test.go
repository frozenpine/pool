package pool

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

func TestBytesPool(t *testing.T) {
	pool := NewBytesPool(0)

	v1 := pool.GetSlice()
	t.Log(len(v1), cap(v1), (*reflect.SliceHeader)(unsafe.Pointer(&v1)).Data)

	v2 := pool.GetSlice()
	t.Log(len(v2), cap(v2), (*reflect.SliceHeader)(unsafe.Pointer(&v2)).Data)

	pool.PutSlice(v1)

	v3 := pool.GetSlice()
	t.Log(len(v3), cap(v3), (*reflect.SliceHeader)(unsafe.Pointer(&v3)).Data)

	pool.PutSlice(make([]byte, 0, 5000))

	for idx := 0; idx < 2; idx++ {
		data := pool.GetSlice()

		t.Log(len(data), cap(data), (*reflect.SliceHeader)(unsafe.Pointer(&data)).Data)
	}

	data := GetByteSlice()
	t.Log(len(data), cap(data))
}

func BenchmarkBytesPool(b *testing.B) {
	pool := NewBytesPool(0)

	ch := make(chan []byte, 1)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for d := range ch {
			pool.PutSlice(d)
		}
	}()

	b.Run("pool", func(b1 *testing.B) {
		for i := 0; i < b1.N; i++ {
			ch <- pool.GetSlice()
		}
	})

	b.Run("make", func(b2 *testing.B) {
		for i := 0; i < b2.N; i++ {
			ch <- make([]byte, MinBytesSize)
		}
	})

	b.Run("stack", func(b3 *testing.B) {
		var buff []byte
		for i := 0; i < b3.N; i++ {
			buff = make([]byte, MinBytesSize)
		}

		b.Log(len(buff))
	})

	close(ch)

	wg.Wait()
}
