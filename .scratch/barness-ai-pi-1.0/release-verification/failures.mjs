// Exercise the public checker against real installed/rebuilt artifacts.
// These failure cases precede its implementation; no oracle is invoked.
import { readFileSync, writeFileSync, mkdtempSync, rmSync, existsSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import assert from 'node:assert/strict';

const here = dirname(fileURLToPath(import.meta.url));
const work = resolve(process.argv[2]);
const output = resolve(process.argv[3]);
const repo = resolve(here, '../../..');
const cases = [
  ['wrong-version', 'version', 'release/node_modules/@earendil-works/pi-ai/package.json',
    b => Buffer.from(b.toString().replace('"version": "1.0.0"', '"version": "0.87.1"'))],
  ['unexplained-git-head', 'source-identity', 'registry-ai.json',
    b => Buffer.from(b.toString().replace('a13d35a742c6ef8462812a28fbe1d8c8b7431c32', 'f07218c4d4bbc12bef056a7058c3dd49dfe41abe'))],
  ['missing-model-data', 'model-data', 'release/node_modules/@earendil-works/pi-ai/dist/providers/data/openai.json',
    () => null],
  ['changed-model-data', 'model-data', 'release/node_modules/@earendil-works/pi-ai/dist/providers/data/openai.json',
    b => Buffer.concat([b, Buffer.from('\n')])],
  ['different-ai-build', 'build-equivalence', 'source/packages/ai/dist/index.js',
    b => Buffer.concat([b, Buffer.from('\n// modified\n')])],
  ['different-telemetry-build', 'build-equivalence', 'source/packages/telemetry/dist/index.js',
    b => Buffer.concat([b, Buffer.from('\n// modified\n')])],
  ['unlocked-dependency', 'dependency-lock', 'release/package-lock.json', b => {
    const lock = JSON.parse(b);
    delete lock.packages['node_modules/openai'].integrity;
    return Buffer.from(JSON.stringify(lock));
  }],
  ['changed-tarball', 'tarball-integrity', 'pi-ai-1.0.0.tgz', b => {
    const changed = Buffer.from(b); changed[0] ^= 1; return changed;
  }],
];
const checks = [];
for (const [name, code, file, mutate] of cases) {
  const path = join(work, file);
  const original = readFileSync(path);
  const out = mkdtempSync(join(tmpdir(), 'barness-release-failure-'));
  try {
    const changed = mutate(original);
    if (changed === null) rmSync(path); else writeFileSync(path, changed);
    const result = spawnSync(process.execPath, [join(here, 'check.mjs'), work, out], { encoding: 'utf8' });
    assert.equal(result.status, 1, name);
    assert.ok(existsSync(join(out, 'report.json')), `${name}: missing FAIL artifact`);
    const report = JSON.parse(readFileSync(join(out, 'report.json')));
    assert.equal(report.status, 'FAIL', name);
    assert.equal(report.failure, code, name);
    checks.push({ name, expected_failure: code, status: 'PASS' });
  } finally {
    writeFileSync(path, original);
    rmSync(out, { recursive: true, force: true });
  }
}
// Audit the bytes actually written through the existing Go command. A finding
// must fail, and neither its diagnostic nor the saved finding may repeat a key.
const auditOut = mkdtempSync(join(tmpdir(), 'barness-release-audit-failure-'));
try {
  const secret = 'barness-release-synthetic-audit-secret-02';
  writeFileSync(join(auditOut, 'manifest.json'), '{}\n');
  writeFileSync(join(auditOut, 'capture.json'), JSON.stringify({ sample: secret }) + '\n');
  const result = spawnSync('go', ['run', './ai/internal/testkit/audit/cmd/auditbundle', auditOut], {
    cwd: repo, encoding: 'utf8', env: { ...process.env, BARNESS_RELEASE_AUDIT_TEST_TOKEN: secret }, timeout: 120_000,
  });
  assert.equal(result.status, 1, 'audit-secret');
  const auditBytes = readFileSync(join(auditOut, 'audit.json'), 'utf8');
  assert.ok(JSON.parse(auditBytes).findings.some(f => f.rule === 'environment'), 'audit-secret finding');
  assert.ok(!auditBytes.includes(secret) && !result.stdout.includes(secret) && !result.stderr.includes(secret), 'secret echoed');
  checks.push({ name: 'audit-secret', expected_failure: 'redaction-audit', status: 'PASS' });
} finally {
  rmSync(auditOut, { recursive: true, force: true });
}
writeFileSync(output, JSON.stringify({ status: 'PASS', cases: checks }, null, 2) + '\n');
console.log(`PASS: ${checks.length} mutated artifact checks`);
