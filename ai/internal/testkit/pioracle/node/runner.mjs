// Frozen pi-ai runner for the barness-ai differential oracle.
//
// Reads one case from stdin, drives the frozen pi-ai against the local
// controlled Provider named by the case, and writes what pi observed to stdout:
// every event (serialized when received, so live partials are snapshots) and
// the final message. It never runs pi's own test suite.
//
// Isolation (spec Testing Decisions §2): before pi is loaded, every outgoing
// socket is restricted to loopback, so a misrouted request can never reach a
// real vendor; the caller passes a scrubbed environment and the runner refuses
// to start if credentials or endpoints leak in through it; the API key comes
// only from the case, and pi's provider-env lookup gets an explicit empty map.

import { createHash } from "node:crypto";
import { readdirSync, readFileSync } from "node:fs";
import net from "node:net";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const LEAKY_ENV = /(API_KEY|_TOKEN|_SECRET|BASE_URL|ENDPOINT|PROXY)$/i;

// One case per API. A table entry is added when that protocol joins the oracle.
const APIS = {
	"openai-responses": {
		api: "@earendil-works/pi-ai/api/openai-responses",
		models: "@earendil-works/pi-ai/providers/openai.models",
		catalog: "OPENAI_MODELS",
	},
};

function fail(message) {
	process.stderr.write(`pi-oracle runner: ${message}\n`);
	process.exit(2);
}

function refuseLeakyEnv() {
	const leaked = Object.keys(process.env).filter((k) => LEAKY_ENV.test(k));
	if (leaked.length > 0) fail(`refusing to run with credential/endpoint environment variables: ${leaked.join(", ")}`);
}

function isLoopbackHost(host) {
	if (host === "localhost") return true;
	const bare = host.replace(/^\[|\]$/g, "");
	if (net.isIPv4(bare)) return bare.startsWith("127.");
	if (net.isIPv6(bare)) return bare === "::1";
	return false;
}

// installLoopbackGuard rejects every TCP connect whose target is not a
// loopback literal, and every IPC (path) connect. fetch (undici), http and the
// vendor SDKs all end up in net.Socket#connect.
function installLoopbackGuard() {
	const connect = net.Socket.prototype.connect;
	net.Socket.prototype.connect = function guardedConnect(...args) {
		let opts = args[0];
		if (Array.isArray(opts)) opts = opts[0];
		let host;
		if (opts !== null && typeof opts === "object") {
			if (opts.path !== undefined) throw new Error("pi-oracle loopback guard: IPC connect refused");
			host = opts.host ?? "localhost";
		} else if (typeof opts === "string" && Number.isNaN(Number(opts))) {
			throw new Error("pi-oracle loopback guard: IPC connect refused");
		} else {
			host = typeof args[1] === "string" ? args[1] : "localhost";
		}
		if (!isLoopbackHost(host)) throw new Error(`pi-oracle loopback guard: refusing non-loopback host ${host}`);
		return connect.apply(this, args);
	};
}

async function readStdin() {
	const chunks = [];
	for await (const chunk of process.stdin) chunks.push(chunk);
	return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

// packageInfo reports the installed pi-ai version and a hash over its generated
// model data, so the Go side can refuse a copy that drifted from PROVENANCE.md.
function packageInfo() {
	const dist = dirname(fileURLToPath(import.meta.resolve("@earendil-works/pi-ai")));
	const pkg = JSON.parse(readFileSync(join(dist, "..", "package.json"), "utf8"));
	const dataDir = join(dist, "providers", "data");
	const hash = createHash("sha256");
	for (const name of readdirSync(dataDir).sort()) {
		const sum = createHash("sha256").update(readFileSync(join(dataDir, name))).digest("hex");
		hash.update(`${sum}  ${name}\n`);
	}
	return { version: pkg.version, modelDataSha256: hash.digest("hex") };
}

// thinkingBudgets evaluates pi's shared thinking budget rules
// (api/simple-options adjustMaxTokensForThinking) for each case. These rules
// have no Responses wire effect, so they are compared directly.
async function thinkingBudgets(cases) {
	const { adjustMaxTokensForThinking } = await import("@earendil-works/pi-ai/api/simple-options");
	return cases.map((c) => adjustMaxTokensForThinking(c.baseMaxTokens ?? undefined, c.modelMaxTokens, c.level, c.budgets));
}

async function main() {
	refuseLeakyEnv();
	installLoopbackGuard();
	const input = await readStdin();
	if (input.entry === "thinkingBudgets") {
		const results = await thinkingBudgets(input.cases);
		process.stdout.write(JSON.stringify({ pi: packageInfo(), node: process.version, results }));
		return;
	}
	const entry = APIS[input.api];
	if (!entry) fail(`unsupported api ${JSON.stringify(input.api)}`);
	if (!input.baseUrl || !isLoopbackHost(new URL(input.baseUrl).hostname)) fail("baseUrl must be a loopback URL");
	if (!input.apiKey) fail("apiKey is required; the runner never falls back to the environment");

	const pi = await import("@earendil-works/pi-ai");
	const api = await import(entry.api);
	const catalog = (await import(entry.models))[entry.catalog];
	const known = Object.values(catalog).find((m) => m.provider === input.provider && m.id === input.model);
	if (!known) fail(`model ${input.provider}/${input.model} is not in pi's ${input.api} catalog`);
	// modelCompat overrides compat flags of the catalog model, the way a pi
	// user configures a custom model; it lets a case exercise a flag that
	// pi's catalog only sets on models barness-ai does not list yet.
	// modelPatch replaces other top-level model fields (thinkingLevelMap,
	// samplingParams, contextWindow, ...) the same way.
	const compat = input.modelCompat ? { ...known.compat, ...input.modelCompat } : known.compat;
	const model = { ...known, ...(input.modelPatch ?? {}), baseUrl: input.baseUrl, ...(compat ? { compat } : {}) };

	const context = pi.normalizeContext(input.context);
	// abortAfterEvents > 0 aborts the call, as a caller canceling it would, once
	// that many events were received.
	const controller = new AbortController();
	const options = { ...(input.options ?? {}), apiKey: input.apiKey, env: {}, signal: controller.signal };
	if (input.entry !== "stream" && input.entry !== "streamSimple") fail(`unsupported entry ${input.entry}`);
	const run = input.entry === "streamSimple" ? api.streamSimple : api.stream;

	const events = [];
	const stream = run(model, context, options);
	for await (const event of stream) {
		events.push(JSON.parse(JSON.stringify(event)));
		if (events.length === input.abortAfterEvents) controller.abort();
	}
	const result = JSON.parse(JSON.stringify(await stream.result()));

	process.stdout.write(JSON.stringify({ pi: packageInfo(), node: process.version, events, result }));
}

main().catch((error) => fail(error?.stack ?? String(error)));
