package arraysource

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewArraySourceFrom_UniqueNames_Concurrent(t *testing.T) {
	const count = 100
	var wg sync.WaitGroup
	wg.Add(count)

	names := make(chan string, count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			items := []Status{{ID: idx, Name: fmt.Sprintf("Status %d", idx)}}
			model := NewArraySourceFrom(items)
			names <- model.TableName()
		}(i)
	}

	wg.Wait()
	close(names)

	uniqueNames := make(map[string]bool)
	for name := range names {
		if uniqueNames[name] {
			t.Errorf("duplicate table name found: %s", name)
		}
		uniqueNames[name] = true
	}

	if len(uniqueNames) != count {
		t.Errorf("expected %d unique table names, got %d", count, len(uniqueNames))
	}
}
