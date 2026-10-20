#!/usr/bin/env python3
"""Read existing implementation evidence only; never execute captured commands."""
import concurrent.futures, collections, hashlib, json, pathlib, re, tarfile, posixpath, datetime
ROOT=pathlib.Path(__file__).resolve().parents[3]
BASE=ROOT/'.scratch/barness-ai-minimax/evidence'
OUT=ROOT/'.scratch/barness-ai-minimax/session-analysis'
START='2026-10-17T22:02:53'; END='2026-10-18T06:44:34'

def safe(v,limit=6000):
 s=v if isinstance(v,str) else json.dumps(v,ensure_ascii=False)
 s=re.sub(r'(?i)(Bearer\s+)[^\s"<>]+',r'\1[redacted]',s)
 return s if len(s)<=limit else s[:limit]+' [内容截断；完整原文见 evidence]'

def scan(p):
 manifests=[]; assertions={}; reports=[]; audits=[]
 with tarfile.open(p,'r|gz') as t:
  for m in t:
   if not m.isfile() or m.size>32*1024*1024: continue
   name=m.name
   wanted=(name.endswith('/manifest.json') or name=='manifest.json' or name.endswith('/assertions.json') or name.endswith(('/live-report.json','/minimax-report.json','/formal-glm-report.json','/evaluation-report.json','/multimodal-report.json','/release-report.json','/audit.json')) or re.search(r'captures/visual-[^/]+/report.json$',name))
   if not wanted:continue
   try:d=json.loads(t.extractfile(m).read())
   except (ValueError,UnicodeError):continue
   if not isinstance(d,dict):continue
   if name.endswith('manifest.json') and isinstance(d.get('cases'),list):
    at=d.get('started_at','')
    if at and START<=at[:19]<=END:manifests.append((name,d))
   elif name.endswith('/assertions.json'):
    if d.get('status')=='FAIL':assertions[name]=d
   elif name.endswith('/audit.json'):
    if d.get('findings'):audits.append((name,d))
   else:reports.append((name,d))
 rows=[]
 case_paths={}
 for mn,md in manifests:
  for mc in md['cases']:
   case_paths[posixpath.join(posixpath.dirname(mn),mc.get('dir',mc.get('id','')))]=mc
 for name,d in manifests:
  mh=hashlib.sha256(json.dumps(d,sort_keys=True,separators=(',',':')).encode()).hexdigest()
  for c in d['cases']:
   if c.get('status')!='FAIL':continue
   an=posixpath.join(posixpath.dirname(name),c.get('dir',c.get('id','')),'assertions.json')
   a=assertions.get(an,{})
   false=[{'name':x.get('name',''), 'detail':safe(x.get('detail',''))} for x in a.get('assertions',[]) or [] if x.get('pass') is False]
   ancestors=[]
   parent=posixpath.dirname(posixpath.dirname(an))
   while parent and parent!='.':
    if parent in case_paths:
     pc=case_paths[parent];ancestors.append({'case':pc.get('id'),'test':pc.get('test'),'status':pc.get('status')})
    parent=posixpath.dirname(parent)
   rows.append({'ancestor_cases':ancestors,'manifest_identity':mh,'started_at':d.get('started_at'),'finished_at':d.get('finished_at'),'source':d.get('git_commit'),'case':c.get('id'),'test':c.get('test'),'false_assertions':false,'no_assertions_retained':not bool(a.get('assertions')),'evidence':[str(p.relative_to(ROOT))+'::'+an]})
 return {'archive':str(p.relative_to(ROOT)),'manifest_count':len(manifests),'failures':rows,'reports':reports,'audits':audits}

paths=sorted(p for p in BASE.rglob('*.tar.gz') if not p.name.startswith(('A-offline-','A-race-')))
for n in ['whole-ordinary-failed-9481a6a','preactivation-activated-whole-de05193','glm-corrective-q-9dc-audit-failure','corrective-approved-prepay-5bf2c16']:
 p=ROOT/'.evidence/minimax-retained'/(n+'.tar.gz')
 if p.exists():paths.append(p)
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as ex:
 for d in ex.map(scan,paths):results.append(d)
merged={}
for result in results:
 for d in result['failures']:
  key=(d['manifest_identity'],d['case'],d['test'])
  if key in merged:merged[key]['evidence']+=d['evidence']
  else:merged[key]=d
rows=sorted(merged.values(),key=lambda d:(d['started_at'] or '',d['case'] or ''))
(OUT/'raw-failed-cases.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
# Keep the exact gate report and provider reports; no raw requests, headers, secrets or thinking content.
selected=[]
for r in results:
 for n,d in r['reports']:
  if r['archive'].endswith(('legacy-paid-3d3.tar.gz','glm-corrective-actual-5bf2c16.tar.gz','international-formal-glm-8900034.tar.gz','international-research-38b7098.tar.gz','actual-auth-failure-9481a6a.tar.gz','release-consumer-ad20469.tar.gz')):
   if n.endswith('release-report.json'):selected.append({'archive':r['archive'],'member':n,'report':d})
(OUT/'scan-index.json').write_text(json.dumps({'archives':[{'archive':r['archive'],'manifests':r['manifest_count'],'fail_cases':len(r['failures']),'audits_with_findings':[{'member':n,'findings':d.get('findings')} for n,d in r['audits']]} for r in results],'unique_failure_occurrences':len(rows)},ensure_ascii=False,indent=2)+'\n')
(OUT/'gate-reports.json').write_text(json.dumps(selected,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'archives':len(results),'unique_failure_occurrences':len(rows),'cases':len(set(x['case'] for x in rows)),'by_archive':dict(collections.Counter(x['evidence'][0].split('::')[0] for x in rows))},ensure_ascii=False,indent=2))
