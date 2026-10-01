package ai

import "context"

// Complete runs one generation turn with full protocol options and returns
// only the final result. It shares Stream's production and merge process but
// queues no events. The error is nil on success and the classified *Error
// when the message ends with StopReason error or aborted; the Result is
// complete in both cases.
func (c *Client) Complete(ctx context.Context, scope CallScope, target Target, req Request, opts Options) (Result, error) {
	return c.run(ctx, newCall(scope, target, req, opts, nil), nil)
}

// CompleteSimple is Complete with protocol-neutral options mapped per model.
func (c *Client) CompleteSimple(ctx context.Context, scope CallScope, target Target, req Request, opts SimpleOptions) (Result, error) {
	return c.run(ctx, newCall(scope, target, req, nil, &opts), nil)
}
