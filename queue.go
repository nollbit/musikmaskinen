package main

import "errors"

type (
	songQueue []*Song

	Queue struct {
		queue        songQueue
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

// add a song to end of the queue
func (q *Queue) QueueAdd(song *Song) error {

	if q.QueueFull() {
		return ErrorQueueFull
	}

	q.queue = append(q.queue, song)

	return nil
}

// remove the last item added to the queue
func (q *Queue) QueueRemove() (*Song, error) {
	if q.QueueEmpty() {
		return nil, ErrorQueueEmpty
	}

	song := q.queue[len(q.queue)-1]
	q.queue = q.queue[0 : len(q.queue)-1]

	return song, nil
}

// Next removes and returns the next song to be played
func (q *Queue) Next() (*Song, error) {
	if q.QueueEmpty() {
		return nil, ErrorQueueEmpty
	}

	song := q.queue[0]
	q.queue = q.queue[1:]

	return song, nil
}

// Returns the queue as is
func (q *Queue) Get() []*Song {
	return q.queue
}

// NewPlayer creates a new queue. It's not thread safe.
func NewQueue(maxQueueSize int) *Queue {
	return &Queue{
		queue:        make(songQueue, 0),
		MaxQueueSize: maxQueueSize,
	}
}
