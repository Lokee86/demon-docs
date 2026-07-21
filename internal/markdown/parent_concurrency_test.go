package markdown

import (
	"fmt"
	"sync"
	"testing"
)

func TestUpdateParentSupportsConcurrentCacheAccess(t *testing.T) {
	const count = 64
	var workers sync.WaitGroup
	workers.Add(count)
	failures := make(chan string, count)
	for index := 0; index < count; index++ {
		go func(index int) {
			defer workers.Done()
			label := fmt.Sprintf("Parent %d", index%8)
			desired := fmt.Sprintf("%s: [Root](./INDEX.md)", label)
			got := UpdateParent("# Document\n", desired, label)
			if got != "# Document\n\n"+desired+"\n" {
				failures <- got
			}
		}(index)
	}
	workers.Wait()
	close(failures)
	for failure := range failures {
		t.Fatalf("unexpected concurrent update result: %q", failure)
	}
}
