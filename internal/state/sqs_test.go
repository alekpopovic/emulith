package state

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSQSVisibilityReceiptAndPurge(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Unix(1000, 0).UTC()
	q, err := s.CreateSQSQueue(ctx, SQSQueue{Name: "queue", URLPath: "/000000000000/queue", VisibilityTimeout: 30, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SendSQSMessage(ctx, q.Name, "body", now); err != nil {
		t.Fatal(err)
	}
	first, err := s.ReceiveSQSMessages(ctx, q.Name, 1, 30, now)
	if err != nil || len(first) != 1 {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	none, err := s.ReceiveSQSMessages(ctx, q.Name, 1, 30, now)
	if err != nil || len(none) != 0 {
		t.Fatalf("none=%#v err=%v", none, err)
	}
	second, err := s.ReceiveSQSMessages(ctx, q.Name, 1, 30, now.Add(31*time.Second))
	if err != nil || len(second) != 1 || second[0].ReceiptHandle == first[0].ReceiptHandle {
		t.Fatalf("second=%#v err=%v", second, err)
	}
	if err := s.DeleteSQSMessage(ctx, q.Name, first[0].ReceiptHandle); err != ErrNotFound {
		t.Fatalf("stale receipt err=%v", err)
	}
	if err := s.PurgeSQSQueue(ctx, q.Name); err != nil {
		t.Fatal(err)
	}
	visible, hidden, err := s.SQSMessageCounts(ctx, q.Name, now.Add(time.Hour))
	if err != nil || visible != 0 || hidden != 0 {
		t.Fatalf("counts=%d/%d err=%v", visible, hidden, err)
	}
}
func TestSQSConcurrentReceive(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now().UTC()
	s.CreateSQSQueue(ctx, SQSQueue{Name: "queue", URLPath: "/000000000000/queue", VisibilityTimeout: 30, CreatedAt: now})
	s.SendSQSMessage(ctx, "queue", "body", now)
	var wg sync.WaitGroup
	results := make(chan []SQSMessage, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); m, _ := s.ReceiveSQSMessages(ctx, "queue", 1, 30, now); results <- m }()
	}
	wg.Wait()
	close(results)
	total := 0
	for messages := range results {
		total += len(messages)
	}
	if total != 1 {
		t.Fatalf("received %d copies", total)
	}
}
