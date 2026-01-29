package ioutil

import "io"

type Closers []io.Closer

func (closers Closers) Close() error {
	if closers != nil {
		for i := len(closers) - 1; i >= 0; i-- {
			_ = closers[i].Close()
		}
	}
	return nil
}
