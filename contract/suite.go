package contract

import (
	"context"
	"testing"
	"time"

	"github.com/brisa-queue/queue"
)

// QueueFactory is a function that returns a clean Queue instance for testing.
// It also returns a teardown function to clean up resources (e.g., flush DB, close connections).
type QueueFactory func(t *testing.T) (q queue.Queue, teardown func())

// RunQueueContractTests runs the standard suite of behavioral tests against
// a Queue implementation provided by the factory.
func RunQueueContractTests(t *testing.T, factory QueueFactory) {
	t.Helper()

	t.Run("EnqueueAndDequeue", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()
		topic := "test-topic"
		payload := []byte("hello world")

		id, err := q.Enqueue(ctx, topic, payload)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
		if id == "" {
			t.Error("Expected non-empty job ID")
		}

		job, err := q.Dequeue(ctx, topic)
		if err != nil {
			t.Fatalf("Dequeue failed: %v", err)
		}

		if string(job.Payload()) != string(payload) {
			t.Errorf("Expected payload %s, got %s", payload, job.Payload())
		}
		if job.ID() != id {
			t.Errorf("Expected job ID %s, got %s", id, job.ID())
		}

		if err := job.Ack(ctx); err != nil {
			t.Errorf("Ack failed: %v", err)
		}
	})

	t.Run("FIFOOrdering", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()
		topic := "test-topic"
		payloads := []string{"job-1", "job-2", "job-3"}

		for _, p := range payloads {
			if _, err := q.Enqueue(ctx, topic, []byte(p)); err != nil {
				t.Fatalf("Enqueue failed for %s: %v", p, err)
			}
		}

		for _, expected := range payloads {
			job, err := q.Dequeue(ctx, topic)
			if err != nil {
				t.Fatalf("Dequeue failed for expected %s: %v", expected, err)
			}

			if string(job.Payload()) != expected {
				t.Errorf("Order violation: expected %s, got %s", expected, job.Payload())
			}

			if err := job.Ack(ctx); err != nil {
				t.Errorf("Ack failed: %v", err)
			}
		}
	})

	t.Run("DequeueBlockingOrError", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		topic := "test-topic"
		done := make(chan error, 1)

		go func() {
			_, err := q.Dequeue(ctx, topic)
			done <- err
		}()

		select {
		case <-done:
			t.Fatal("Dequeue returned early on empty queue")
		case <-time.After(50 * time.Millisecond):
			// correct: still blocking
		}

		cancel()

		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected error after context cancellation")
			}
		case <-time.After(time.Second):
			t.Fatal("Dequeue did not unblock after context cancellation")
		}
	})

	t.Run("NackRequeues", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()
		payload := []byte("retry-me")
		topic := "test-topic"

		if _, err := q.Enqueue(ctx, topic, payload); err != nil {
			t.Fatalf("enqueue failed: %v", err)
		}

		job, err := q.Dequeue(ctx, topic)
		if err != nil {
			t.Fatalf("dequeue failed: %v", err)
		}

		if err := job.Nack(ctx, context.Canceled); err != nil {
			t.Fatalf("nack failed: %v", err)
		}

		job2, err := q.Dequeue(ctx, topic)
		if err != nil {
			t.Fatalf("dequeue after nack failed: %v", err)
		}

		if got := string(job2.Payload()); got != string(payload) {
			t.Fatalf("payload mismatch after nack: got %q want %q", got, payload)
		}
	})

	t.Run("IsolationBetweenTopics", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()
		topicA := "topic-a"
		topicB := "topic-b"

		if _, err := q.Enqueue(ctx, topicA, []byte("data-a")); err != nil {
			t.Fatal(err)
		}

		ctxB, cancelB := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancelB()

		if jobB, err := q.Dequeue(ctxB, topicB); err == nil {
			t.Errorf("Expected error/timeout on empty topic B, but got job: %s", string(jobB.Payload()))
		}

		ctxA, cancelA := context.WithTimeout(ctx, 1*time.Second)
		defer cancelA()

		jobA, err := q.Dequeue(ctxA, topicA)
		if err != nil {
			t.Fatalf("Failed to dequeue from topic A: %v", err)
		}

		if string(jobA.Payload()) != "data-a" {
			t.Errorf("Got wrong data from topic A: expected 'data-a', got '%s'", jobA.Payload())
		}

		if err := jobA.Ack(ctxA); err != nil {
			t.Errorf("Ack failed: %v", err)
		}
	})

	t.Run("AckTwiceFails", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()

		topic := "test-topic"
		_, err := q.Enqueue(ctx, topic, []byte("data"))
		if err != nil {
			t.Fatal(err)
		}

		job, err := q.Dequeue(ctx, "test-topic")
		if err != nil {
			t.Fatal(err)
		}

		if err := job.Ack(ctx); err != nil {
			t.Fatalf("first ack failed: %v", err)
		}

		if err := job.Ack(ctx); err == nil {
			t.Fatal("expected error on second ack, got nil")
		}
	})

	t.Run("NackRequeuesAtBack", func(t *testing.T) {
		q, teardown := factory(t)
		defer teardown()

		ctx := context.Background()

		topic := "test-topic"

		q.Enqueue(ctx, topic, []byte("job-1"))
		q.Enqueue(ctx, topic, []byte("job-2"))

		job1, _ := q.Dequeue(ctx, topic)
		job2, _ := q.Dequeue(ctx, topic)

		_ = job1.Nack(ctx, context.Canceled)
		_ = job2.Ack(ctx)

		job3, _ := q.Dequeue(ctx, topic)

		if string(job3.Payload()) != "job-1" {
			t.Fatalf("expected job-1 after nack, got %s", job3.Payload())
		}
	})
}
