package memory

import (
	"testing"

	"github.com/brisa-queue/queue"
	"github.com/brisa-queue/queue/contract"
)

func TestMemoryQueue(t *testing.T) {
	contract.RunQueueContractTests(t, func(t *testing.T) (queue.Queue, func()) {
		q := New()
		return q, func() {
			/* NOTE: No cleanup needed for in-memory queue. This is handled by the GC. */
		}
	})
}
