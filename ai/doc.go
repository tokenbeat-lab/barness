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
// # Attribution
//
// Every Result carries its call's immutable CallAttribution (in
// CallMetadata): the trusted scope, the binding named and, once resolved,
// the provider, API, model and account that actually served the call. Every
// stream event is published with the attribution as it stood then
// (Stream.Envelope returns both as an EventEnvelope), so a host merging
// several streams routes each event by its own attribution, never by
// arrival order. A call refused before resolution reports Resolved=false.
//
// # Host responsibilities
//
// The host authenticates callers and builds each CallScope from that, never
// from request content; it reads history by tenant and session and vouches
// for stored native state with TrustNativeState; it executes tools and
// starts every next turn as a new logical call with a new, globally unique
// RequestID; it cancels a call's context, or closes its Stream, when the
// downstream goes away. ai/examples holds minimal local, cloud host and tool
// round trip examples, each run as an offline E2E.
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
// idle and protocol (timeoutMs) limits, as CodeDeadlineExceeded; the Gemini
// Developer API has no protocol timeout, as in pi (ADR-0012). Closing a
// Stream or canceling its context ends the call's I/O and returns its
// permits; nothing depends on reading further events.
//
// # Observability
//
// A host that sets Config.Observer receives a record when each call starts
// and finishes and when each attempt starts and finishes, attributed to the
// trusted scope and the resolved configuration snapshot. Records are
// delivered asynchronously through a queue bounded by
// ResourcePolicy.MaxQueuedObservations; a slow or failing Observer never
// blocks or changes a call, and Client.ObserverStats counts what was
// dropped or failed. Records hold no key, content, tool arguments, native
// state or error text, and barness-ai writes no logs of its own.
//
// # Usage and cost
//
// A message's Usage keeps pi-ai's numbers, including an estimated Cost from
// the catalog's prices (Model.Cost); a zero Usage never means free. Whether
// and how completely the provider reported usage is recorded per attempt
// (Attempt.UsageReporting and Attempt.Usage), and CallMetadata names the
// catalog version and hash the cost was estimated with. Costs are
// estimates, not billing.
package ai
