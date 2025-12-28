package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Queue struct {
	mu     sync.Mutex
	topics map[string]*topicQueue
}

type topicQueue struct {
	pending  []*job
	inflight map[string]*job
	cond     *sync.Cond
}

// New creates a new in-memory queue.
func New() *Queue {
	// TODO: Implement
	return nil
}

func (q *Queue) Enqueue(ctx context.Context, topic string, payload []byte) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tq := q.getOrCreateTopic(topic)

	id := generateID()
	j := &job{
		id:      id,
		payload: payload,
		queue:   q,
		topic:   topic,
	}

	tq.pending = append(tq.pending, j)
	return id, nil
}

func (q *Queue) Dequeue(ctx context.Context, topic string) (*job, error) {
	// NOTE: (to myself) Context-aware blocking dequeue implementation:
	//
	// sync.Cond.Wait() blocks until Signal()/Broadcast() but doesn't respect context cancellation.
	// To make this cancellable, we spawn a goroutine that watches ctx.Done() and calls Signal()
	// to wake the Wait() loop when the context is cancelled.
	//
	// The 'done' channel prevents the goroutine from leaking after a successful dequeue by
	// signaling it to exit. We must acquire the lock before Signal() in the goroutine because
	// Signal() requires the associated lock to be held (per sync.Cond semantics).
	//
	// Main goroutine holds q.mu throughout to safely access tq.pending and tq.inflight.

	q.mu.Lock()
	defer q.mu.Unlock()

	tq := q.getOrCreateTopic(topic)

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			q.mu.Lock()
			tq.cond.Signal()
			q.mu.Unlock()
		case <-done:
		}
	}()

	defer close(done)

	for len(tq.pending) == 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tq.cond.Wait()
	}

	j := tq.pending[0]
	tq.pending = tq.pending[1:]

	j.state = jobInFlight
	tq.inflight[j.id] = j

	return j, nil
}

func (q *Queue) getOrCreateTopic(topic string) *topicQueue {
	tq, ok := q.topics[topic]
	if ok {
		return tq
	}

	tq = &topicQueue{
		pending:  make([]*job, 0),
		inflight: make(map[string]*job),
	}

	/* NOTE: cond MUST use the SAME mutex that guards pending/inflight. */
	tq.cond = sync.NewCond(&q.mu)

	q.topics[topic] = tq
	return tq
}

// TODO: move this to a helper function.
func generateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)

	rand.Read(randomBytes)

	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}
