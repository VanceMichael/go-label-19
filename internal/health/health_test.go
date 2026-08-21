package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestParallelReadinessPreservesEveryRegisteredCheck(t *testing.T) {
	registry := New()
	const count = 12
	var started sync.WaitGroup
	started.Add(count)
	release := make(chan struct{})
	for index := 0; index < count; index++ {
		name := fmt.Sprintf("dependency-%02d", index)
		failed := index%3 == 0
		registry.Register(name, func(context.Context) error {
			started.Done()
			<-release
			if failed {
				return errors.New("dependency unavailable")
			}
			return nil
		})
	}

	results := make(chan []Status, 1)
	go func() { results <- registry.Run(context.Background()) }()
	started.Wait()
	close(release)
	statuses := <-results
	if len(statuses) != count {
		t.Fatalf("readiness returned %d checks, want %d", len(statuses), count)
	}
	for index, status := range statuses {
		wantName := fmt.Sprintf("dependency-%02d", index)
		wantOK := index%3 != 0
		if status.Name != wantName || status.OK != wantOK {
			t.Fatalf("status %d = %+v, want name=%s ok=%v", index, status, wantName, wantOK)
		}
		if wantOK && status.Error != "" || !wantOK && status.Error != "dependency unavailable" {
			t.Fatalf("status %s error = %q", status.Name, status.Error)
		}
	}
}
