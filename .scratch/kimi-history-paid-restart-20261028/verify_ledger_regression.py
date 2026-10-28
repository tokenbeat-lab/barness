"""Offline subprocess regression for the actual two entry-point ledger loops.

Only the ledger-reading loop is compiled, so STOP/scope/producer code is never
executed. Mutation cases own isolated files; retained Kimi records are read-only.
"""
import argparse
import ast
import hashlib
import json
from pathlib import Path
import sys

HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(HERE))
import common


def worker(args):
    common.R=Path(args.root).resolve()
    tree=ast.parse(Path(args.entry).read_text())
    loops=[node for node in ast.walk(tree) if isinstance(node,ast.For)
           and isinstance(node.iter,ast.Subscript)
           and isinstance(node.iter.value,ast.Name) and node.iter.value.id=='lp'
           and isinstance(node.iter.slice,ast.Constant) and node.iter.slice.value=='records']
    if len(loops)!=1:
        raise ValueError('expected exactly one actual ledger-reading loop')
    proof=json.loads(Path(args.proof).read_text())
    namespace={**vars(common),'lp':proof}
    exec(compile(ast.Module(body=loops,type_ignores=[]),args.entry,'exec'),namespace)
    print(json.dumps({'checked':len(proof['records'])}))


def main(args):
    root=common.R
    out=Path(args.out).resolve();out.mkdir()
    env,_=common.environment('e2e')
    entries=[HERE/'preflight.py',HERE/'offline_review.py']
    selected=common.load(root/'.evidence/provider-delivery-offline/selected-evidence-v1.json')
    results=[]
    names=['path','reference','path-copy-mismatch','reference-copy-mismatch',
           'path-authority-mismatch','reference-authority-mismatch',
           'reference-missing-path','reference-missing-hash','invalid-copy-type',
           'reference-invalid-path-type','path-missing-hash']
    for entry in entries:
        for name in names:
            case=out/entry.stem/name;case.mkdir(parents=True)
            fixture=case/'fixture';fixture.mkdir()
            authority=fixture/'.git/barness-kimi-grants/record.json'
            authority.parent.mkdir(parents=True)
            copy=fixture/'retained.json';payload=b'isolated ledger regression fixture\n'
            authority.write_bytes(payload);copy.write_bytes(payload)
            digest=hashlib.sha256(payload).hexdigest()
            reference=name.startswith('reference')
            record={'trustedGitCommonRelativePath':'barness-kimi-grants/record.json',
                    'preservedCopy':{'path':'retained.json','sha256':digest} if reference else 'retained.json'}
            if not reference:record['sha256']=digest
            if 'copy-mismatch' in name:copy.write_bytes(b'changed retained copy\n')
            if 'authority-mismatch' in name:authority.write_bytes(b'changed trusted record\n')
            if name=='reference-missing-path':del record['preservedCopy']['path']
            if name=='reference-missing-hash':del record['preservedCopy']['sha256']
            if name=='invalid-copy-type':record['preservedCopy']=42
            if name=='reference-invalid-path-type':record['preservedCopy']['path']=[]
            if name=='path-missing-hash':del record['sha256']
            proof=case/'proof.json';common.save(proof,{'records':[record]})
            results.append(run_case(entry,name,proof,fixture,case,env,name in ['path','reference']))
        for row in selected['kimi']:
            proof=common.check_ref(row['original']['ledgerProof'])
            name='original-'+row['route'];case=out/entry.stem/name;case.mkdir(parents=True)
            results.append(run_case(entry,name,proof,root,case,env,True))
    passed=all(x['pass'] for x in results)
    common.save(out/'report.json',dict(pass_=passed,results=results,
        entryFiles={str(p.relative_to(root)):common.sha(p) for p in entries},
        common=common.ref(HERE/'common.py'),credentialsRead=0,grantsIssued=0,vendorCalls=0))
    print(json.dumps({'pass':passed,'checks':len(results),
        'failed':[x['entry']+'/'+x['case'] for x in results if not x['pass']],
        'report':str((out/'report.json').relative_to(root))}))
    return 0 if passed else 1


def run_case(entry,name,proof,root,case,env,expected_success):
    argv=[sys.executable,str(Path(__file__).resolve()),'--worker','--entry',str(entry),
          '--proof',str(proof),'--root',str(root)]
    result=common.capture(argv,common.R,env,case,20,[])
    success=result['status']=='PASS' and result['exitCode']==0
    value={'entry':entry.stem,'case':name,'expectedSuccess':expected_success,
           'pass':success==expected_success,'argv':argv,'proof':common.ref(proof),**result}
    if not expected_success:
        try:
            error=json.loads((case/'stderr.log').read_text())
            value['pass']=value['pass'] and error.get('errorType') in ['ValueError','AssertionError']
        except ValueError:
            value['pass']=False
    value['files']={p.name:common.sha(p) for p in case.iterdir() if p.is_file()}
    common.save(case/'receipt.json',value)
    return value


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out');parser.add_argument('--worker',action='store_true')
    parser.add_argument('--entry');parser.add_argument('--proof');parser.add_argument('--root')
    args=parser.parse_args()
    if args.worker:
        try:worker(args)
        except Exception as error:
            print(json.dumps({'errorType':type(error).__name__,'message':str(error)}),file=sys.stderr)
            sys.exit(1)
    else:sys.exit(main(args))
