package localassembly

import "github.com/tokenbeat-lab/barness/ai"

// MixedPolicy enables chat, image and classifier for a local program with
// 256 MiB available for model buffers. E08-mixed-pressure-local-design-load
// (BARNESS_AI_PRESSURE=1) measures four simultaneous calls: one chat with
// 128 KiB history/16 KiB output, OpenAI and Google edits each with two
// 256 KiB references (OpenAI also a 256 KiB mask), two 512 KiB output
// images per edit, and a classifier with 128 KiB state/eight 4 KiB questions.
// Images expand by 4/3 in base64; unary JSON and every image are read and
// validated in full before publication. The report includes reachable heap,
// Provider request copies, throughput and <=75% byte-limit use. These
// values apply to that load, not to arbitrary chat or image deployments.
// Re-run the probe before increasing concurrency or sizes. The larger
// LocalPolicy remains the separate chat-only example.
// An additional bounded batch (E08-mixed-pressure-*-classifier-images-design-load)
// fills all 4/8 permits with four Luna/Decisions calls per tenant: each has
// 128 KiB state, eight 4000-byte instructions and two 256 KiB PNGs. A canceled
// batch and a recovery batch repeat that load. Each batch keeps the same
// 256/512 MiB budget; the original four-route batch runs in full separately.
// Classifier image count uses 2/4, each image 256 KiB/1 MiB, decoded total
// 512 KiB/1 MiB; full request use includes base64, data URLs and JSON.
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
	p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 16, MaxStateBytes: 256 << 10, MaxQuestionBytes: 8 << 10,
		Images: &ai.ClassifierImagePolicy{MaxImages: 4, MaxTotalImageBytes: 1 << 20},
	}
	return p
}
