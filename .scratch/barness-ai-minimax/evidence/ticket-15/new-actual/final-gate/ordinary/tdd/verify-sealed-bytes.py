from pathlib import Path
import sys,json
base=Path(__file__).resolve().parent.parent;sys.path.insert(0,str(base));import readers
b=readers.small();index=b.doc((base/'index.json').read_bytes());pins,m,archive=readers.full().verify(b,base,Path('/Users/cyber/RestoX/barness'),index);print(json.dumps(dict(verified=True,archive=archive,sourceCommit=index['sourceCommit'],credentialsRead=False,paidCalls=0),indent=2))
