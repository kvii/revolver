package revolver

import (
	"errors"
	"io"
	"sync"
)

func New(r Rotator) io.WriteCloser {
	return &revolver{r: r}
}

type Rotator interface {
	Next() (wc io.WriteCloser, ok bool, err error)
}

var (
	ErrClosed = errors.New("revolver: closed")
)

type revolver struct {
	r        Rotator
	cur      io.WriteCloser
	isClosed bool
	mu       sync.Mutex
}

func (wc *revolver) Write(p []byte) (int, error) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if wc.isClosed {
		return 0, ErrClosed
	}

	cur, ok, err := wc.r.Next()
	if err != nil {
		return 0, err
	}
	if !ok {
		return wc.cur.Write(p)
	}

	if wc.cur != nil {
		wc.cur.Close()
	}

	wc.cur = cur
	return wc.cur.Write(p)
}

func (wc *revolver) Close() error {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if wc.isClosed {
		return ErrClosed
	}
	wc.isClosed = true

	if wc.cur != nil {
		return wc.cur.Close()
	}
	return nil
}
