package adapter

import "context"

// Throttle wraps an LLMAdapter and bounds the number of concurrent Polish calls
// with a counting semaphore. It keeps Pollex within the NaN gateway's
// account-wide concurrency cap (5) and leaves headroom for other consumers
// (Hermes, qq), avoiding self-inflicted 429s under bursts. It is itself an
// LLMAdapter, so it composes with FallbackChain.
//
// Note: this bounds *concurrency*, not request *rate* (the gateway also enforces
// ~100 RPM); a token-bucket rate limiter would be needed for the latter.
type Throttle struct {
	Adapter LLMAdapter
	sem     chan struct{} // nil => unlimited
}

// NewThrottle bounds the wrapped adapter to max concurrent Polish calls. A max
// of 0 or less disables throttling (pass-through).
func NewThrottle(a LLMAdapter, max int) *Throttle {
	var sem chan struct{}
	if max > 0 {
		sem = make(chan struct{}, max)
	}
	return &Throttle{Adapter: a, sem: sem}
}

func (t *Throttle) Name() string    { return t.Adapter.Name() }
func (t *Throttle) Available() bool { return t.Adapter.Available() }

func (t *Throttle) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	if t.sem != nil {
		select {
		case t.sem <- struct{}{}:
			defer func() { <-t.sem }()
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return t.Adapter.Polish(ctx, text, systemPrompt)
}
