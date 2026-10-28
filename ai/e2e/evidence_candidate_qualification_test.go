package e2e

import "testing"

// A complete real-mode measurement remains candidate in the immutable input.
// This private synthetic namespace tests qualification, never vendor support.
func TestCompleteCandidateCanQualifyWithoutRelabeling(t *testing.T) {
	native := miniMaxAuthorityBuild(t, "native", "./ai/live", true)
	consumer := miniMaxAuthorityBuild(t, "releasegate", "./ai/release/cmd/releasegate", false)
	ev := run.Case(t, "VS55-candidate-complete-qualification")
	controlled := miniMaxAuthorityCapture(t, ev, native, "chat")
	candidate := miniMaxAuthorityFormalFixture(t, controlled, native, "candidate")
	exit, output := miniMaxAuthorityCLI(t, ev, consumer, candidate, "-formal")
	ev.Check("complete candidate measurement qualifies using current model data", exit == 0, "exit=%d %s", exit, output)
	ev.Check("original candidate stage remains unchanged", candidate.report.Stage == "candidate", "stage rewritten")
}
