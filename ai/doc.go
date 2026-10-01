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
package ai
