package main

import (
	"io"
	"os"

	"github.com/hajimehoshi/oto"

	"github.com/hajimehoshi/go-mp3"
)

type TrackStatus struct {
	Length    int
	Remaining int
	Done      bool
	Err       error
}

type Mp3Player struct {
	otoPlayer *oto.Player
}

const (
	bufSize = 8192
)

func NewMp3Player() (*Mp3Player, error) {
	p, err := oto.NewPlayer(44100, 2, 2, bufSize)
	if err != nil {
		return nil, err
	}
	//defer p.Close()

	return &Mp3Player{otoPlayer: p}, nil
}

func (m *Mp3Player) Close() {
	m.otoPlayer.Close()
}

func (m *Mp3Player) PlayTrack(filename string, statusChan chan *TrackStatus, abortChan chan bool) {

	signalError := func(err error) {
		statusChan <- &TrackStatus{Err: err}
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
			nw, ew := m.otoPlayer.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)

				trackStatus := &TrackStatus{
					Length:    int(lengthSeconds),
					Remaining: int((length - written) / byteRate),
				}
				statusChan <- trackStatus
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

	trackStatus := &TrackStatus{
		Length:    int(lengthSeconds),
		Remaining: 0,
		Done:      true,
	}
	statusChan <- trackStatus

	return
}
