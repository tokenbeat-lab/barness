"""Add finite duration scopes; retain all prior decisions and compare every field."""
import json
from pathlib import Path
root=Path('ai/e2e/testdata')
p=root/'pidiff/ledger.json';d=json.loads(p.read_text())
d['decisions']=[x for x in d['decisions'] if x.get('ref') not in ['docs/adr/0025-barness-ai-thinking-level-sampling.md#durationms','.scratch/barness-ai-thinking-sampling/release-verification/catalog-changes.json']]
manifest=Path('.evidence/barness-ai-thinking-sampling/migration/20261011T024344.369361000Z/manifest.json')
by={key:set() for key in ['openai-responses','openai-completions','anthropic-messages','google-generative-ai']}
for c in json.loads(manifest.read_text())['cases']:
 pcase=manifest.parent/c['dir']/'pidiff.json'
 if not pcase.exists():continue
 r=json.loads(pcase.read_text())
 if r['protocol'] in by:by[r['protocol']].add(c['id'])
for route,proto,prefix in [('responses','openai-responses','P01'),('chat','openai-completions','P04'),('deepseek-chat','openai-completions','P06'),('anthropic','anthropic-messages','P02'),('gemini','google-generative-ai','P03')]:
 for s in json.loads((root/route/'thinking-sampling.json').read_text())['scenarios']:
  by[proto].add('PIDIFF-'+prefix+'-'+s['id']+'-'+s.get('entry','stream'))
for proto,cases in by.items():
 for path in ['result.durationMs','events[*].message.durationMs','events[*].error.durationMs','events[*].partial.durationMs']:
  d['decisions'].append({'protocol':proto,'path':path,'kind':'changed','cases':sorted(cases),'classification':'extension','decision':'Maintainer approved 2026-10-11: both independent runs expose validated nonnegative integer terminal/result durationMs; elapsed numerical values differ. No field is dropped from comparison.','ref':'docs/adr/0025-barness-ai-thinking-level-sampling.md#durationms'})
 for kind in ['only_pi','only_barness']:
  d['decisions'].append({'protocol':proto,'path':'events[*].partial.durationMs','kind':kind,'cases':sorted(cases),'classification':'extension','decision':'Existing live PartialView contract: the shared view may advance to its settled duration before a queued nonterminal event is read. Finite timing-field scope only; final/terminal presence and equality are asserted before classification.','ref':'docs/adr/0025-barness-ai-thinking-level-sampling.md#durationms'})
for case in ['simple-context-clamp','simple-context-clamp-usage','simple-context-skip-failed']:
 d['decisions'].append({'protocol':'openai-responses','path':'request.body.max_output_tokens','kind':'changed','cases':['PIDIFF-P01-E04-options-'+case+'-streamSimple'],'classification':'fixed','decision':'pi 1.1.0 context estimator uses ceil(chars/3.5); public fixture expectations and estimator updated from verified source. Any recurring difference is a regression.','ref':'.scratch/barness-ai-thinking-sampling/release-verification/catalog-changes.json'})
# Keep each decision atomic on one line; existing long case arrays stay reviewable
# as a single finite scope rather than thousands of repeated path entries.
p.write_text('{\n  "decisions": [\n'+',\n'.join('    '+json.dumps(x,ensure_ascii=False,separators=(',',':')) for x in d['decisions'])+'\n  ],\n  "routes": '+json.dumps(d['routes'],ensure_ascii=False,indent=2)+'\n}\n')
print({k:len(v) for k,v in by.items()})
