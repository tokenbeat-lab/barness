// Package ai is barness-ai: one trusted Client that runs generation turns
// against model providers for many tenants (spec Implementation Decisions).
//
// # Entry points
//
// Stream and StreamSimple return a Stream of events at once and produce in
// the background; Complete and CompleteSimple return only the final Result.
// Callers that only need the final message should use Complete: a Stream
// queues every event for its consumer, and a Stream whose events are not read
// (awaiting only Result) ends with a resource_limit error once its output
// outgrows the event queue (ResourcePolicy.MaxQueuedEvents and
// MaxQueuedEventBytes). Complete queues nothing.
//
// # Resource policy
//
// Every Client is built with an explicit, finite ResourcePolicy (ADR-0002);
// there are no built-in defaults. Byte limits are enforced while a request is
// encoded and while responses are read, so neither a slow consumer nor a
// misbehaving provider can make a call's memory grow without bound.
//
// # Admission and time limits
//
// Every attempt, retries included, runs under a concurrency permit of the
// built-in per-tenant and process limits and, when the host injects one
// (Config.Admission), of the host's admission, which receives the tenant and
// the vendor account. At capacity an attempt is refused or waits a bounded,
// cancelable time. A call ends at the earliest of the host's deadline, the
// policy's CallTimeout and, per attempt, the connect, response header, read
// idle and protocol (timeoutMs) limits, as CodeDeadlineExceeded. Closing a
// Stream or canceling its context ends the call's I/O and returns its
// permits; nothing depends on reading further events.
package ai
