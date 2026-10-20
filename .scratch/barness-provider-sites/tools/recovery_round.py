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
        closed=c.common_root()/r.BASE/'closed-cohort-3/seal.json'
        if a.mode!='summarize' and (closed.exists() or closed.is_symlink()):raise RuntimeError('recovery3 permanently closed; history only')
        from recovery_flow import fixed3
        flow=fixed3();s=flow.s;e=flow.e
        if a.mode=='summarize':
            if closed.exists():
                import recovery4_history
                result=recovery4_history.facts()
            else:result=e.summarize()
        else:
            out=s.path(a.out)
            if a.mode=='prepare':result=s.prepare(out)
            elif a.mode=='check':result=s.checks(out)
            elif a.mode=='native-grant':
                result=flow.authority.create(out)
            elif a.mode in ('final-plan','finalize'):
                result=flow.f.plan(out,a.inputs) if a.mode=='final-plan' else flow.f.finalize(out,a.inputs)
            else:
                s.guard(out)
                if a.route not in r.jobs(out):raise RuntimeError('one exact fixed recovery3 job required')
                result=r.jobs(out)[a.route] if a.mode=='preview' else e.run_selected(out,r.jobs(out)[a.route])
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))
if __name__=='__main__':
    try:main()
    except Exception as error:print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
