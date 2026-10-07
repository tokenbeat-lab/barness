// Issue 02's artifact boundary: inspect a disposable release/source workspace.
// run.mjs supplies it; this command never runs a model or refreshes a catalog.
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync, readdirSync, mkdirSync, lstatSync, existsSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const here = dirname(fileURLToPath(import.meta.url));
const work = resolve(process.argv[2]);
const out = resolve(process.argv[3]);
const pins = JSON.parse(readFileSync(join(here, 'inputs.json')));
const sha = b => createHash('sha256').update(b).digest('hex');
const json = p => JSON.parse(readFileSync(p));
const write = (name, value) => writeFileSync(join(out, name), JSON.stringify(value, null, 2) + '\n');
let stage = 'inputs';
const report = { status: 'FAIL', scope: 'release-artifact-equivalence', barness_1_0_differential: 'NOT_RUN' };
function requireThat(ok) { if (!ok) throw new Error(stage); }

// Reject symlinks and unexpected objects rather than comparing their targets.
function files(root, prefix = '') {
  return readdirSync(join(root, prefix)).sort().flatMap(name => {
    const relative = prefix ? `${prefix}/${name}` : name;
    const stat = lstatSync(join(root, relative));
    requireThat(stat.isDirectory() || stat.isFile());
    return stat.isDirectory() ? files(root, relative) : [relative];
  });
}
function git(...args) {
  const result = spawnSync('git', ['-C', join(work, 'source'), ...args], {
    env: { PATH: process.env.PATH, GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: '/dev/null' },
    encoding: 'utf8', timeout: 30_000,
  });
  requireThat(result.status === 0);
  return result.stdout.trim();
}
function lockInventory(root, expectedHash) {
  const bytes = readFileSync(join(root, 'package-lock.json'));
  const lock = JSON.parse(bytes);
  requireThat(lock.lockfileVersion === 3 && sha(bytes) === expectedHash);
  const entries = [];
  for (const [path, entry] of Object.entries(lock.packages)) {
    if (!path.includes('node_modules/') || entry.link) continue;
    requireThat(/^\d+\.\d+\.\d+(?:[-+].*)?$/.test(entry.version ?? '') &&
      entry.resolved?.startsWith('https://registry.npmjs.org/') &&
      /^sha512-[A-Za-z0-9+/]+=*$/.test(entry.integrity ?? ''));
    const installed = existsSync(join(root, path, 'package.json'));
    // The frozen upstream lock contains four stale extraneous entries; npm ci
    // omits them. Preserve and report them, without installing or editing them.
    requireThat(installed || entry.optional === true || entry.devOptional === true || entry.extraneous === true);
    if (installed) requireThat(json(join(root, path, 'package.json')).version === entry.version);
    entries.push({ path, version: entry.version, resolved: entry.resolved, integrity: entry.integrity,
      optional: entry.optional === true || entry.devOptional === true, extraneous: entry.extraneous === true, installed });
  }
  return { sha256: sha(bytes), external_packages: entries.length, entries };
}

