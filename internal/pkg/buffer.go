package pkg

import "bytes"

type WriteCloserBuffer struct {
	bytes.Buffer
}

func (b *WriteCloserBuffer) Close() error {
	return nil
}
