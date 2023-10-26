package readers

import (
	"io"
)

func ReadSeekNopCloser(r io.Reader) ReadSeekerNopClose {
	return ReadSeekerNopClose{r}
}

type ReadSeekerNopClose struct {
	r io.Reader
}

func IsReaderSeekable(r io.Reader) bool {
	switch v := r.(type) {
	case ReadSeekerNopClose:
		return v.IsSeeker()
	case *ReadSeekerNopClose:
		return v.IsSeeker()
	case io.ReadSeeker:
		return true
	default:
		return false
	}
}

func (r ReadSeekerNopClose) Read(p []byte) (int, error) {
	switch t := r.r.(type) {
	case io.Reader:
		return t.Read(p)
	}
	return 0, nil
}

func (r ReadSeekerNopClose) Seek(offset int64, whence int) (int64, error) {
	switch t := r.r.(type) {
	case io.Seeker:
		return t.Seek(offset, whence)
	}
	return int64(0), nil
}

func (r ReadSeekerNopClose) Close() error {
	return nil
}

func (r ReadSeekerNopClose) IsSeeker() bool {
	_, ok := r.r.(io.Seeker)
	return ok
}

func (r ReadSeekerNopClose) HasLen() (int, bool) {
	type lenner interface {
		Len() int
	}

	if lr, ok := r.r.(lenner); ok {
		return lr.Len(), true
	}

	return 0, false
}

func (r ReadSeekerNopClose) GetLen() (int64, error) {
	if l, ok := r.HasLen(); ok {
		return int64(l), nil
	}

	if s, ok := r.r.(io.Seeker); ok {
		return seekerLen(s)
	}

	return -1, nil
}

func seekerLen(s io.Seeker) (int64, error) {
	curOffset, err := s.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	endOffset, err := s.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = s.Seek(curOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return endOffset - curOffset, nil
}
