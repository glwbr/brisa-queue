package memory

import (
	"context"
)

type job struct {
	id      string
	payload []byte
	queue   *Queue
	topic   string
}

func (j *job) ID() string {
	return j.id
}

func (j *job) Payload() []byte {
	return j.payload
}

func (j *job) Ack(ctx context.Context) error {
	// TODO: Implement
	return nil
}

func (j *job) Nack(ctx context.Context, reason error) error {
	// TODO: Implement
	return nil
}
