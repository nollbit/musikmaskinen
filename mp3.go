package main

import (
	"io"
	"os"

	"github.com/hajimehoshi/oto"

	"github.com/hajimehoshi/go-mp3"
)

type SongStatus struct {
	Length    int
	Remaining int
	Done      bool
	Err       error
}

func PlaySong(filename string, statusChan chan *SongStatus, abortChan chan bool) {

	bufSize := 8192

	signalError := func(err error) {
		statusChan <- &SongStatus{Err: err}
	}

	f, err := os.Open(filename)
	if err != nil {
		signalError(err)
		return
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		signalError(err)
		return
	}

	byteRate := (int64(d.SampleRate()) * 2 * 2)

	defer d.Close()

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, bufSize)
	if err != nil {
		signalError(err)
		return
	}
	defer p.Close()

	length := d.Length()
	lengthSeconds := length / byteRate

	// copy manually since we want to keep track of the number of bytes copied
	// so we can translate that into time remaining
	buf := make([]byte, bufSize)
	written := int64(0)

	done := false
	for {
		select {
		case <-abortChan:
			done = true
		default:
		}

		if done {
			break
		}

		nr, er := d.Read(buf)
		if nr > 0 {
			nw, ew := p.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)

				songStatus := &SongStatus{
					Length:    int(lengthSeconds),
					Remaining: int((length - written) / byteRate),
				}
				statusChan <- songStatus
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	if err != nil {
		signalError(err)
	}

	songStatus := &SongStatus{
		Length:    int(lengthSeconds),
		Remaining: 0,
		Done:      true,
	}
	statusChan <- songStatus

	return
}
