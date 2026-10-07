// Disposable equivalence verification per ADR-0004. No model generation or
// oracle execution occurs. All installs skip lifecycle scripts.
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync, cpSync, mkdirSync, mkdtempSync, readdirSync, rmSync, existsSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { tmpdir, homedir, hostname } from 'node:os';
import { spawnSync } from 'node:child_process';

const here = dirname(fileURLToPath(import.meta.url));
const repo = resolve(here, '../../..');
const pins = JSON.parse(readFileSync(join(here, 'inputs.json')));
if (process.argv.length !== 3) throw new Error('usage: node run.mjs <new-evidence-directory>');
const out = resolve(process.argv[2]);
if (existsSync(out)) throw new Error('evidence directory must be new');
mkdirSync(out, { recursive: true });
const work = mkdtempSync(join(tmpdir(), 'barness-pi-release-'));
const logs = [];
const commands = [];
let stage = 'setup';
let pass = false;
const sha = b => createHash('sha256').update(b).digest('hex');
const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');
const env = { PATH: process.env.PATH, TMPDIR: work, LC_ALL: 'C',
  GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: join(work, 'empty-git-config'),
  GIT_TERMINAL_PROMPT: '0', NPM_CONFIG_USERCONFIG: join(work, 'empty-npm-userconfig'),
  NPM_CONFIG_GLOBALCONFIG: join(work, 'empty-npm-globalconfig'), NPM_CONFIG_CACHE: join(work, 'npm-cache'),
  NPM_CONFIG_REGISTRY: 'https://registry.npmjs.org/', NPM_CONFIG_IGNORE_SCRIPTS: 'true' };
