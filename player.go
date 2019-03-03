package main

import "fmt"

type (
	State int

	QueuedSong struct {
		Song *Song
		// time in seconds until this song starts playing
		TimeUntilStart int
	}

	// any changes in the queue are signalled heree
	PlayerQueueStatus struct {
		Queue []*QueuedSong
	}

	// continous updates as songs are played
	PlayerSongStatus struct {
		*SongStatus
		Song *Song
	}

	Player struct {
		State                State
		playing              *Song
		queue                *Queue
		QueueEvents          chan *PlayerQueueStatus
		SongEvents           chan *PlayerSongStatus
		currentSongRemaining int
		skipTrackChan        chan bool
		mp3Player            *Mp3Player
	}
)

const (
	StateStopped State = iota
	StatePlaying State = iota
)

func (p *Player) QueueFull() bool {
	return p.queue.QueueFull()
}

func (p *Player) QueueEmpty() bool {
	return p.queue.QueueEmpty() && p.playing == nil // include current playing song in the "queue"
}

// add a song to end of the queue
func (p *Player) QueueAdd(song *Song) error {
	err := p.queue.QueueAdd(song)
	if err != nil {
		return err
	}

	p.queueChanged()
	return nil
}

// remove the last item added to the queue
func (p *Player) QueueRemove() error {
	if !p.queue.QueueEmpty() {
		_, err := p.queue.QueueRemove()

		if err != nil {
			return err
		}
	} else if p.playing != nil {
		p.Skip()
	}

	p.queueChanged()
	return nil
}

func (p *Player) Skip() {
	if p.State != StatePlaying {
		return
	}

	go func() {
		p.skipTrackChan <- true
	}()
}

func (p *Player) GetQueue() []*QueuedSong {
	songs := p.queue.Get()

	q := make([]*QueuedSong, 0, len(songs))

	remaining := p.currentSongRemaining

	for _, s := range songs {
		qs := &QueuedSong{
			Song:           s,
			TimeUntilStart: remaining,
		}
		q = append(q, qs)

		remaining += s.Length
	}

	return q
}

func (p *Player) playNextSong() {
	if p.QueueEmpty() || p.State == StatePlaying {
		return
	}

	if p.playing != nil {
		fmt.Printf("p.playing = %v\n", p.playing)
		panic("p.playing should not be set when entering playUntilQueueIsEmpty()")
	}

	if p.State != StateStopped {
		panic("state should only be StateStopped when entering playUntilQueueIsEmpty()")
	}

	p.State = StatePlaying

	nextSong, err := p.queue.Next()
	if err != nil {
		panic(err)
	}

	mp3AbortChan := make(chan bool)
	mp3SongStatusChan := make(chan *SongStatus)

	p.playing = nextSong
	go p.mp3Player.PlaySong(p.playing.Path, mp3SongStatusChan, mp3AbortChan)

	go func() {
		// will this ever be cleaned up?
		<-p.skipTrackChan
		mp3AbortChan <- true
	}()

	go func() {
		for {
			ss := <-mp3SongStatusChan

			p.currentSongRemaining = ss.Remaining

			songStatus := &PlayerSongStatus{
				SongStatus: ss,
				Song:       p.playing,
			}
			p.SongEvents <- songStatus

			if ss.Done {
				p.playing = nil
				p.State = StateStopped
				go p.playNextSong()
			}
		}
	}()

}

func (p *Player) queueChanged() {
	e := &PlayerQueueStatus{Queue: p.GetQueue()}

	go func() {
		p.QueueEvents <- e
	}()

	p.playNextSong()

}

func (p *Player) Close() {
	p.mp3Player.Close()
}

// NewPlayer creates a new player. It's not thread safe.
func NewPlayer(maxQueueSize int) (*Player, error) {
	queue := NewQueue(maxQueueSize)

	mp3Player, err := NewMp3Player()

	if err != nil {
		return nil, err
	}

	p := &Player{
		State:         StateStopped,
		playing:       nil,
		queue:         queue,
		SongEvents:    make(chan *PlayerSongStatus),
		QueueEvents:   make(chan *PlayerQueueStatus),
		skipTrackChan: make(chan bool),
		mp3Player:     mp3Player,
	}

	return p, nil
}
