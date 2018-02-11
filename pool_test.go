package gmmpool

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
)

func makeBytes(n int) []byte {
	p := make([]byte, n)
	for i := 0; i < len(p); i++ {
		p[i] = byte(i % 256)
	}
	return p
}

func checksum(p []byte) string {
	return fmt.Sprintf("%x", md5.Sum(p))
}

func TestPoolReadAll(t *testing.T) {
	pool := NewPool(2, 10)

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {
		wg.Add(1)
		go func() {
			for i := 5; i < 15; i++ {
				p := makeBytes(i)
				sum1 := checksum(p)
				b := pool.Get()
				data, err := b.ReadAll(bytes.NewReader(p))
				if err != nil {
					if i > 10 {
						continue
					}
					t.Fatalf("read failed, i:%d, %v", i, err)
				}
				sum2 := checksum(data)
				if sum1 != sum2 {
					t.Fatalf("i:%d, %s != %s", i, sum1, sum2)
				}
				pool.Put(b)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMultiLevelPoolReadAll(t *testing.T) {
	mlPool := NewMultiLevelPool([]PoolOpt{
		PoolOpt{Num: 100, Size: BufSize},
		PoolOpt{Num: 100, Size: BufSize * 2},
	})

	wg := sync.WaitGroup{}
	for j := 0; j < 10; j++ {
		wg.Add(1)
		go func() {
			for i := 5; i < 15; i++ {
				p := makeBytes(i)
				sum1 := checksum(p)
				b := mlPool.Get(i)
				data, err := b.ReadAll(bytes.NewReader(p))
				if err != nil {
					if i > 10 {
						continue
					}
					t.Fatalf("read failed, i:%d, %v", i, err)
				}
				sum2 := checksum(data)
				if sum1 != sum2 {
					t.Fatalf("i:%d, %s != %s", i, sum1, sum2)
				}
				mlPool.Put(b)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

const (
	BufSize = 1024*10 + 1
)

func BenchmarkPoolReadAll(b *testing.B) {
	pool := NewPool(100, BufSize)
	p := makeBytes(BufSize)

	for i := 0; i < b.N; i++ {
		size := rand.Intn(BufSize/2) + BufSize/2
		buf := pool.Get()
		_, err := buf.ReadAll(bytes.NewReader(p[:size]))
		if err != nil {
			b.Fatalf("readall failed:%v", err)
		}
		pool.Put(buf)
	}

}

func BenchmarkStdReadAll(b *testing.B) {
	p := makeBytes(BufSize)

	for i := 0; i < b.N; i++ {

		size := rand.Intn(BufSize/2) + BufSize/2
		_, err := ioutil.ReadAll(bytes.NewReader(p[:size]))
		if err != nil {
			b.Fatalf("readall failed:%v", err)
		}
	}

}

func BenchmarkMultiLevelPool(b *testing.B) {
	mlPool := NewMultiLevelPool([]PoolOpt{
		PoolOpt{Num: 100, Size: BufSize},
		PoolOpt{Num: 100, Size: BufSize * 2},
	})
	p := makeBytes(BufSize)

	for i := 0; i < b.N; i++ {
		size := rand.Intn(BufSize/2) + BufSize/2
		buf := mlPool.Get(size)
		_, err := buf.ReadAll(bytes.NewReader(p[:size]))
		if err != nil {
			b.Fatalf("readall failed:%v", err)
		}
		mlPool.Put(buf)
	}

}
