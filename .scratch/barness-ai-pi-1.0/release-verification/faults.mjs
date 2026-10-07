// E2E at the run.mjs CLI: replace its external network/filesystem with faulting
// boundary doubles. No verifier function is imported or tested in isolation.
import { readFileSync, writeFileSync, mkdtempSync, rmSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import assert from 'node:assert/strict';

const here = dirname(fileURLToPath(import.meta.url));
const results = [];
for (const mode of ['download-overflow', 'evidence-write-failure']) {
  const testDir = mkdtempSync(join(tmpdir(), 'barness-release-boundary-'));
  let state;
  try {
    const loader = `
      import fs from 'node:fs';
      import { syncBuiltinESMExports } from 'node:module';
      const originalWrite = fs.writeFileSync;
      const originalTemp = fs.mkdtempSync;
      const state = { cancelled: false, chunks: 0, workspace: null };
      const mode = ${JSON.stringify(mode)};
      fs.mkdtempSync = (...args) => state.workspace = originalTemp(...args);
      fs.writeFileSync = (path, ...args) => {
        if (mode === 'evidence-write-failure' && String(path).endsWith('/commands.json'))
          throw new Error('synthetic ENOSPC');
        return originalWrite(path, ...args);
      };
      syncBuiltinESMExports();
      globalThis.fetch = async () => mode === 'evidence-write-failure'
        ? new Response('', { status: 503 })
        : new Response(new ReadableStream({
          pull(controller) {
            if (state.chunks === 32) return controller.close();
            state.chunks++; controller.enqueue(new Uint8Array(1024 * 1024));
          },
          cancel() { state.cancelled = true; },
        }));
      process.on('exit', () => originalWrite(${JSON.stringify(join(testDir, 'state.json'))},
        JSON.stringify({ ...state, removed: state.workspace !== null && !fs.existsSync(state.workspace) })));
    `;
    const result = spawnSync(process.execPath, ['--import', 'data:text/javascript,' + encodeURIComponent(loader),
      join(here, 'run.mjs'), join(testDir, 'evidence')], { encoding: 'utf8', timeout: 120_000 });
    state = JSON.parse(readFileSync(join(testDir, 'state.json')));
    assert.notEqual(result.status, 0, mode);
    assert.equal(state.removed, true, `${mode}: temporary workspace leaked`);
    if (mode === 'download-overflow') {
      assert.equal(state.cancelled, true, 'oversized response was not cancelled');
      assert.ok(state.chunks <= 22, 'response was fully buffered before the limit');
      assert.equal(JSON.parse(readFileSync(join(testDir, 'evidence/report.json'))).status, 'FAIL');
    } else {
      const report = JSON.parse(readFileSync(join(testDir, 'evidence/report.json')));
      assert.equal(report.status, 'FAIL');
      assert.equal(report.failure, 'evidence-finalization', 'write failure did not invalidate the report');
    }
    results.push({ name: mode, status: 'PASS' });
  } finally {
    // Clean up even a regression's leaked workspace after recording the failure.
    if (state?.workspace) rmSync(state.workspace, { recursive: true, force: true });
    rmSync(testDir, { recursive: true, force: true });
  }
}
writeFileSync(resolve(process.argv[2]), JSON.stringify({ status: 'PASS', cases: results }, null, 2) + '\n');
console.log(`PASS: ${results.length} boundary fault checks`);
