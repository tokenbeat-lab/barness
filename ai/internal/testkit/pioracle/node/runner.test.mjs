// Startup failures must be rejected before importing/running a provider.
import assert from "node:assert/strict";
import { mkdtempSync, cpSync, readFileSync, writeFileSync, mkdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";

const source = dirname(fileURLToPath(import.meta.url));
for (const [field, value, message] of [
	["version", "0.0.0", "version mismatch"],
	["data", "altered", "model data hash mismatch"],
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
