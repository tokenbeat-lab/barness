package hostintegration

import (
	"context"
	"sync"

	"github.com/tokenbeat-lab/barness/ai"
)

// Merge forwards the envelopes of several streams to one channel, as a host
// does that multiplexes a caller's parallel turns over one connection or
// feeds a process-wide event bus. Envelopes of different streams interleave
// in whatever order they arrive; a consumer routes each by its Call
// (TenantID, RequestID), never by order or by which stream it expects next.
//
// Merge consumes the streams; the channel closes once every stream ended.
// The consumer must read the channel until it closes or cancel ctx: Merge
// never drops an envelope, so a consumer that stops reading without
// canceling holds every stream open. When ctx ends, Merge stops forwarding
// and closes the streams it was still reading, canceling their generations.
func Merge(ctx context.Context, streams ...*ai.Stream) <-chan ai.EventEnvelope {
	out := make(chan ai.EventEnvelope)
	var wg sync.WaitGroup
	for _, s := range streams {
		wg.Go(func() {
			defer s.Close()
			for s.Next() {
				select {
				case out <- s.Envelope():
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
