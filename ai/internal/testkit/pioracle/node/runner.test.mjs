// Startup failures must be rejected before importing/running a provider.
import assert from "node:assert/strict";
import { mkdtempSync, cpSync, readFileSync, writeFileSync, mkdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";
import { spawn } from "node:child_process";
import http from "node:http";

const source = dirname(fileURLToPath(import.meta.url));
for (const [field, value, message] of [
	["version", "0.0.0", "version mismatch"],
	["commit", "0".repeat(40), "commit mismatch"],
	["data", "altered", "model data hash mismatch"],
	["source", "altered", "source hash mismatch"],
]) {
	test(field, () => {
		const dir = mkdtempSync(join(tmpdir(), "barness-oracle-refusal-"));
		try {
			cpSync(join(source, "runner.mjs"), join(dir, "runner.mjs"));
			const installed = join(dir, "node_modules", "@earendil-works", "pi-ai");
			mkdirSync(dirname(installed), { recursive: true });
			cpSync(join(source, "node_modules", "@earendil-works", "pi-ai"), installed, { recursive: true });
			cpSync(join(source, "provenance.json"), join(dir, "provenance.json"));
			if (field === "version") {
				const path = join(installed, "package.json");
				const p = JSON.parse(readFileSync(path, "utf8"));
				p.version = value;
				writeFileSync(path, JSON.stringify(p));
			} else if (field === "commit") {
				const path = join(dir, "provenance.json");
				const p = JSON.parse(readFileSync(path, "utf8"));
				p.pi_commit = value;
				writeFileSync(path, JSON.stringify(p));
			} else if (field === "source") {
				writeFileSync(join(installed, "dist", "api", "openai-decisions.js"), value);
			} else {
				writeFileSync(join(installed, "dist", "providers", "data", "openai.json"), value);
			}
			const out = spawnSync(process.execPath, [join(dir, "runner.mjs")], {
				input: JSON.stringify({ entry: "models", models: [] }), encoding: "utf8", env: {}, timeout: 10000,
			});
			assert.equal(out.status, 2);
			assert.match(out.stderr, new RegExp(message));
			assert.equal(out.stdout, "");
		} finally {
			rmSync(dir, { recursive: true, force: true });
		}
	});
}

function run(input, env = {}) {
	return new Promise((resolve, reject) => {
		const child = spawn(process.execPath, [join(source, "runner.mjs")], { env });
		let stdout = "", stderr = "";
		child.stdout.on("data", (b) => { stdout += b; });
		child.stderr.on("data", (b) => { stderr += b; });
		child.on("error", reject);
		child.on("close", (status) => resolve({ status, stdout, stderr }));
		child.stdin.end(JSON.stringify(input));
	});
}

const classifier = {
	api: "openai-decisions", provider: "openai", model: "gpt-6-luna", entry: "classify",
	apiKey: "oracle-fixture-key", options: { maxRetries: 0 },
	context: { state: { text: "hello" }, questions: { accepted: { type: "bool", instructions: "Accepted?", criteria: {} } } },
};

test("Decisions selects the classifier catalog independently of chat Luna", async () => {
	const out = await run({ entry: "models", models: [
		{ provider: "openai", api: "openai-decisions", id: "gpt-6-luna" },
		{ provider: "openai", api: "openai-responses", id: "gpt-6-luna" },
		{ provider: "typesafe", api: "openai-decisions", id: "gpt-6-luna" },
	] });
	assert.equal(out.status, 0, out.stderr);
	const [c, chat, foreign] = JSON.parse(out.stdout).results;
	assert.equal(c.type, "classifier");
	assert.equal(c.api, "openai-decisions");
	assert.equal(c.cost.input, 0.1);
	assert.equal(chat.type, "chat");
	assert.equal(foreign, null);
});

test("Decisions uses the controlled HTTP endpoint and named answers", async () => {
	const captures = [];
	const server = http.createServer(async (req, res) => {
		let raw = "";
		for await (const b of req) raw += b;
		captures.push({ path: req.url, auth: req.headers.authorization, body: JSON.parse(raw) });
		res.setHeader("content-type", "application/json");
		res.end(JSON.stringify({ answers: [{ type: "predicate", name: "accepted", probability: 0.75 }] }));
	});
	await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
	try {
		const out = await run({ ...classifier, baseUrl: `http://127.0.0.1:${server.address().port}/v1` });
		assert.equal(out.status, 0, out.stderr);
		assert.equal(JSON.parse(out.stdout).result.answers.accepted.probability, 0.75);
		assert.deepEqual(captures, [{ path: "/v1/decisions", auth: "Bearer oracle-fixture-key", body: {
			model: "gpt-6-luna", input: '{"text":"hello"}', questions: [{ type: "predicate", name: "accepted", instructions: "Accepted?" }],
		} }]);
		const wrong = await run({ ...classifier, modelPatch: { type: "chat" }, baseUrl: `http://127.0.0.1:${server.address().port}/v1` });
		assert.equal(wrong.status, 2);
		assert.equal(captures.length, 1);
	} finally { await new Promise((resolve) => server.close(resolve)); }
});

test("credentials, non-loopback destinations and redirected loopback are refused", async () => {
	const leaked = await run({ entry: "models", models: [] }, { OPENAI_API_KEY: "must-never-be-read" });
	assert.equal(leaked.status, 2);
	assert.doesNotMatch(leaked.stderr, /must-never-be-read/);
	const remote = await run({ ...classifier, baseUrl: "https://api.openai.com/v1" });
	assert.equal(remote.status, 2);
	assert.match(remote.stderr, /controlled loopback/);
	let reached = false;
	const target = http.createServer((req, res) => { reached = true; res.end("{}"); });
	await new Promise((resolve) => target.listen(0, "127.0.0.1", resolve));
	const redirect = http.createServer((req, res) => {
		res.writeHead(307, { location: `http://127.0.0.1:${target.address().port}/v1/decisions` });
		res.end();
	});
	await new Promise((resolve) => redirect.listen(0, "127.0.0.1", resolve));
	try {
		const out = await run({ ...classifier, baseUrl: `http://127.0.0.1:${redirect.address().port}/v1` });
		assert.equal(out.status, 0, out.stderr);
		assert.equal(JSON.parse(out.stdout).result.stopReason, "error");
		assert.equal(reached, false);
	} finally {
		await Promise.all([new Promise((r) => redirect.close(r)), new Promise((r) => target.close(r))]);
	}
});
