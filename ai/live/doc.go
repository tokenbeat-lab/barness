// Package live is barness-ai's real official API smoke (spec Testing
// Decisions §2, §5): the same public Client a host uses, connected to a
// vendor's real API with a low-privilege test account, run before a release
// and after an SDK or model upgrade. Its results feed the support matrix
// (support-matrix.json in this directory).
//
// Its tests are behind two switches, so `go test ./...` never reaches the
// network: the build tag live and BARNESS_AI_LIVE=1. One process runs one
// Operation × Provider × API combination and holds only that combination's key:
//
//	BARNESS_AI_LIVE=1 \
//	BARNESS_AI_LIVE_COMBO=deepseek-chat \
//	BARNESS_AI_LIVE_ACCOUNT_ALIAS=ci-deepseek-lowpriv@cn \
//	BARNESS_AI_LIVE_KEY_DEEPSEEK_CHAT=<injected by the CI secret store> \
//	go test -tags live -count=1 ./ai/live
//
// Combinations (BARNESS_AI_LIVE_COMBO, key variable suffix):
// openai-responses, openai-chat, anthropic-messages, google-gemini,
// deepseek-responses, deepseek-chat, typesafe-classifier, openai-images; the key variable is
// BARNESS_AI_LIVE_KEY_ followed by the name in upper case with "-" as "_".
// TypeSafe uses fixed jev-1.13.0; at most 6 calls / 10 questions and 131072
// state bytes per call. Its public Client context probe reports the actual
// guard (observed 400 max_tokens_exceeded); 422 is explicitly unconfirmed.
// Local dotenv injection and current evidence are described in issue 08.
// OpenAI Images uses the fixed gpt-image-2.5-sunburst-2026-09-08 snapshot.
// Its first required check is an inline JSON edit; generation and mask run only
// after that check passes in this process. At most four logical calls / four
// requested images (including faults and retries), one per call, at 1024x1024.
// Binding retries are disabled. A mask-specific unsupported response is optional;
// JSON editing/generation must PASS. A sourced host probe catalog is used before
// inclusion and for future revalidation. PNG/JPEG/WebP rasters and checksums are
// saved independently of bounded wire captures; a truncated capture stays labelled.
// The explicit local dotenv launcher and verification are in issue 11's evidence.
// Account privileges and region must be configured by the host: an alias records
// provenance and does not prove key permissions.
// Each combination's models are fixed in the suite (combos_test.go), chosen
// to have every capability the combination promises.
//
// The key is read from that one variable only. Nothing falls back to
// OPENAI_API_KEY or any other ambient credential, and a process that also
// sees another combination's key refuses to run: keys are injected per
// combination process, never shared. Without a key every scenario is
// NOT_RUN; a vendor fault is retried within the budget and then FAIL, never
// a skip. Every outcome is PASS, FAIL, NOT_RUN or UNSUPPORTED.
//
// Each run leaves an evidence bundle (.evidence/barness-ai/<run-id>, or
// BARNESS_AI_EVIDENCE_DIR) with redacted wire captures per scenario and
// live-report.json; the key never enters it. Merge reports into the matrix
// with:
//
//	go run ./ai/live/cmd/supportmatrix <bundle-dir>...
//
// A combination's row changes only through its own report: a pass through a
// shared adapter never marks another combination.
package live
