package ioutil

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	xbzip2 "github.com/dsnet/compress/bzip2"
	"github.com/ulikunitz/xz"
)

type WriteCloserWrapper struct {
	w       io.Writer
	closers Closers
}

func NewWriteCloserWrapper(w io.Writer, closers []io.Closer) io.WriteCloser {
	return &WriteCloserWrapper{
		w:       w,
		closers: closers,
	}
}

func (wrapper *WriteCloserWrapper) Write(buf []byte) (n int, err error) {
	return wrapper.w.Write(buf)
}

func (wrapper *WriteCloserWrapper) Close() error {
	closer, ok := wrapper.w.(io.Closer)
	if ok && (closer != nil) {
		_ = closer.Close()
	}
	_ = wrapper.closers.Close()
	return nil
}

func OpenFileForWriting(path string) (io.WriteCloser, string, error) {
	var closers Closers

	f, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}
	closers = append(closers, f)

	ext := filepath.Ext(path)
	switch ext {
	case ".gz":
		gzipWriter := gzip.NewWriter(f)
		return NewWriteCloserWrapper(gzipWriter, closers), path[:len(path)-len(ext)], nil

	case ".bz2":
		bzip2Writer, err := xbzip2.NewWriter(f, nil)
		if err != nil {
			_ = closers.Close()
			return nil, "", err
		}
		return NewWriteCloserWrapper(bzip2Writer, closers), path[:len(path)-len(ext)], nil

	case ".xz":
		xzWriter, err := xz.NewWriter(f)
		if err != nil {
			_ = closers.Close()
			return nil, "", err
		}
		return NewWriteCloserWrapper(xzWriter, closers), path[:len(path)-len(ext)], nil
	}

	return f, path, nil
}
