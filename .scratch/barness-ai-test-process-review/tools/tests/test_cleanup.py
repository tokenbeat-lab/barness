from datetime import datetime, timedelta, timezone
import json
import sys
from process_support import ProcessCase, TOOLS, write_json


class CleanupProcess(ProcessCase):
    def setUp(self):
        super().setUp()
        (self.repo / '.gitignore').write_text('/.evidence/\n')
        for p in ('recovery7-preparation/index.json', 'recovery6-stopped/index.json', 'closed-cohort-6/seal.json'):
            write_json(self.repo / '.evidence/provider-sites/final-acceptance-20261020' / p, {'synthetic': True})
        write_json(self.repo / '.evidence/minimax-retained/receipt.json', {'status': 'PASS', 'exitCode': 0})
        write_json(self.repo / '.evidence/unused/receipt.json', {'status': 'PASS', 'exitCode': 0})
        (self.repo / '.evidence/unused/body.txt').write_text('synthetic body\n')

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
