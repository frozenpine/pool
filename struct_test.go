package pool

import (
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestStructSize(t *testing.T) {
	v := TestStruct{}
	t.Log(reflect.TypeOf("").Size(), reflect.TypeOf(1).Size())
	t.Log(reflect.TypeOf(v).Size(), reflect.TypeOf(&v).Size())
	t.Log(unsafe.Sizeof(v), unsafe.Sizeof(&v))
}

func TestStructPool(t *testing.T) {
	_, err := NewStructPool[int]()
	if err == nil {
		t.Fatal("Pool data type assert failed")
	} else {
		t.Log(err)
	}

	pool, err := NewStructPool[TestStruct]()
	if err != nil {
		t.Fatal(err)
	}

	v1 := pool.GetData(false)
	t.Logf("%#v, %+v", v1, unsafe.Pointer(v1))

	v1.Name = "test"
	v1.Age = 123

	pool.ReleaseData(v1)

	var ptr *TestStruct
	for idx := 0; idx < 20; idx++ {
		ptr = pool.GetData(true)
		t.Logf("%#v, %+v", ptr, unsafe.Pointer(ptr))
		ptr.Age = idx
		ptr.Name = strconv.Itoa(idx)
		runtime.GC()
		time.Sleep(time.Second)
	}

	v3 := pool.GetEmptyData(false)
	t.Logf("%#v, %+v", v3, unsafe.Pointer(v3))
}

var testPool = sync.Pool{New: func() any { return new(TestStruct) }}

func poolGet(f bool) *TestStruct {
	v := testPool.Get().(*TestStruct)
	if f {
		runtime.SetFinalizer(v, poolReturn)
	}
	return v
}

func poolReturn(v *TestStruct) {
	if v == nil {
		return
	}

	testPool.Put(v)
}

func BenchmarkFinalizer(b *testing.B) {
	cache := map[string]*TestStruct{}

	b.Run("no set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache["no set"] = poolGet(false)

			poolReturn(cache["no set"])
		}
	})

	b.Run("set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache["set"] = poolGet(true)
		}
	})
}

func BenchmarkStructPool(b *testing.B) {
	pool, _ := NewStructPool[TestStruct]()

	b.Run("origin", func(b *testing.B) {
		cache := map[uintptr]*TestStruct{}
		for i := 0; i < b.N; i++ {
			v := poolGet(false)
			cache[uintptr(unsafe.Pointer(v))] = v
		}

		b.Log(len(cache), b.N)
	})

	b.Run("generic", func(b *testing.B) {
		cache := map[uintptr]*TestStruct{}
		for i := 0; i < b.N; i++ {
			v := pool.GetData(false)
			cache[uintptr(unsafe.Pointer(v))] = v
		}

		b.Log(len(cache), b.N)
	})

	b.Run("generic_gc", func(b *testing.B) {
		cache := map[uintptr]*TestStruct{}
		for i := 0; i < b.N; i++ {
			v := pool.GetData(true)
			cache[uintptr(unsafe.Pointer(v))] = v
		}

		b.Log(len(cache), b.N)
	})

	ch1 := make(chan *TestStruct)
	ch2 := make(chan *TestStruct)
	ch3 := make(chan *TestStruct)
	exit := make(chan struct{})
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		var (
			r1, r2, r3 = true, true, true
			c1, c2, c3 = map[uintptr]*TestStruct{}, map[uintptr]*TestStruct{}, map[uintptr]*TestStruct{}
			n1, n2, n3 = 0, 0, 0
			ptr        *TestStruct
		)

		// gc_count := 0

		for r1 || r2 || r3 {
			select {
			case <-exit:
				close(ch1)
				close(ch2)
				close(ch3)
			case ptr = <-ch1:
				if ptr == nil {
					r1 = false
					continue
				}
				c1[uintptr(unsafe.Pointer(ptr))] = ptr
				n1++
				poolReturn(ptr)
			case ptr = <-ch2:
				if ptr == nil {
					r2 = false
					continue
				}
				c2[uintptr(unsafe.Pointer(ptr))] = ptr
				n2++
				pool.ReleaseData(ptr)
			case ptr = <-ch3:
				if ptr == nil {
					r3 = false
					continue
				}
				c3[uintptr(unsafe.Pointer(ptr))] = ptr
				n3++
				pool.ReleaseData(ptr)
			}
		}

		b.Log("C1", len(c1), n1)
		b.Log("C2", len(c2), n2)
		b.Log("C3", len(c3), n3)
	}()

	b.Run("origin_return", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch1 <- poolGet(false)
		}

		b.Log("C1", b.N)
	})

	b.Run("generic_return", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch2 <- pool.GetData(false)
		}

		b.Log("C2", b.N)
	})

	b.Run("generic_gc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ch3 <- pool.GetData(true)
		}

		b.Log("C3", b.N)
	})

	exit <- struct{}{}
	wg.Wait()
}

func BenchmarkGetPool(b *testing.B) {
	var (
		ptr  = atomic.Pointer[StructPool[TestStruct]]{}
		pool *StructPool[TestStruct]
	)

	b.Run("pointer", func(b *testing.B) {
		if pool == nil {
			pool, _ = NewStructPool[TestStruct]()
		}
	})

	b.Run("atomic", func(b *testing.B) {
		if ptr.Load() == nil {
			p, _ := NewStructPool[TestStruct]()
			ptr.CompareAndSwap(nil, p)
		}
	})
}

func TestReturnReffed(t *testing.T) {
	pool, _ := NewStructPool[TestStruct]()
	cache := []*TestStruct{}

	data := pool.GetData(false)
	cache = append(cache, data)
	pool.ReleaseData(data)

	for idx := 0; idx < 1000000; idx++ {
		data = pool.GetData(false)

		if data == cache[0] {
			t.Log("pool get reffed object")
			break
		}

		pool.ReleaseData(data)
	}
}
