package hostintegration

import "github.com/tokenbeat-lab/barness/ai"

// CloudMixedPolicy reserves 512 MiB for model buffers in one instance.
// E08-mixed-pressure-cloud-design-load (BARNESS_AI_PRESSURE=1) measures
// eight simultaneous calls across two tenants, four per tenant: chat
// (128 KiB history/16 KiB output), OpenAI and Google edits (two 256 KiB
// references, OpenAI also a 256 KiB mask, two 512 KiB outputs each), and
// classification (128 KiB state/eight 4 KiB questions). Unlike streamed
// chat, image base64 expands by 4/3 and complete unary JSON/output checks
// keep multiple copies alive. The probe reports reachable heap including
// Provider request copies, throughput and <=75% byte-limit use. Re-run it
// before raising sizes/concurrency; CloudInteractivePolicy's 32 chat
// permits are not an image capacity recommendation. Per-account and
// cross-instance quotas still belong to the host's injected Admission.
func CloudMixedPolicy() *ai.ResourcePolicy {
	p := CloudInteractivePolicy()
	p.MaxConcurrentPerTenant, p.MaxConcurrentProcess = 4, 8
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
