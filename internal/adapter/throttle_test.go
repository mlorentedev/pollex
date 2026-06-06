package adapter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// blockingAdapter records peak concurrency and blocks in Polish until `gate`
// closes, so a test can hold several calls in-flight simultaneously.
type blockingAdapter struct {
	inFlight int32
	maxSeen  int32
	gate     chan struct{}
	calls    int32
}

func (b *blockingAdapter) Name() string    { return "blocking" }
func (b *blockingAdapter) Available() bool { return true }
func (b *blockingAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	atomic.AddInt32(&b.calls, 1)
	n := atomic.AddInt32(&b.inFlight, 1)
	for {
		old := atomic.LoadInt32(&b.maxSeen)
		if n <= old || atomic.CompareAndSwapInt32(&b.maxSeen, old, n) {
			break
		}
	}
	<-b.gate
	atomic.AddInt32(&b.inFlight, -1)
	return "ok", nil
}

func TestThrottle_BoundsConcurrency(t *testing.T) {
	const limit = 2
	b := &blockingAdapter{gate: make(chan struct{})}
	th := NewThrottle(b, limit)

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			th.Polish(context.Background(), "x", "p")
		}()
	}
	time.Sleep(50 * time.Millisecond) // let the first wave acquire slots + block
	close(b.gate)                     // release everyone
	wg.Wait()

	if got := atomic.LoadInt32(&b.maxSeen); got > limit {
		t.Errorf("peak concurrency = %d, want <= %d", got, limit)
	}
	if got := atomic.LoadInt32(&b.calls); got != 6 {
		t.Errorf("all calls should eventually run: got %d, want 6", got)
	}
}

func TestThrottle_ContextCancelWhileWaiting(t *testing.T) {
	b := &blockingAdapter{gate: make(chan struct{})}
	th := NewThrottle(b, 1)

	// Occupy the single slot.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); th.Polish(context.Background(), "x", "p") }()
	time.Sleep(20 * time.Millisecond)

	// A second call with an already-cancelled context must not wait/run.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := th.Polish(ctx, "x", "p"); err == nil {
		t.Error("expected error when context is cancelled while waiting for a slot")
	}
	if got := atomic.LoadInt32(&b.calls); got != 1 {
		t.Errorf("cancelled call must not reach the wrapped adapter: calls=%d, want 1", got)
	}

	close(b.gate)
	wg.Wait()
}

func TestThrottle_ZeroMeansUnlimited(t *testing.T) {
	const n = 8
	b := &blockingAdapter{gate: make(chan struct{})}
	th := NewThrottle(b, 0) // 0 => no limit

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); th.Polish(context.Background(), "x", "p") }()
	}
	time.Sleep(50 * time.Millisecond)
	close(b.gate)
	wg.Wait()

	if got := atomic.LoadInt32(&b.maxSeen); got != n {
		t.Errorf("unlimited throttle peak = %d, want %d", got, n)
	}
}

func TestThrottle_DelegatesNameAndAvailable(t *testing.T) {
	th := NewThrottle(&blockingAdapter{gate: make(chan struct{})}, 1)
	if th.Name() != "blocking" {
		t.Errorf("Name not delegated: %q", th.Name())
	}
	if !th.Available() {
		t.Error("Available not delegated")
	}
}