mkdirSync(out, { recursive: true });
try {
  requireThat(sha(readFileSync(join(here, 'model-data.sha256'))) === pins.model_data_sha256);
  stage = 'version';
  for (const suffix of ['ai', 'telemetry']) {
    const name = `@earendil-works/pi-${suffix}`;
    requireThat(json(join(work, `registry-${suffix}.json`)).version === '1.0.0');
    for (const dir of [`release/node_modules/${name}`, `source/packages/${suffix}`]) {
      const pkg = json(join(work, dir, 'package.json'));
      requireThat(pkg.name === name && pkg.version === '1.0.0');
    }
  }
  stage = 'source-identity';
  const head = git('rev-parse', 'HEAD');
  const tagCommit = git('rev-parse', `refs/tags/${pins.tag}^{commit}`);
  requireThat(head === pins.commit && tagCommit === pins.commit && git('status', '--porcelain') === '');
  const packages = [];
  for (const suffix of ['ai', 'telemetry']) {
    const meta = json(join(work, `registry-${suffix}.json`));
    const expected = pins.packages.find(p => p.name === `@earendil-works/pi-${suffix}`);
    requireThat(meta.name === expected.name && meta.gitHead === pins.commit && meta.gitHead === expected.gitHead);
    packages.push(meta);
  }
  report.source = { tag: pins.tag, target_commit: pins.commit, checked_out_commit: head, tag_commit: tagCommit,
    changed_ai_telemetry_files: [], explanation: 'Both npm gitHeads, the tag and the target commit are identical.' };
  stage = 'tarball-integrity';
  for (const suffix of ['ai', 'telemetry']) {
    const expected = pins.packages.find(p => p.name === `@earendil-works/pi-${suffix}`);
    const meta = packages.find(p => p.name === expected.name);
    requireThat(meta.dist.tarball === expected.dist.tarball && meta.dist.integrity === expected.dist.integrity);
    const bytes = readFileSync(join(work, `pi-${suffix}-1.0.0.tgz`));
    requireThat('sha512-' + createHash('sha512').update(bytes).digest('base64') === expected.dist.integrity);
  }
  write('packages.json', packages);
  stage = 'dependency-lock';
  const release = lockInventory(join(work, 'release'), pins.release_lock_sha256);
  const source = lockInventory(join(work, 'source'), pins.source_lock_sha256);
  const pkg = json(join(work, 'release/package.json'));
  for (const p of pins.packages) requireThat(pkg.dependencies[p.name] === '1.0.0');
  requireThat(pkg.overrides['@earendil-works/pi-telemetry'] === '1.0.0');
  write('dependencies.json', { release, source });
  report.dependencies = { release_lock_sha256: release.sha256, source_lock_sha256: source.sha256,
    release_packages: release.external_packages, source_packages: source.external_packages,
    source_optional_not_installed: source.entries.filter(p => !p.installed && p.optional).length,
    source_extraneous_not_installed: source.entries.filter(p => !p.installed && p.extraneous).length };
  stage = 'model-data';
  const dataRoot = join(work, 'release/node_modules/@earendil-works/pi-ai/dist/providers/data');
  const names = files(dataRoot);
  const dataList = names.map(name => `${sha(readFileSync(join(dataRoot, name)))}  ${name}\n`).join('');
  requireThat(dataList === readFileSync(join(here, 'model-data.sha256'), 'utf8') && sha(dataList) === pins.model_data_sha256);
  const dataManifest = json(join(dataRoot, '.manifest.json'));
  report.model_data = { files: names.length, sha256: sha(dataList), schema_version: dataManifest.schemaVersion,
    generated_at: dataManifest.generatedAt, structure_hash: dataManifest.structureHash, origin: 'published dist/providers/data; no generation' };
  writeFileSync(join(out, 'model-data.sha256'), dataList);
  stage = 'build-equivalence';
  report.comparisons = [];
  for (const suffix of ['ai', 'telemetry']) {
    const published = join(work, `release/node_modules/@earendil-works/pi-${suffix}/dist`);
    const rebuilt = join(work, `source/packages/${suffix}/dist`);
    const publishedNames = files(published);
    const rebuiltNames = files(rebuilt);
    const all = [...new Set([...publishedNames, ...rebuiltNames])].sort();
    const differences = [];
    const inventory = [];
    for (const name of all) {
      const a = existsSync(join(published, name)) ? readFileSync(join(published, name)) : null;
      const b = existsSync(join(rebuilt, name)) ? readFileSync(join(rebuilt, name)) : null;
      if (!a || !b || !a.equals(b)) differences.push(name);
      inventory.push(`${a ? sha(a) : 'MISSING'}  ${b ? sha(b) : 'MISSING'}  ${name}\n`);
    }
    writeFileSync(join(out, `dist-${suffix}.sha256`), inventory.join(''));
    report.comparisons.push({ package: suffix, scope: 'entire dist tree including hidden files, declarations, maps and model data',
      published_files: publishedNames.length, rebuilt_files: rebuiltNames.length,
      compared_files: all.length, differences: differences.length, different_files: differences });
  }
  requireThat(report.comparisons.every(c => c.compared_files > 0 && c.differences === 0));
  report.status = 'PASS';
} catch {
  report.status = 'FAIL';
  report.failure = stage;
  process.exitCode = 1;
} finally {
  write('report.json', report);
  console.log(`${report.status}: ${report.failure ?? 'release artifacts are byte-identical'}`);
}