// The audit gets the real environment to detect secrets. Install/build commands
// get only the above allowlist, with empty user/global config and a fresh cache.
function run(label, cmd, args, cwd, options = {}) {
  stage = label;
  console.log(label);
  const result = spawnSync(cmd, args, { cwd, env, encoding: 'utf8', timeout: 300_000,
    maxBuffer: 16 * 1024 * 1024, ...options });
  commands.push({ stage: label, command: [cmd === process.execPath ? 'node' : cmd, ...args].map(s => s.replaceAll(here, '<inputs>').replaceAll(work, '<workspace>').replaceAll(out, '<evidence>')),
    exit_code: result.status });
  logs.push(result.stdout ?? '', result.stderr ?? '');
  if (result.status !== 0) throw new Error(label);
  return result.stdout.trim();
}
async function download(url, binary = false) {
  const response = await fetch(url, { signal: AbortSignal.timeout(45_000) });
  if (!response.ok || !response.body) {
    await response.body?.cancel();
    throw new Error(stage);
  }
  const reader = response.body.getReader();
  const chunks = [];
  let total = 0;
  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      total += value.byteLength;
      if (total > 20 * 1024 * 1024) throw new Error(stage);
      chunks.push(Buffer.from(value));
    }
  } finally {
    try { await reader.cancel(); } finally { reader.releaseLock(); }
  }
  const data = Buffer.concat(chunks, total);
  return binary ? data : JSON.parse(data);
}
function safeLog(text) {
  for (const [name, value] of Object.entries(process.env)) {
    if (/KEY|TOKEN|SECRET|PASSWORD|PASSWD|CREDENTIAL|AUTH/i.test(name) && value.length >= 6)
      text = text.replaceAll(value, '<redacted>');
  }
  for (const [value, label] of [[work, '<workspace>'], [here, '<inputs>'], [repo, '<repository>'],
    [out, '<evidence>'], [homedir(), '<home>'], [hostname(), '<host>']]) {
    if (value.length >= 6) text = text.replaceAll(value, label);
  }
  return text;
}
try {
  for (const name of ['empty-git-config', 'empty-npm-userconfig', 'empty-npm-globalconfig'])
    writeFileSync(join(work, name), '');
  stage = 'dependency-lock';
  if (sha(readFileSync(join(here, 'package-lock.json'))) !== pins.release_lock_sha256) throw new Error(stage);
  const toolchain = { node: process.version, npm: run('npm-version', 'npm', ['--version'], work),
    platform: `${process.platform}/${process.arch}` };
  for (const suffix of ['ai', 'telemetry']) {
    stage = `download-${suffix}`;
    console.log(stage);
    const expected = pins.packages.find(p => p.name === `@earendil-works/pi-${suffix}`);
    const meta = await download(`https://registry.npmjs.org/${expected.name}/1.0.0`);
    if (meta.name !== expected.name || meta.version !== '1.0.0' || meta.gitHead !== pins.commit ||
        meta.dist.integrity !== expected.dist.integrity || meta.dist.tarball !== expected.dist.tarball) throw new Error(stage);
    const selected = { name: meta.name, version: meta.version, gitHead: meta.gitHead,
      dist: Object.fromEntries(['tarball', 'integrity', 'shasum', 'fileCount', 'unpackedSize'].map(k => [k, meta.dist[k]])) };
    writeFileSync(join(work, `registry-${suffix}.json`), JSON.stringify(selected, null, 2) + '\n');
    const tarball = await download(expected.dist.tarball, true);
    if ('sha512-' + createHash('sha512').update(tarball).digest('base64') !== expected.dist.integrity) throw new Error(stage);
    writeFileSync(join(work, `pi-${suffix}-1.0.0.tgz`), tarball);
  }
  mkdirSync(join(work, 'release'));
  for (const name of ['package.json', 'package-lock.json']) cpSync(join(here, name), join(work, 'release', name));
  run('install-release-ignore-scripts', 'npm', ['ci', '--ignore-scripts', '--no-audit', '--no-fund'], join(work, 'release'));
  run('clone-frozen-tag', 'git', ['clone', '--depth', '1', '--branch', pins.tag, '--single-branch', pins.repository, 'source'], work);
  stage = 'frozen-source';
  const source = join(work, 'source');
  if (run(stage, 'git', ['rev-parse', 'HEAD'], source) !== pins.commit ||
      sha(readFileSync(join(source, 'package-lock.json'))) !== pins.source_lock_sha256) throw new Error(stage);
  run('install-source-ignore-scripts', 'npm', ['ci', '--ignore-scripts', '--no-audit', '--no-fund'], source);
  stage = 'hydrate-from-release';
  cpSync(join(work, 'release/node_modules/@earendil-works/pi-ai/dist/providers/data'),
    join(source, 'packages/ai/src/providers/data'), { recursive: true });
  const buildEnv = { ...env, NPM_CONFIG_OFFLINE: 'true' };
  for (const name of ['chord', 'pi-tui', 'pi-telemetry'])
    run(`build-${name}`, 'npm', ['run', 'build', `--workspace=@earendil-works/${name}`], source, { env: buildEnv });
  run('build-ai-offline', 'npm', ['run', 'build:offline', '--workspace=@earendil-works/pi-ai'], source, { env: buildEnv });
  toolchain.typescript = JSON.parse(readFileSync(join(source, 'node_modules/typescript/package.json'))).version;
  write('toolchain.json', toolchain);
  run('check-release-equivalence', process.execPath, [join(here, 'check.mjs'), work, out], repo);
  run('check-failure-rejections', process.execPath, [join(here, 'failures.mjs'), work, join(out, 'failure-checks.json')], repo,
    { env: process.env });
  run('check-boundary-faults', process.execPath, [join(here, 'faults.mjs'), join(out, 'boundary-checks.json')], repo,
    { env: process.env });
  pass = true;
} catch {
  write('report.json', { status: 'FAIL', failure: stage, scope: 'release-artifact-equivalence', barness_1_0_differential: 'NOT_RUN' });
  process.exitCode = 1;
} finally {
  try {
    write('commands.json', commands);
    writeFileSync(join(out, 'commands.log'), safeLog(logs.join('\n')).trimEnd() + '\n');
    // Keep a complete replay-input checksum list beside every result.
    const inputFiles = ['run.mjs', 'check.mjs', 'failures.mjs', 'faults.mjs', 'inputs.json', 'model-data.sha256', 'package.json', 'package-lock.json']
      .map(name => `.scratch/barness-ai-pi-1.0/release-verification/${name}`);
    inputFiles.push('ai/internal/testkit/audit/cmd/auditbundle/main.go');
    writeFileSync(join(out, 'inputs.sha256'), inputFiles.sort()
      .map(name => `${sha(readFileSync(join(repo, name)))}  ${name}\n`).join(''));
    write('manifest.json', { scope: 'pi-ai-1.0.0-release-equivalence', installs_ignore_scripts: true,
      builds: ['chord', 'tui', 'telemetry', 'ai:build:offline'], model_generation: 'NOT_RUN',
      workspace: 'temporary, removed after verification', replay: 'node .scratch/barness-ai-pi-1.0/release-verification/run.mjs <new-evidence-directory>' });
    // Audit even failed runs. The existing Go audit owns all redaction rules.
    const audit = spawnSync('go', ['run', './ai/internal/testkit/audit/cmd/auditbundle', out], {
      cwd: repo, env: process.env, encoding: 'utf8', timeout: 120_000,
    });
    if (audit.status !== 0) {
      pass = false;
      process.exitCode = 1;
      write('report.json', { status: 'FAIL', failure: 'redaction-audit', scope: 'release-artifact-equivalence', barness_1_0_differential: 'NOT_RUN' });
    }
    writeFileSync(join(out, 'SHA256SUMS'), readdirSync(out).filter(name => name !== 'SHA256SUMS').sort()
      .map(name => `${sha(readFileSync(join(out, name)))}  ${name}\n`).join(''));
  } catch {
    // A failed finalization must invalidate an earlier equivalence PASS too.
    pass = false;
    process.exitCode = 1;
    stage = 'evidence-finalization';
    write('report.json', { status: 'FAIL', failure: stage, scope: 'release-artifact-equivalence', barness_1_0_differential: 'NOT_RUN' });
  } finally {
    // Evidence writes and auditing can also fail; cleanup is unconditional.
    rmSync(work, { recursive: true, force: true });
  }
  console.log(pass ? 'PASS: audited release equivalence and failure checks' : `FAIL: ${stage}`);
}
