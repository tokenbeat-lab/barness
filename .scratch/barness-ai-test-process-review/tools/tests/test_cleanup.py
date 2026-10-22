from datetime import datetime, timedelta, timezone
import json
import sys
from process_support import ProcessCase, TOOLS, write_json, digest


class CleanupProcess(ProcessCase):
    def setUp(self):
        super().setUp()
        (self.repo / '.gitignore').write_text('/.evidence/\n')
        for p in ('recovery7-preparation/index.json', 'recovery6-stopped/index.json', 'closed-cohort-6/seal.json'):
            write_json(self.repo / '.evidence/provider-sites/final-acceptance-20261020' / p, {'synthetic': True})
        write_json(self.repo / '.evidence/minimax-retained/receipt.json', {'status': 'PASS', 'exitCode': 0})
        (self.repo / '.evidence/unused').mkdir(parents=True)
        (self.repo / '.evidence/unused/body.txt').write_text('synthetic body\n')
        write_json(self.repo / '.evidence/unused/receipt.json', {'status': 'PASS', 'exitCode': 0, 'files': {'body.txt': digest(self.repo / '.evidence/unused/body.txt')}})

    def tool(self, *args, **kwargs):
        return self.call(sys.executable, TOOLS / 'cleanup.py', *args, **kwargs)

    def prepare(self):
        result = self.tool('scan', '--repo', self.repo, expected=0)
        self.batch = json.loads(result.stdout)['batch']
        self.meta = self.repo / '.evidence/_quarantine' / self.batch / 'metadata'
        result = self.tool('plan', '--batch', self.batch, expected=0)
        self.plan_sha = json.loads(result.stdout)['planSHA256']

    def move_verify(self):
        self.prepare()
        self.tool('move', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=0)
        self.tool('verify', '--batch', self.batch, expected=0)

    def age(self):
        path = self.meta / 'state.json'
        state = json.loads(path.read_text())
        state['verifiedAt'] = (datetime.now(timezone.utc) - timedelta(days=8)).isoformat()
        state['lastMutationAt'] = state['verifiedAt']
        write_json(path, state)

    def test_relative_log_and_indirect_index_remain_protected(self):
        base = self.repo / '.scratch/example/evidence'
        (base / 'tdd').mkdir(parents=True)
        (base / 'tdd/final.log').write_text('linked\n')
        (base / 'handoff.md').write_text('[log](tdd/final.log)\n[run](../../../.evidence/used/index.json)\n')
        write_json(self.repo / '.evidence/used/index.json', {'path': '../indirect/receipt.json'})
        write_json(self.repo / '.evidence/indirect/receipt.json', {'status': 'PASS', 'exitCode': 0})
        self.move_verify()
        self.assertTrue((base / 'tdd/final.log').is_file())
        self.assertTrue((self.repo / '.evidence/indirect/receipt.json').is_file())
        self.assertFalse((self.repo / '.evidence/unused').exists())

    def test_existing_missing_and_examples_do_not_become_new_damage(self):
        (self.repo / 'notes.md').write_text('Historical `.evidence/old/missing.json`.\nExample `.evidence/NEW_RELEASE_CHECK`.\n')
        self.move_verify()
        existing = json.loads((self.meta / 'preexisting-missing.json').read_text())
        self.assertTrue(any('old/missing.json' in row['path'] for row in existing))

    def test_new_reference_and_wrong_hash_block_move(self):
        self.prepare()
        self.tool('move', '--batch', self.batch, '--plan-sha256', '0' * 64, expected=1)
        (self.repo / 'new.md').write_text('Retain `.evidence/unused/body.txt`.\n')
        self.tool('move', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=1)
        self.assertTrue((self.repo / '.evidence/unused/body.txt').exists())

    def test_purge_requires_observation_and_fresh_references(self):
        self.move_verify()
        self.tool('purge', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=1)
        self.age()
        (self.repo / 'new.md').write_text('Restore `.evidence/unused/body.txt`.\n')
        self.tool('purge', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=1)
        payload = self.meta.parent / 'payload/.evidence/unused/body.txt'
        self.assertTrue(payload.exists())
        self.tool('restore', '--batch', self.batch, expected=0)
        self.assertTrue((self.repo / '.evidence/unused/body.txt').exists())

    def test_restore_never_overwrites_changed_original(self):
        self.move_verify()
        write_json(self.repo / '.evidence/unused/receipt.json', {'new': 'different'})
        self.tool('restore', '--batch', self.batch, expected=1)
        self.assertEqual(json.loads((self.repo / '.evidence/unused/receipt.json').read_text()), {'new': 'different'})

    def test_valid_aged_batch_purges_only_payload(self):
        self.move_verify()
        self.age()
        self.tool('purge', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=0)
        self.assertFalse((self.meta.parent / 'payload').exists())
        self.assertTrue((self.meta / 'refs.jsonl').exists())
        self.assertTrue((self.repo / '.evidence/minimax-retained/receipt.json').exists())

    def test_lock_and_missing_required_file_block_scan(self):
        lock = self.repo / '.evidence/_quarantine/LOCK'
        lock.parent.mkdir(parents=True)
        lock.write_text('owned\n')
        self.tool('scan', '--repo', self.repo, expected=1)
        lock.unlink()
        required = self.repo / '.evidence/provider-sites/final-acceptance-20261020/closed-cohort-6/seal.json'
        required.unlink()
        self.tool('scan', '--repo', self.repo, expected=1)

    def test_missing_current_dependency_blocks_before_any_movement(self):
        required = self.repo / '.evidence/provider-sites/final-acceptance-20261020/recovery7-preparation/index.json'
        write_json(required, {'path': '../missing-current/index.json'})
        self.tool('scan', '--repo', self.repo, expected=1)
        self.assertTrue((self.repo / '.evidence/unused/body.txt').exists())

    def test_mixed_success_never_hides_failed_unknown_or_authority_contents(self):
        for name in ('failed', 'unknown', 'authority', 'no-audit'):
            base = self.repo / '.evidence' / ('mixed-' + name)
            write_json(base / 'good/receipt.json', {'status': 'PASS', 'exitCode': 0})
            if name == 'failed':
                write_json(base / 'old/manifest.json', {'cases': [{'id': 'original-failure', 'status': 'FAIL'}]})
            elif name == 'unknown':
                (base / 'unclassified.bin').write_bytes(b'unknown payload')
            elif name == 'no-audit':
                write_json(base / 'native/manifest.json', {'cases': [{'id': 'case', 'status': 'PASS', 'artifacts': {}}]})
            else:
                write_json(base / 'grant.json', {'id': 'synthetic-authority', 'status': 'closed'})
        self.prepare()
        plan = json.loads((self.meta / 'plan.json').read_text())
        marks = {c['path']: c['mark'] for c in plan['candidates']}
        for name in ('failed', 'unknown', 'authority', 'no-audit'):
            self.assertEqual(marks['.evidence/mixed-' + name], 'BLOCKED')
        self.assertEqual(marks['.evidence/unused'], 'QUARANTINE')

    def test_unindexed_failed_run_retains_transitive_dependencies(self):
        write_json(self.repo / '.evidence/unindexed-failure/receipt.json', {
            'status': 'FAIL', 'exitCode': 7, 'input': '../dependency/index.json'})
        write_json(self.repo / '.evidence/dependency/receipt.json', {
            'status': 'PASS', 'exitCode': 0, 'files': {}})
        dependency = self.repo / '.evidence/dependency/index.json'
        write_json(dependency, {'input': '../transitive/receipt.json'})
        write_json(self.repo / '.evidence/dependency/receipt.json', {
            'status': 'PASS', 'exitCode': 0, 'files': {'index.json': digest(dependency)}})
        write_json(self.repo / '.evidence/transitive/receipt.json', {'status': 'PASS', 'exitCode': 0})
        self.prepare()
        plan = json.loads((self.meta / 'plan.json').read_text())
        marks = {c['path']: c['mark'] for c in plan['candidates']}
        self.assertNotEqual(marks['.evidence/unindexed-failure'], 'QUARANTINE')
        self.assertEqual(marks['.evidence/dependency'], 'KEEP')
        self.assertEqual(marks['.evidence/transitive'], 'KEEP')
        snapshot = json.loads((self.meta / 'snapshot.json').read_text())
        self.assertIn('.evidence/unindexed-failure/receipt.json', snapshot['sources'])

    def test_state_write_failure_after_move_preserves_restore(self):
        import resource
        import signal
        self.prepare()
        self.tool('move', '--batch', self.batch, '--plan-sha256', self.plan_sha, expected=0)
        path = self.meta / 'state.json'
        before = path.read_bytes()
        def bounded_disk_failure():
            signal.signal(signal.SIGXFSZ, signal.SIG_IGN)
            resource.setrlimit(resource.RLIMIT_FSIZE, (32, 32))
        self.tool('verify', '--batch', self.batch, expected=1, child_setup=bounded_disk_failure)
        self.assertEqual(path.read_bytes(), before, 'failed verification destroyed the last durable MOVED state')
        self.assertEqual(json.loads(path.read_text())['status'], 'MOVED')
        self.tool('restore', '--batch', self.batch, expected=0)
        self.assertEqual((self.repo / '.evidence/unused/body.txt').read_text(), 'synthetic body\n')
