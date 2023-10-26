package readers

import (
	"errors"
	"io"
)

type RepeatableReader struct {
	in io.Reader // Input reader
	i  int64     // current reading index
	b  []byte    // internal cache buffer
}

var _ io.ReadSeeker = (*RepeatableReader)(nil)

func (r *RepeatableReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	cacheLen := int64(len(r.b))
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.i + offset
	case io.SeekEnd:
		abs = cacheLen + offset
	default:
		return 0, errors.New("RepeatableReader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("RepeatableReader.Seek: negative position")
	}
	if abs > cacheLen {
		return offset - (abs - cacheLen), errors.New("RepeatableReader.Seek: offset is unavailable")
	}
	r.i = abs
	return abs, nil
}

func (r *RepeatableReader) Read(b []byte) (n int, err error) {
	cacheLen := int64(len(r.b))
	if r.i == cacheLen {
		n, err = r.in.Read(b)
		if n > 0 {
			r.b = append(r.b, b[:n]...)
		}
	} else {
		n = copy(b, r.b[r.i:])
	}
	r.i += int64(n)
	return n, err
}

func NewRepeatableReader(r io.Reader) *RepeatableReader {
	return &RepeatableReader{in: r}
}

func NewRepeatableReaderSized(r io.Reader, size int) *RepeatableReader {
	return &RepeatableReader{
		in: r,
		b:  make([]byte, 0, size),
	}
}

func NewRepeatableLimitReader(r io.Reader, size int) *RepeatableReader {
	return NewRepeatableReaderSized(io.LimitReader(r, int64(size)), size)
}

func NewRepeatableReaderBuffer(r io.Reader, buf []byte) *RepeatableReader {
	return &RepeatableReader{
		in: r,
		b:  buf[:0],
	}
}
