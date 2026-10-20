#!/usr/bin/env python3
"""Fixed provider-sites-final-acceptance-2 operator; no old authority is reused."""
import argparse
import json
from pathlib import Path
import sys
sys.dont_write_bytecode = True
import final_contract as r
c = r.c


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=('inventory', 'prepare', 'preview', 'run-selected', 'check', 'summarize', 'final-plan', 'finalize', 'native-grant', 'promote', 'prepare-responses-retry', 'run-responses-retry'))
    parser.add_argument('out', nargs='?')
    parser.add_argument('route', nargs='?')
    parser.add_argument('--phase', choices=r.PHASES)
    parser.add_argument('--inputs')
    parser.add_argument('--execution-root')
    a = parser.parse_args()
    if a.mode == 'inventory':
        result = r.inventory(c.ROOT / '.evidence/provider-sites/final-acceptance-20261020/unprepared')
    elif a.mode == 'summarize':
        import final_execute
        result = final_execute.summarize()
    else:
        import final_state as state
        out = state.path(a.out)
        if a.mode == 'prepare-responses-retry':
            if not a.execution_root:raise RuntimeError('fixed original-commit execution checkout required')
            result=state.prepare(out,'glm',a.execution_root)
        elif a.mode=='run-responses-retry':
            import final_retry,final_execute
            result=final_execute.run_selected(out,final_retry.jobs(out)[final_retry.NAME])
        elif a.mode == 'prepare':
            if a.phase is None:
                raise RuntimeError('prepare requires one fixed phase')
            result = state.prepare(out, a.phase)
        elif a.mode == 'native-grant':
            import final_minimax
            result = final_minimax.create(out)
        elif a.mode == 'promote':
            import final_native
            result = final_native.promote(out)
        elif a.mode in ('final-plan', 'finalize'):
            if not a.inputs:
                raise RuntimeError('all23+four exact inputs required')
            import final_finalize
            result = final_finalize.plan(out,a.inputs) if a.mode == 'final-plan' else final_finalize.finalize(out,a.inputs)
        elif a.mode == 'check':
            import final_execute
            result = final_execute.checks(out)
        else:
            prepared = state.guard(out)
            jobs = state.phase_plan(out,prepared['phase'])['routes']
            if a.route not in jobs:
                raise RuntimeError('one fixed own job in this phase required')
            job = jobs[a.route]
            if a.mode == 'preview':
                result = job | dict(credentialsRead=False)
            else:
                import final_execute
                result = final_execute.run_selected(out, job)
    print(json.dumps(c.redact(result), ensure_ascii=False, indent=2))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('REFUSED: ' + str(c.redact(str(error))), file=sys.stderr)
        sys.exit(1)
