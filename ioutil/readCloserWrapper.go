package ioutil

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

type ReadCloserWrapper struct {
	r       io.Reader
	closers Closers
}

func NewReadCloserWrapper(r io.Reader, closers []io.Closer) io.ReadCloser {
	return &ReadCloserWrapper{
		r:       r,
		closers: closers,
	}
}

func (wrapper *ReadCloserWrapper) Read(buf []byte) (n int, err error) {
	return wrapper.r.Read(buf)
}

func (wrapper *ReadCloserWrapper) Close() error {
	closer, ok := wrapper.r.(io.Closer)
	if ok && (closer != nil) {
		_ = closer.Close()
	}
	_ = wrapper.closers.Close()
	return nil
}

func OpenFileForReading(path string) (io.ReadCloser, error) {
	var closers Closers

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	closers = append(closers, f)

	var reader io.Reader
	reader = f
	if strings.HasSuffix(path, ".gz") {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			_ = closers.Close()
			return nil, err
		}
		closers = append(closers, gzipReader)
		reader = gzipReader
	} else if strings.HasSuffix(path, ".bz2") {
		bzip2Reader := bzip2.NewReader(reader)
		reader = bzip2Reader
	} else if strings.HasSuffix(path, ".xz") {
		xzReader, err := xz.NewReader(reader)
		if err != nil {
			_ = closers.Close()
			return nil, err
		}
		reader = xzReader
	}

	return NewReadCloserWrapper(reader, closers), nil
}
