package main

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
		SongStatus,
		Song *Song
	}

	Player struct {
		State                State
		playing              *Song
		queue                *Queue
		QueueEvents          chan *PlayerQueueStatus
		SongEvents           chan *PlayerSongStatus
		currentSongRemaining int
	}
)

const (
	StateStopped State = iota
	StatePlaying State = iota
)

func (p *Player) Play() {
	p.State = StatePlaying
}

func (p *Player) Stop() {
	p.State = StateStopped
}

func (p *Player) loop() {
	//songStatusChan := make(chan *SongStatus)
	for {

	}
}

func (p *Player) QueueFull() bool {
	return p.queue.QueueFull()
}

func (p *Player) QueueEmpty() bool {
	return p.queue.QueueEmpty()
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
func (p *Player) QueueRemove() (*Song, error) {
	return p.queue.QueueRemove()
}

func (p *Player) queueChanged() {
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

	e := &PlayerQueueStatus{Queue: q}

	p.QueueEvents <- e
}

// NewPlayer creates a new player. It's not thread safe.
func NewPlayer(maxQueueSize int) *Player {
	queue := NewQueue(maxQueueSize)

	p := &Player{
		State:       StateStopped,
		queue:       queue,
		SongEvents:  make(chan *PlayerSongStatus),
		QueueEvents: make(chan *PlayerQueueStatus),
	}

	go p.loop()

	return p
}
