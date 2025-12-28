package queue

import "errors"

var (
	// ErrInvalidJobState is returned when Ack or Nack is called on a job that is not in-flight.
	ErrInvalidJobState = errors.New("invalid job state")

	// ErrJobNotFound is returned when a job is not found.
	ErrJobNotFound = errors.New("job not found")

	// ErrQueueClosed is returned when operations are attempted on a closed queue (future).
	ErrQueueClosed = errors.New("queue closed")

	// ErrContextCanceled is returned when dequeue is interrupted.
	ErrContextCanceled = errors.New("context canceled")
)
