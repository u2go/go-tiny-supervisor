package writer

import (
	"bytes"
	"github.com/u2go/go-tiny-supervisor/lib/fn"
	"io"
)

func NewWriter(name string, writer io.Writer, successFlag string) *Writer {
	w := &Writer{
		Name:      []byte("[" + name + "]: "),
		Writer:    writer,
		SuccessCh: make(chan fn.Empty, 1),
	}
	if successFlag != "" {
		w.SuccessFlag = []byte(successFlag)
	}
	return w
}

type Writer struct {
	Name        []byte
	Writer      io.Writer
	SuccessFlag []byte
	SuccessCh   chan fn.Empty
}

func (w *Writer) Write(p []byte) (n int, err error) {
	// 成功检查
	if w.SuccessFlag != nil {
		if bytes.Contains(p, w.SuccessFlag) {
			w.SuccessCh <- fn.Empty{}
			w.SuccessFlag = nil
		}
	}
	// 写入
	return w.Writer.Write(append(w.Name, p...))
}
