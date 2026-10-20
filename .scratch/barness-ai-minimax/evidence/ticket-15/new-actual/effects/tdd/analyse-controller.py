from pathlib import Path
import sys,json
p=Path('/tmp/barness-minimax-routed-validation/.scratch/barness-ai-minimax/evidence/ticket-15/new-actual/effects').resolve();sys.path.insert(0,str(p));import verify as v
b=v.readers.small();i=b.doc((p/'index.json').read_bytes());a,m=b.verify_archive(Path('/Users/cyber/RestoX/barness'),i['capsule']);s,sc=v.analyse(b,i,m);more=v.secondary(b,i,Path('/Users/cyber/RestoX/barness'),m)
for name,value in [('summary.json',s),('secondary-summary.json',more)]: (p/name).write_text(json.dumps(value,ensure_ascii=False,separators=(',',':'))+'\n')
(p/'scenarios.jsonl').write_text(''.join(json.dumps(x,ensure_ascii=False,separators=(',',':'))+'\n' for x in sc));print(json.dumps(dict(calls=s['calls'],attempts=s['attempts'],usage=s['usageReporting'],knownInputSide=s['knownInputSide'],knownOutput=s['knownOutput'],secondary=more),indent=2))
