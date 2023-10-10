package revolver

import (
	"io"
	"time"
)

func NewFileRotator(o RotateOptions) *FileRotator {
	return &FileRotator{opts: o}
}

type RotateOptions struct {
	Year   int
	Month  int
	Day    int
	Create CreateFunc
}

type CreateFunc = func(time.Time) (io.WriteCloser, error)

type FileRotator struct {
	opts    RotateOptions
	date    time.Time
	NowFunc func() time.Time
}

func (r *FileRotator) Next() (io.WriteCloser, bool, error) {
	date, ok := r.next()
	if !ok {
		return nil, false, nil
	}

	wc, err := r.opts.Create(date)
	if err != nil {
		return nil, false, err
	}

	r.date = date
	return wc, true, nil
}

func (r *FileRotator) next() (time.Time, bool) {
	today := r.today()

	t := today.AddDate(-r.opts.Year, -r.opts.Month, -r.opts.Day)
	if r.date.After(t) {
		return time.Time{}, false
	}

	return today, true
}

func (r *FileRotator) today() time.Time {
	var t time.Time
	if fn := r.NowFunc; fn != nil {
		t = fn()
	} else {
		t = time.Now()
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
