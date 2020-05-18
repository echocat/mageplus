package support

import (
	"fmt"
	"io"
)

func Close(closers ...io.Closer) {
	if closers != nil {
		for _, closer := range closers {
			if err := closer.Close(); err != nil {
				panic(fmt.Sprintf("cannot close resource: %v", err))
			}
		}
	}
}

func CloseQuietly(closers ...io.Closer) {
	if closers != nil {
		for _, closer := range closers {
			_ = closer.Close()
		}
	}
}
