package revolver

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestRevolver_write(t *testing.T) {
	b1 := []byte("b1")
	b2 := []byte("b2")

	var isClosed1 bool
	var isClosed2 bool

	wc := New(RevolverRotate(
		// wc1
		writeCloser{
			write: func(p []byte) (int, error) {
				if !bytes.Equal(p, b1) {
					t.Fatalf("wc1: expected %q, got %q", b1, p)
				}
				return len(p), nil
			},
			close: func() error { isClosed1 = true; return nil },
		},

		// wc2
		writeCloser{
			write: func(p []byte) (int, error) {
				if !bytes.Equal(p, b2) {
					t.Fatalf("wc2: expected %q, got %q", b2, p)
				}
				return len(p), nil
			},
			close: func() error { isClosed2 = true; return nil },
		},
	))

	// 1. write (create): use wc1
	_, err := wc.Write(b1)
	if err != nil {
		t.Fatal(err)
	}

	// 2. write (rotate): close wc1, use wc2
	_, err = wc.Write(b2)
	if err != nil {
		t.Fatal(err)
	}
	if !isClosed1 {
		t.Fatal("wc1 not closed")
	}

	// 3. close: close wc2
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !isClosed2 {
		t.Fatal("wc2 not closed")
	}
}

func TestRevolver_Write_multiple_times(t *testing.T) {
	var i int
	a := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	var isClosed bool

	wc := New(NoRotate(writeCloser{
		write: func(p []byte) (int, error) {
			if !bytes.Equal(p, a[i]) {
				t.Fatalf("wc1: expected %q, got %q", a[i], p)
			}
			i++
			return len(p), nil
		},
		close: func() error {
			if isClosed {
				t.Fatal("wc closed")
			}
			isClosed = true
			return nil
		},
	}))

	// 1. write (create)
	_, err := wc.Write(a[i])
	if err != nil {
		t.Fatal(err)
	}

	// 2. write
	_, err = wc.Write(a[i])
	if err != nil {
		t.Fatal(err)
	}

	// 3. write
	_, err = wc.Write(a[i])
	if err != nil {
		t.Fatal(err)
	}

	// 4. close: close wc
	err = wc.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !isClosed {
		t.Fatal("wc not closed")
	}
}

func TestRevolver_Close(t *testing.T) {
	wc := New(NoRotate(discord{}))

	go wc.Write([]byte("abc"))

	err := wc.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRevolver_Close_Twice(t *testing.T) {
	wc := New(NoRotate(discord{}))

	err := wc.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = wc.Close()
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("not closed err: %v", err)
	}
}

func NoRotate(wc io.WriteCloser) Rotator {
	return &noRotate{wc: wc}
}

type noRotate struct {
	b  bool
	wc io.WriteCloser
}

func (r *noRotate) Next() (io.WriteCloser, bool, error) {
	if r.b {
		return nil, false, nil
	}

	r.b = true
	return r.wc, true, nil
}

func RevolverRotate(a ...io.WriteCloser) Rotator {
	return &revolverRotate{a: a}
}

type revolverRotate struct {
	i int
	a []io.WriteCloser
}

func (r *revolverRotate) Next() (io.WriteCloser, bool, error) {
	wc := r.a[r.i]
	r.i++
	return wc, true, nil
}

type writeCloser struct {
	write func(p []byte) (n int, err error)
	close func() error
}

func (wc writeCloser) Write(p []byte) (int, error) { return wc.write(p) }
func (wc writeCloser) Close() error                { return wc.close() }
