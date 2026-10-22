import json
import sys
sys.path.insert(0, '/Users/cyber/RestoX/barness/.scratch/barness-ai-test-process-review/tools/tests')
from process_support import write_json
from test_cleanup import CleanupProcess
case = CleanupProcess('test_unindexed_failed_run_retains_transitive_dependencies')
case.setUp()
try:
    failure = case.repo / '.evidence/unindexed-failure/receipt.json'
    write_json(failure, {'status': 'FAIL', 'exitCode': 7, 'input': '../unused'})
    case.prepare()
    plan = json.loads((case.meta / 'plan.json').read_text())
    snapshot = json.loads((case.meta / 'snapshot.json').read_text())
    print(json.dumps({'phase': 'plan', 'marks': {c['path']: c['mark'] for c in plan['candidates']}, 'dependency_reference_rows': [r for r in snapshot['refs'] if r['source'] == '.evidence/unindexed-failure/receipt.json']}))
    case.tool('move', '--batch', case.batch, '--plan-sha256', case.plan_sha, expected=0)
    case.tool('verify', '--batch', case.batch, expected=0)
    case.age()
    case.tool('purge', '--batch', case.batch, '--plan-sha256', case.plan_sha, expected=0)
    print(json.dumps({'phase': 'purged', 'failure_retained': failure.exists(), 'referenced_directory_exists': (case.repo / '.evidence/unused').exists(), 'verified_status': json.loads((case.meta / 'state.json').read_text())['status']}))
finally:
    case.doCleanups()
