// TODO: Add the documentation for the package.
// See: https://go.dev/doc/comment
package queue

import (
	"context"
)

/* TODO: Create a "Binder" for our domain logic?. */
type Queue interface {
	// Enqueue adds a new job to the queue. It returns the assigned job ID or an error.
	Enqueue(ctx context.Context, topic string, payload []byte) (string, error)

	// Dequeue blocks until a job is available in the queue.
	Dequeue(ctx context.Context, topic string) (Job, error)
}
