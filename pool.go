package gmmpool

import (
	"errors"
	"io"
	"sort"
	"sync"
)

var (
	ErrBufNotEnough = errors.New("buffer not enough")
)

type Buffer struct {
	buf  []byte
	next *Buffer // next free buffer
	off  int     // read at &buf[off], write at &buf[len(buf)]
	size int     // maximum buf size
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.off >= len(b.buf) {
		b.Reset()
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	return
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	need_size := len(p) + len(b.buf)
	if need_size > b.size {
		return 0, ErrBufNotEnough
	}
	m := len(b.buf)
	b.buf = b.buf[:need_size]
	return copy(b.buf[m:need_size], p), nil
}

func (b *Buffer) ReadAll(r io.Reader) ([]byte, error) {
	if b.off >= b.size {
		b.Reset()
	}
	for {
		m, err := r.Read(b.buf[len(b.buf):b.size])
		b.buf = b.buf[:len(b.buf)+m]
		if err == io.EOF {
			break
		}
		if err != nil {
			return b.Bytes(), err
		}
		if m == 0 && len(b.buf) == b.size {
			return b.Bytes(), ErrBufNotEnough
		}
	}
	return b.Bytes(), nil // err is EOF, so return nil explicitly
}

// Pool is a buffer pool.
type Pool struct {
	lock sync.Mutex
	free *Buffer
	max  int
	num  int
	size int
}

// NewPool new a memory buffer pool struct.
func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return
}

// Init init the memory buffer.
func (p *Pool) Init(num, size int) {
	p.init(num, size)
	return
}

// init init the memory buffer.
func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	p.max = num * size
	p.grow()
}

// grow grow the memory buffer size, and update free pointer.
func (p *Pool) grow() {
	var (
		i   int
		b   *Buffer
		bs  []Buffer
		buf []byte
	)
	buf = make([]byte, p.max)
	bs = make([]Buffer, p.num)
	p.free = &bs[0]
	b = p.free
	for i = 1; i < p.num; i++ {
		b.buf = buf[(i-1)*p.size : i*p.size]
		b.size = p.size
		b.next = &bs[i]
		b = b.next
	}
	b.buf = buf[(i-1)*p.size : i*p.size]
	b.size = p.size
	b.next = nil
	return
}

// Get get a free memory buffer.
func (p *Pool) Get() (b *Buffer) {
	p.lock.Lock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}
	p.free = b.next
	p.lock.Unlock()
	b.Reset()
	return
}

// Put put back a memory buffer to free.
func (p *Pool) Put(b *Buffer) {
	p.lock.Lock()
	b.next = p.free
	p.free = b
	p.lock.Unlock()
	return
}

type MultiLevelPool struct {
	pools []*Pool
}

type PoolOpt struct {
	Num  int
	Size int
}

type PoolOptList []PoolOpt

func (a PoolOptList) Len() int           { return len(a) }
func (a PoolOptList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PoolOptList) Less(i, j int) bool { return a[i].Size < a[j].Size }

func NewMultiLevelPool(opts []PoolOpt) *MultiLevelPool {
	optList := PoolOptList(opts)
	sort.Sort(optList)

	pools := make([]*Pool, optList.Len())
	for i, opt := range optList {
		pools[i] = NewPool(opt.Num, opt.Size)
	}
	return &MultiLevelPool{
		pools: pools,
	}
}

func (mlp *MultiLevelPool) Get(size int) (b *Buffer) {
	for _, p := range mlp.pools {
		if p.size >= size {
			return p.Get()
		}
	}

	return nil
}

func (mlp *MultiLevelPool) Put(b *Buffer) {
	for _, p := range mlp.pools {
		if p.size == b.size {
			p.Put(b)
		}
	}
}
