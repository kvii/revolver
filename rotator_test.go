package revolver

import (
	"io"
	"testing"
	"time"
)

func TestFileRotator(t *testing.T) {
	var i1 int
	a1 := []time.Time{
		time.Date(2006, 01, 02, 10, 0, 0, 0, time.UTC), // day 1
		time.Date(2006, 01, 02, 20, 0, 0, 0, time.UTC), // day 1
		time.Date(2006, 01, 03, 10, 0, 0, 0, time.UTC), // day 2
	}
	nowFunc := func() time.Time {
		t := a1[i1]
		i1++
		return t
	}

	var i2 int
	a2 := []time.Time{
		time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC), // day 1
		time.Date(2006, 01, 03, 0, 0, 0, 0, time.UTC), // day 2
	}
	f := func(date time.Time) (io.WriteCloser, error) {
		if a2[i2] != date {
			t.Fatalf("expected: %v, got: %v", a2[i2], date)
		}
		i2++
		return discord{}, nil
	}

	r := NewFileRotator(RotateOptions{
		Day:    1,
		Create: f,
	})
	r.NowFunc = nowFunc

	// 1. first time: create day 1
	wc, ok, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("!ok")
	}
	if wc == nil {
		t.Fatal("wc == nil")
	}

	// 2. second time: no create
	_, ok, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ok")
	}

	// 3. third time: create day 2
	wc, ok, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("!ok")
	}
	if wc == nil {
		t.Fatal("wc == nil")
	}
}

type discord struct{}

func (discord) Write(p []byte) (int, error) { return len(p), nil }
func (discord) Close() error                { return nil }
