// TODO: Add the documentation for the package.
// See: https://go.dev/doc/comment
package queue

import (
	"context"
)

/* State Machine for a Job:
 *  Pending   →  In-Flight   →  Done
 *             ↘ retry ↗
 */

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

/* TODO: Create a "Binder" for our domain logic?. */
type Queue interface {
	// Enqueue adds a new job to the queue. It returns the assigned job ID or an error.
	Enqueue(ctx context.Context, topic string, payload []byte) (string, error)

	// Dequeue blocks until a job is available in the queue.
	Dequeue(ctx context.Context, topic string) (Job, error)
}
