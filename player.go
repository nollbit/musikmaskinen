package main

import "fmt"

type (
	State int

	QueuedTrack struct {
		Track *Track
		// time in seconds until this tracks starts playing
		TimeUntilStart int
	}

	// any changes in the queue are signalled heree
	PlayerQueueStatus struct {
		Queue []*QueuedTrack
	}

	// continous updates as tracks are played
	PlayerTrackStatus struct {
		*TrackStatus
		Track *Track
	}

	Player struct {
		State                 State
		playing               *Track
		queue                 *Queue
		QueueEvents           chan *PlayerQueueStatus
		TrackEvents           chan *PlayerTrackStatus
		currentTrackRemaining int
		skipTrackChan         chan bool
		mp3Player             *Mp3Player
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
	return p.queue.QueueEmpty() && p.playing == nil // include current playing track in the "queue"
}

// add a tracks to end of the queue
func (p *Player) QueueAdd(track *Track) error {
	err := p.queue.QueueAdd(track)
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

func (p *Player) GetQueue() []*QueuedTrack {
	tracks := p.queue.Get()

	q := make([]*QueuedTrack, 0, len(tracks))

	remaining := p.currentTrackRemaining

	for _, s := range tracks {
		qs := &QueuedTrack{
			Track:          s,
			TimeUntilStart: remaining,
		}
		q = append(q, qs)

		remaining += s.Length
	}

	return q
}

func (p *Player) playNextTrack() {
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

	nextTrack, err := p.queue.Next()
	if err != nil {
		panic(err)
	}

	mp3AbortChan := make(chan bool)
	mp3TrackStatusChan := make(chan *TrackStatus)

	p.playing = nextTrack
	go p.mp3Player.PlayTrack(p.playing.Path, mp3TrackStatusChan, mp3AbortChan)

	go func() {
		// will this ever be cleaned up?
		<-p.skipTrackChan
		mp3AbortChan <- true
	}()

	go func() {
		for {
			ss := <-mp3TrackStatusChan

			p.currentTrackRemaining = ss.Remaining

			trackStatus := &PlayerTrackStatus{
				TrackStatus: ss,
				Track:       p.playing,
			}
			p.TrackEvents <- trackStatus

			if ss.Done {
				p.playing = nil
				p.State = StateStopped
				go p.playNextTrack()
			}
		}
	}()

}

func (p *Player) queueChanged() {
	e := &PlayerQueueStatus{Queue: p.GetQueue()}

	go func() {
		p.QueueEvents <- e
	}()

	p.playNextTrack()

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
		TrackEvents:   make(chan *PlayerTrackStatus),
		QueueEvents:   make(chan *PlayerQueueStatus),
		skipTrackChan: make(chan bool),
		mp3Player:     mp3Player,
	}

	return p, nil
}
