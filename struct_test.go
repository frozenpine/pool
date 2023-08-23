package pool

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestStructPool(t *testing.T) {
	_, err := NewStructPool[int](true)
	if err == nil {
		t.Fatal("Pool data type assert failed")
	} else {
		t.Log(err)
	}

	pool, err := NewStructPool[TestStruct](true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(unsafe.Sizeof(reflect.StringHeader{}), unsafe.Sizeof(1))

	v1 := pool.GetData()
	t.Logf("%#v", v1)

	v1.Name = "test"
	v1.Age = 123

	pool.PutData(v1)

	for idx := 0; idx < 10; idx++ {
		v := pool.GetData()
		t.Logf("%#v", v)
		// pool.PutData(v)
		runtime.GC()
	}

	v3 := pool.GetEmptyData()
	t.Logf("%#v", v3)
}

func BenchmarkStructPool(b *testing.B) {
	p1 := sync.Pool{New: func() any { return new(TestStruct) }}

	p2, _ := NewStructPool[TestStruct](false)

	var ptr *TestStruct

	b.Run("origin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ptr = p1.Get().(*TestStruct)
		}
	})

	b.Run("generic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ptr = p2.GetData()
		}
	})

	b.Logf("%#v, %+v", ptr, ptr)
	ptr = nil

	ch1 := make(chan *TestStruct, 1)
	ch2 := make(chan *TestStruct, 1)
	exit := make(chan struct{})
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		var (
			c1, c2 = true, true
		)

		for c1 || c2 {
			select {
			case <-exit:
				close(ch1)
				close(ch2)
			case v := <-ch1:
				if v == nil {
					c1 = false
				}
				p1.Put(v)
			case v := <-ch2:
				if v == nil {
					c2 = false
				}
				p2.PutData(v)
			}
		}
	}()

	b.Run("origin_return", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch1 <- p1.Get().(*TestStruct)
		}
	})

	b.Run("generic_return", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch2 <- p2.GetData()
		}
	})

	exit <- struct{}{}
	wg.Wait()
}
