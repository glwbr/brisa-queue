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
}

// New creates a new in-memory queue.
func New() *Queue {
	// TODO: Implement
	return nil
}

func (q *Queue) Enqueue(ctx context.Context, topic string, payload []byte) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.topics == nil {
		q.topics = make(map[string]*topicQueue)
	}

	tq, ok := q.topics[topic]
	if !ok {
		tq = &topicQueue{
			inflight: make(map[string]*job),
		}
		q.topics[topic] = tq
	}

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
	// TODO: Implement
	return nil, nil
}

// TODO: move this to a helper function.
func generateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)

	rand.Read(randomBytes)

	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}
