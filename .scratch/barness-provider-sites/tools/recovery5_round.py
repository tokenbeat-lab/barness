#!/usr/bin/env python3
"""The fixed fifth recovery only; pending approval is never paid permission."""
import argparse
import json
import sys
sys.dont_write_bytecode=True
import recovery5_contract as r
c=r.c

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode',choices=('inventory','prepare','preview','check','summarize','native-grant','run-selected','final-plan','finalize'))
    parser.add_argument('out',nargs='?');parser.add_argument('route',nargs='?');parser.add_argument('--inputs');args=parser.parse_args()
    if args.mode=='inventory':result=r.inventory(c.ROOT/'.evidence/provider-sites/unprepared-recovery5')
    else:
        closed=c.common_root()/r.BASE/'closed-cohort-5/seal.json'
        if args.mode!='summarize' and (closed.exists() or closed.is_symlink()):raise RuntimeError('recovery5 permanently closed; history only')
        from recovery_flow import fixed5
        flow=fixed5()
        if args.mode=='summarize':
            if closed.exists():
                import recovery6_history
                result=recovery6_history.facts()
            else:result=flow.e.summarize()
        else:
            out=flow.s.path(args.out)
            if args.mode=='prepare':result=flow.s.prepare(out)
            elif args.mode=='check':result=flow.s.checks(out)
            elif args.mode=='native-grant':result=flow.authority.create(out)
            elif args.mode in ('final-plan','finalize'):
                result=flow.f.plan(out,args.inputs) if args.mode=='final-plan' else flow.f.finalize(out,args.inputs)
            else:
                flow.s.guard(out)
                if args.route not in r.jobs(out):raise RuntimeError('one exact fixed5 job required')
                result=r.jobs(out)[args.route] if args.mode=='preview' else flow.e.run_selected(out,r.jobs(out)[args.route])
    print(json.dumps(c.redact(result),ensure_ascii=False,indent=2))
if __name__=='__main__':
    try:main()
    except Exception as error:print('REFUSED: '+str(c.redact(str(error))),file=sys.stderr);sys.exit(1)
