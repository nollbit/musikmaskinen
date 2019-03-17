package spotify

import (
	"errors"

	"github.com/zmb3/spotify"
)

type (
	trackQueue []spotify.FullTrack

	Queue struct {
		queue        trackQueue
		MaxQueueSize int
	}
)

var (
	ErrorQueueFull  = errors.New("Queue is full")
	ErrorQueueEmpty = errors.New("Queue is empty")
)

func (q *Queue) QueueFull() bool {
	return len(q.queue) >= q.MaxQueueSize
}

func (q *Queue) QueueEmpty() bool {
	return len(q.queue) == 0
}

// add a track to end of the queue
func (q *Queue) QueueAdd(track spotify.FullTrack) error {

	if q.QueueFull() {
		return ErrorQueueFull
	}

	q.queue = append(q.queue, track)

	return nil
}

// remove the last item added to the queue
func (q *Queue) QueueRemove() (*spotify.FullTrack, error) {
	if q.QueueEmpty() {
		return nil, ErrorQueueEmpty
	}

	track := q.queue[len(q.queue)-1]
	q.queue = q.queue[0 : len(q.queue)-1]

	return &track, nil
}

// Next removes and returns the next track to be played
func (q *Queue) Next() (*spotify.FullTrack, error) {
	if q.QueueEmpty() {
		return nil, ErrorQueueEmpty
	}

	track := q.queue[0]
	q.queue = q.queue[1:]

	return &track, nil
}

// Returns the queue as is
func (q *Queue) Get() []spotify.FullTrack {
	return q.queue
}

// NewPlayer creates a new queue. It's not thread safe.
func NewQueue(maxQueueSize int) *Queue {
	return &Queue{
		queue:        make(trackQueue, 0),
		MaxQueueSize: maxQueueSize,
	}
}
