package main

import (
	tail "github.com/papertrail/go-tail/follower"
	"io"
)

// AsyncEventReader defines a StartStopper interface for an async read worker
type AsyncEventReader interface {
	Start()
	Stop()
}

// ReaderOption defines the interface for event enrichment
type ReaderOption interface {
	Apply(FailedConnEvent) (FailedConnEvent, error)
}

type fileReader struct {
	parser    EventParser
	respChan  chan FailedConnEvent
	errorChan chan error
	done      chan bool
	options   []ReaderOption
	filename  string
}

// NewFileReader returns an instance of reader
func NewFileReader(filename string, parser EventParser, respChan chan FailedConnEvent, errorChan chan error, options ...ReaderOption) AsyncEventReader {
	done := make(chan bool)
	return fileReader{
		filename:  filename,
		parser:    parser,
		respChan:  respChan,
		errorChan: errorChan,
		done:      done,
		options:   options,
	}
}

func (fr fileReader) Stop() {
	fr.done <- true
}

func (fr fileReader) Start() {
	t, err := tail.New(fr.filename, tail.Config{
		Whence: io.SeekStart,
		Offset: 0,
		Reopen: true,
	})
	if err != nil {
		panic("Error tracking: " + err.Error())
	}

	linesChan := t.Lines()

	for {
		select {
		case s := <-linesChan:
			ev, err := fr.parser.Parse(string(s.Bytes()))
			if err != nil {
				fr.errorChan <- err
				continue // wrong format is not considered an error, I'll handle this better later
			}

			for _, opt := range fr.options {
				ev, err = opt.Apply(ev)
				if err != nil {
					fr.errorChan <- err
					continue
				}
			}

			fr.respChan <- ev
		case <-fr.done:
			close(linesChan)
			return
		}
	}
}
