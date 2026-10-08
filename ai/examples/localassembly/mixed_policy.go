package localassembly

import "github.com/tokenbeat-lab/barness/ai"

// MixedPolicy enables chat, image and classifier for a local program with
// 256 MiB available for model buffers. E08-mixed-pressure-local-design-load
// (BARNESS_AI_PRESSURE=1) measures four simultaneous calls: one chat with
// 128 KiB history/32 KiB output, OpenAI and Google edits each with two
// 256 KiB references (OpenAI also a 256 KiB mask), two 512 KiB output
// images per edit, and a classifier with 128 KiB state/eight 4 KiB questions.
// Images expand by 4/3 in base64; unary JSON and every image are read and
// validated in full before publication. The report includes reachable heap,
// Provider request copies, throughput and <=75% byte-limit use. These
// values apply to that load, not to arbitrary chat or image deployments.
// Re-run the probe before increasing concurrency or sizes. The larger
// LocalPolicy remains the separate chat-only example.
func MixedPolicy() *ai.ResourcePolicy {
	p := LocalPolicy()
	p.MaxConcurrentPerTenant, p.MaxConcurrentProcess = 4, 4
	p.MaxAdmissionWaiters = 2
	p.MaxRequestBytes, p.MaxImageBytes = 4<<20, 1<<20
	p.MaxFrameBytes, p.MaxOutputBytes = 256<<10, 8<<20
	p.Image = &ai.ImagePolicy{
		MaxInputImages: 4, MaxOutputImages: 4,
		MaxOutputImageBytes: 1 << 20, MaxTotalOutputImageBytes: 2 << 20,
	}
	p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 16, MaxStateBytes: 256 << 10, MaxQuestionBytes: 8 << 10}
	return p
}
