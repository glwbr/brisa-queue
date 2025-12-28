package memory

import (
	"context"

	"github.com/brisa-queue/queue"
)

type jobState int

const (
	jobPending jobState = iota
	jobInFlight
	jobDone
)

type job struct {
	id      string
	payload []byte
	queue   *Queue
	topic   string
	state   jobState
}

func (j *job) ID() string {
	return j.id
}

func (j *job) Payload() []byte {
	return j.payload
}

func (j *job) Ack(ctx context.Context) error {
	q := j.queue

	q.mu.Lock()
	defer q.mu.Unlock()

	tq := q.topics[j.topic]

	if j.state != jobInFlight {
		return queue.ErrInvalidJobState
	}

	delete(tq.inflight, j.id)
	j.state = jobDone

	return nil
}

func (j *job) Nack(ctx context.Context, reason error) error {
	q := j.queue

	q.mu.Lock()
	defer q.mu.Unlock()

	tq := q.topics[j.topic]

	if j.state != jobInFlight {
		return queue.ErrInvalidJobState
	}

	delete(tq.inflight, j.id)

	j.state = jobPending
	tq.pending = append(tq.pending, j)

	tq.cond.Signal()

	return nil
}
