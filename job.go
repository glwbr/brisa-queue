package queue

import "context"

type Job interface {
	// ID returns the unique identifier of the job.
	ID() string

	// Payload returns the raw byte content of the job which is agnostic to the serialization format (JSON, Gob, Protobuf).
	Payload() []byte

	// Ack signals that the job has been successfully processed and can be removed from the queue.
	Ack(ctx context.Context) error

	// Nack signals that processing failed. The implementation determines whether to retry, dead-letter, or discard based on the provided reason and internal policies.
	Nack(ctx context.Context, reason error) error
}
