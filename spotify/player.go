package spotify

import (
	"fmt"
	"time"

	"github.com/nollbit/spotify"
	log "github.com/sirupsen/logrus"
)

type (
	State int

	QueuedTrack struct {
		Track spotify.FullTrack
		// time in seconds until this tracks starts playing
		TimeUntilStart int
	}

	// any changes in the queue are signalled heree
	PlayerQueueStatus struct {
		Queue []*QueuedTrack
	}

	// continous updates as tracks are played
	PlayerTrackStatus struct {
		Length    int
		Remaining int
		Done      bool
		Err       error
		Track     *spotify.FullTrack
	}

	Player struct {
		State                 State
		playing               *spotify.FullTrack
		queue                 *Queue
		QueueEvents           chan *PlayerQueueStatus
		TrackEvents           chan *PlayerTrackStatus
		currentTrackRemaining int
		client                *spotify.Client
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
func (p *Player) QueueAdd(track spotify.FullTrack) error {
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

func (p *Player) CurrentlyPlaying() *spotify.FullTrack {
	return p.playing
}

func (p *Player) Skip() error {
	if p.State != StatePlaying {
		return nil
	}

	// simply tell the spotify player to skip the currently playing song
	// polling will detect that we're no longer playing and kick off
	// the next song
	return p.client.Next()
}

func (p *Player) GetQueue() []*QueuedTrack {
	tracks := p.queue.Get()

	//log.Debug("Queue has tracks %v", tracks)

	q := make([]*QueuedTrack, 0, len(tracks))

	remaining := p.currentTrackRemaining

	for _, s := range tracks {
		qs := &QueuedTrack{
			Track:          s,
			TimeUntilStart: remaining,
		}
		q = append(q, qs)

		remaining += (s.Duration / 1000)
	}

	return q
}

func (p *Player) playNextTrackIfNotAlready() {
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

	p.playing = nextTrack

	for {
		log.Debugf("Trying to start track %s", nextTrack.URI)
		err = p.client.PlayOpt(&spotify.PlayOptions{
			URIs: []spotify.URI{nextTrack.URI},
		})
		if err != nil {
			log.WithError(err).Warn("Unable to start playing track")
			time.Sleep(3 * time.Second)
		} else {
			log.WithField("nextTrack", nextTrack.URI).Debug("Started playing track")
			break
		}
	}

	// poll for track play status
	go func() {

		var cp *spotify.CurrentlyPlaying

		// make sure we actually start playing the track before going in to the track loop
		for {
			//log.Debug("Polling for track start")
			cp, err = p.client.PlayerCurrentlyPlaying()
			if err != nil {
				log.WithError(err).Warn("Unable to poll currently playing")
				time.Sleep(1 * time.Second)
				continue
			}

			if !cp.Playing {
				time.Sleep(1 * time.Second)
				continue
			}

			break
		}

		trackLengthMillis := p.playing.Duration
		trackProgressMillis := cp.Progress
		almostDone := false

		// we don't query spotify all the time, so keep track on when we did it last
		latestFullUpdate := time.Now()
		for {
			elapsedSinceFullUpdate := time.Now().Sub(latestFullUpdate)

			if int(elapsedSinceFullUpdate.Seconds()) > 10 || almostDone {

				cp, err = p.client.PlayerCurrentlyPlaying()
				if err != nil {
					log.WithError(err).Warn("Unable to poll currently playing")
					continue
				}
				latestFullUpdate = time.Now()

				trackProgressMillis = cp.Progress
			} else {
				trackProgressMillis = cp.Progress + int(elapsedSinceFullUpdate.Nanoseconds()/1000000)
			}

			currentTrackRemainingMillis := trackLengthMillis - trackProgressMillis
			currentTrackRemaining := currentTrackRemainingMillis / 1000
			p.currentTrackRemaining = currentTrackRemaining

			done := !cp.Playing

			trackStatus := &PlayerTrackStatus{
				Length:    trackLengthMillis / 1000,
				Remaining: currentTrackRemaining,
				Err:       nil,
				Done:      done,
				Track:     p.playing,
			}
			p.TrackEvents <- trackStatus

			if done {
				p.playing = nil
				p.State = StateStopped
				go p.playNextTrackIfNotAlready()
				return
			}

			if currentTrackRemainingMillis < 400 {
				almostDone = true
			}

			time.Sleep(200 * time.Millisecond)
		}
	}()
}

func (p *Player) queueChanged() {
	e := &PlayerQueueStatus{Queue: p.GetQueue()}

	go func() {
		p.QueueEvents <- e
	}()
	p.playNextTrackIfNotAlready()

}

func (p *Player) Close() {
}

// NewPlayer creates a new player. It's not thread safe.
func NewPlayer(client *spotify.Client, maxQueueSize int) (*Player, error) {
	queue := NewQueue(maxQueueSize)

	p := &Player{
		State:       StateStopped,
		playing:     nil,
		queue:       queue,
		TrackEvents: make(chan *PlayerTrackStatus),
		QueueEvents: make(chan *PlayerQueueStatus),
		client:      client,
	}

	return p, nil
}
