#!/usr/bin/env python3
"""Fixed recovery3; no ambient round override, paid restart or old2 authority."""
import argparse
import json
import sys
sys.dont_write_bytecode=True
import recovery_contract as r
c=r.c

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('mode',choices=('inventory','prepare','preview','check','summarize','native-grant','run-selected','final-plan','finalize'))
    p.add_argument('out',nargs='?');p.add_argument('route',nargs='?');p.add_argument('--inputs');a=p.parse_args()
    if a.mode=='inventory':result=r.inventory(c.ROOT/'.evidence/provider-sites/unprepared-recovery3')
    else:
        import recovery_state as s
        import recovery_execute as e
        if a.mode=='summarize':result=e.summarize()
        else:
            out=s.path(a.out)
            if a.mode=='prepare':result=s.prepare(out)
            elif a.mode=='check':result=s.checks(out)
            elif a.mode=='native-grant':
                import recovery_native
                result=recovery_native.create(out)
            elif a.mode in ('final-plan','finalize'):
                import recovery_finalize as f
                result=f.plan(out,a.inputs) if a.mode=='final-plan' else f.finalize(out,a.inputs)
            else:
                s.guard(out)
                if a.route not in r.jobs(out):raise RuntimeError('one exact fixed recovery3 job required')
                result=r.jobs(out)[a.route] if a.mode=='preview' else e.run_selected(out,r.jobs(out)[a.route])
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))
if __name__=='__main__':
    try:main()
    except Exception as error:print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
