package revolver

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func ExampleNew() {
	var i int
	ts := []time.Time{
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // day 1
		time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), // day 2
		time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), // day 3
	}

	// mock day changes
	nowFunc := func() time.Time {
		t := ts[i]
		i++
		return t
	}

	fr := NewFileRotator(RotateOptions{
		Day: 1,
		Create: func(t time.Time) (io.WriteCloser, error) {
			fmt.Printf("in %s.log: ", t.Format("2006_01_02"))
			return NoOpCloser(os.Stdout), nil
		},
	})
	fr.NowFunc = nowFunc

	// set revolver to logger
	wc := New(fr)
	logger := log.New(wc, "", 0)
	logger.Println("a")
	logger.Println("b")
	logger.Println("c")
	wc.Close()

	//Output:
	// in 2020_01_01.log: a
	// in 2020_01_02.log: b
	// in 2020_01_03.log: c
}

func NoOpCloser(w io.Writer) io.WriteCloser {
	return &noOpCloser{w}
}

type noOpCloser struct {
	io.Writer
}

func (wc *noOpCloser) Close() error { return nil }
