"""Only the new fixed round's same-source3PASS canonical bytes may publish data."""
import io,stat,zipfile
from pathlib import PurePosixPath
import paid_core as c
import routed_evidence as e
import repair_contract as r
import renewal_common as u
FIELDS={'formalReportSha256':'formal-glm-report.json','integritySha256':'glm-integrity.json','supportReportSha256':'live-report.json','executionSourceSha256':'glm-execution-source.json'}

def require(route):
 job=r.jobs(c.ROOT/'.evidence/unused-publication','glm')[r.ROUND+'-glm-'+route];actual=e.settled(job,contract=r)
 if any(actual.get(k)!=v for k,v in dict(state='REPORTED',status='PASS',calls=22,attempts=22).items()):raise RuntimeError('new exact reviewed/approved own full22 PASS canonical capture required; no old round fallback')
 with e.Root(c.common_root()) as root:raw=root.read('.evidence/minimax-launches/'+job['name']+'.json')
 return e.document(raw),u.sha(raw)

def zipped(entry,wanted):
 raw=u.file_bytes(u.repo_path(entry['archive']),32<<20)
 if u.sha(raw)!=entry['archiveSHA256']:raise RuntimeError('new published archive pin changed')
 with zipfile.ZipFile(io.BytesIO(raw)) as archive:
  infos=archive.infolist()
  if len(infos)>512 or len({v.filename for v in infos})!=len(infos) or sum(v.file_size for v in infos)>128<<20:raise RuntimeError('fixed published archive inventory bound')
  for value in infos:
   e.parts(value.filename)
   if value.file_size>16<<20 or stat.S_IFMT(value.external_attr>>16)!=stat.S_IFREG:raise RuntimeError('bounded regular published members required')
  selected={}
  for name in wanted:
   e.parts(name)
   with archive.open(name) as stream:data=stream.read((16<<20)+1)
   if len(data)>16<<20 or len(data)!=archive.getinfo(name).file_size:raise RuntimeError('published member byte limit')
   selected[name]=data
 return selected,{v.filename for v in infos}

def validate(index,views=None):
 from renewal_glm import known_usage,USAGE_BASIS,BASE
 if index.get('collectionRound')!=r.ROUND or set(index.get('routes',{}))!=set(u.ROUTES) or set(index.get('launcherReceiptSHA256',{}))!=set(u.ROUTES):raise RuntimeError('exact new three-route publication index required')
 expected_usage={};commits=set();captures={};originals={}
 for route in u.ROUTES:
  receipt,receipt_sha=require(route)
  if index['launcherReceiptSHA256'][route]!=receipt_sha:raise RuntimeError('new canonical receipt changed after publication')
  relative=receipt['collectionRoot'].removeprefix('$REPO/');e.parts(relative);base=u.bundle(relative+'/glm-'+route,'formal-glm-report.json');files=u.tree(base);report=e.document(files['formal-glm-report.json']);pins=receipt['capturePins'];commits.add(receipt['sourceCommit'])
  if pins!={'schema':1,**{key:u.sha(files[leaf]) for key,leaf in FIELDS.items()}}:raise RuntimeError('new original canonical capture pins changed')
  row=index['routes'][route];prefix=base.name+'/';required=[prefix+leaf for leaf in files]
  if row.get('reportPath')!=prefix+'formal-glm-report.json' or row.get('reportSHA256')!=pins['formalReportSha256'] or index.get('sourceManifestSHA256')!=report.get('sourceManifestHash'):raise RuntimeError('new raw selector source/report identity changed')
  selected,members=zipped(row,required)
  if members!=set(required) or any(selected[prefix+leaf]!=raw for leaf,raw in files.items()):raise RuntimeError('new raw archive must preserve every original canonical capture byte')
  expected_usage[route]={k:row[k] for k in ('archiveSHA256','reportSHA256')}|known_usage(files,report);captures[route]=pins;originals[route]=files
 if len(commits)!=1 or index.get('originalLiveSourceCommit')!=next(iter(commits)):raise RuntimeError('new own GLM routes must share actual source')
 if u.file_bytes(c.ROOT/BASE/'live-captures.sha256')!=(c.digest(c.ROOT/BASE/'live-captures.json')+'\n').encode():raise RuntimeError('raw selector SHA companion changed')
 expected=dict(schema=1,sourceCommit=next(iter(commits)),basis=USAGE_BASIS,routes=expected_usage)
 if u.json_doc(c.ROOT/BASE/'known-usage.json')!=expected:raise RuntimeError('independent known usage changed after publication')
 if views is None:views=u.json_doc(c.ROOT/BASE/'support-views.json')
 if views.get('kind')!='no-paid-derived-glm-support-views' or set(views.get('routes',{}))!=set(u.ROUTES):raise RuntimeError('own new three-route support views required')
 index_sha=c.digest(c.ROOT/BASE/'live-captures.json')
 for route in u.ROUTES:
  row=views['routes'][route];files=originals[route];required=['original/'+leaf for leaf in files]+['glm-support-derivation.json'];selected,members=zipped(row,required);pins=captures[route];derived=e.document(selected['glm-support-derivation.json'])
  if row.get('reportPath')!='live-report.json' or row.get('files')!=len(members) or {n for n in members if n.startswith('original/')}!={'original/'+n for n in files} or any(selected['original/'+leaf]!=raw for leaf,raw in files.items()):raise RuntimeError('new view must embed original canonical capture')
  expected=dict(schema=1,kind='glm-derived-live-support',route=route,rawCaptureIndexSHA256=index_sha,rawArchiveSHA256=index['routes'][route]['archiveSHA256'],originalFormalReportSHA256=pins['formalReportSha256'],originalSupportReportSHA256=pins['supportReportSha256'],originalGitCommit=next(iter(commits)))
  if any(derived.get(k)!=v for k,v in expected.items()):raise RuntimeError('new view derivation provenance changed')
 return dict(round=r.ROUND,sourceCommit=next(iter(commits)),threeOwnPASS=True,credentialsRead=False,paidCalls=0)
