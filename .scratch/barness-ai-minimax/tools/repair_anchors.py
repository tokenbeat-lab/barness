"""Certified original inventories are independent of newly prepared reuse hashes."""
import gzip,io
import paid_core as c
import routed_evidence as e
import renewal_common as u
LEGACY='.scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/'
M3='.scratch/barness-ai-minimax/evidence/ticket-10/international-formal-glm/international-formal-glm-8900034.json'
PINS={LEGACY+'index.json':'d0a17474721deef08bd10e004762b117f9b669288b9dbfac7d8041b725622d62',LEGACY+'inventory.json.gz':'7639de8797a2a6034446eda072b4cad528e75c7c6a9546d0be5137a2adfcd025',M3:'121075a2a34a35c9966c6bf2c81eff29f20b587df103c3eb4f071ebf703183b4'}

def inventories():
 with e.Root(c.ROOT) as reader:
  raw={n:reader.read(n,2<<20) for n in PINS}
 if any(u.sha(raw[n])!=sha for n,sha in PINS.items()):raise ValueError('certified original inventory metadata changed')
 with gzip.GzipFile(fileobj=io.BytesIO(raw[LEGACY+'inventory.json.gz'])) as stream:decoded=stream.read((8<<20)+1)
 if len(decoded)>8<<20 or u.sha(decoded)!='bb1a7c604a884199a926879d0196542e04533c1ebd0b39f1df2b2a360cd1f7c3':raise ValueError('original inventory decoded pin/limit')
 index=e.document(raw[LEGACY+'index.json']);legacy=e.document(decoded);m3=e.document(raw[M3]);result={}
 for name,row in index['routes'].items():
  prefix=row['root']+'/'
  result[name]={n.removeprefix(prefix):v for n,v in legacy.items() if n.startswith(prefix)}
 for route in ('chat','responses','messages'):
  marker='/.evidence/minimax-global-formal-glm/mm-global-formal-m3-'+route+'/'
  result['minimax-'+route]={}
  for row in m3['files'].values():
   if marker not in row.get('original',''):continue
   _,_,tail=row['original'].partition(marker);_,_,name=tail.partition('/');e.parts(name)
   result['minimax-'+route][name]={k:row[k] for k in ('bytes','sha256')}
  if not result['minimax-'+route]:raise ValueError('original M3 inventory missing')
 return result,m3

def require(name,path,current,certified):
 prefix=path+'/';actual={n.removeprefix(prefix):v for n,v in current.items()}
 if actual!=certified[name]:raise ValueError('original native bytes/inventory changed; cannot rehash as original')

OWNER='bf574a3ed52ff02819aa91b7a137d9abac7c4ba7333701e584cde352ac83aa72'
RECEIPTS='.scratch/barness-ai-minimax/evidence/ticket-14/gap2/actual-six-check-observed.json'

def original_receipts():
 with e.Root(c.ROOT) as reader:raw=reader.read(RECEIPTS,1<<20)
 if u.sha(raw)!='a9d8cef476da498fd9af6082a621eb15537d7d8edfcec9926ef47949396b9672':raise ValueError('certified original39 receipt anchor changed')
 table=e.document(raw);result={n.removesuffix('.json'):h for n,h in table['original18CanonicalPins'].items()}
 result.update({n:r['receiptSHA256'] for n,r in table['original21PriorPins'].items()})
 if len(result)!=39 or table['original39Unchanged'] is not True:raise ValueError('original39 receipt inventory required')
 return result
