"""Synthetic public Client fixtures. Expected profile keys are literal worked
examples from acceptance.md, never obtained by importing the implementation.
Event scripts/result semantics reuse existing protocol acceptance fixtures.
"""
from pathlib import Path
import json,copy
root=Path('ai/e2e/testdata')
levels=['off','minimal','low','medium','high','xhigh','max']
for route in ['responses','chat','deepseek-chat','deepseek-responses']:
 if route=='responses':
  d=json.loads((root/route/'text-basic.json').read_text())
  base={'model':d['model'],'context':{'systemPrompt':d['systemPrompt'],'messages':[{'role':'user','content':[{'type':'text','text':d['user']}]}]},'events':d['events'],'expect':{'request':d['expect']['request'],'stopReason':'stop','rawStopReason':'completed','responseId':'resp_001','responseModel':'','content':[{'type':'text','text':'Hello, world!','textSignature':d['expect']['textSignature']}],'usage':d['expect']['usage']}}
 else:
  base=next(s for s in json.loads((root/route/'text.json').read_text())['scenarios'] if s['id']=='text')
 scenarios=[]
 for entry in ['stream','streamSimple']:
  sc=copy.deepcopy(base);sc['id']='thinking-sampling-S01-merge-'+entry;sc['entry']=entry
  sc['about']='S01/R04: synthetic values; named -> general -> high profile -> call, shallow replacement retaining zero/false/null.'
  general={'temperature':1,'top_p':0.95,'extra':{'a':1,'b':2},'seed':1,'array':[1,2]}
  profile={'temperature':0.8,'top_k':20,'extra':{'b':3},'seed':False,'array':[3]}
  call={'temperature':0,'top_p':None,'extra':{'z':0},'array':[False,None]}
  sc['modelPatch']={'reasoning':True,'thinkingLevelMap':{},'samplingParams':general,'samplingParamsByThinkingLevel':{'high':profile}}
  sc['options']={'maxTokens':16,'temperature':0.2,'reasoning' if entry=='streamSimple' else 'reasoningEffort':'high','samplingParams':call}
  sc['expect']['request'].pop('method',None)
  body=sc['expect']['request']['body'];body['temperature']=0;body.update({'top_p':None,'top_k':20,'extra':{'z':0},'seed':False,'array':[False,None]})
  if route in ['responses','deepseek-responses']:
   body['input'][0]['role']='developer';body['max_output_tokens']=16;body['reasoning']={'effort':'high','summary':'auto'}
   if route=='responses':body['include']=['reasoning.encrypted_content']
  else:
   body['max_tokens' if route=='deepseek-chat' else 'max_completion_tokens']=16;body['reasoning_effort']='high'
   if route=='chat':body['messages'][0]['role']='developer'
   if route=='deepseek-chat':body['thinking']={'type':'enabled'}
  if route=='deepseek-responses':sc['pidiffSkip']='ADR-0014: DeepSeek Responses is an extension route, no native pi route.'
  scenarios.append(sc)

 # These expected levels are literal acceptance examples, not a clamp algorithm.
 cases=[]
 for level in ['minimal','low','medium','high']:
  cases.append(('S02-supported-'+level,level,{},level,True))
 for requested,disabled,expected in [
  ('minimal',{'minimal':None},'low'),
  ('low',{'low':None,'medium':None},'high'),
  ('medium',{'medium':None},'high'),
  ('high',{'high':None},'medium'),
  ('medium',{'medium':None,'high':None},'low'),
 ]:
  cases.append(('S03-clamp-'+requested+'-'+expected,requested,disabled,expected,True))
 cases+= [
  ('S04-canonical-minimal','minimal',{'minimal':'low','low':'low'},'minimal',True),
  ('S04-canonical-low','low',{'minimal':'low','low':'low'},'low',True),
  ('S05-xhigh-undeclared','xhigh',{},'high',True),
  ('S05-max-undeclared','max',{},'high',True),
  ('S05-xhigh-declared','xhigh',{'xhigh':'ultra'},'xhigh',True),
  ('S05-max-declared','max',{'max':'maximum'},'max',True),
  ('S05-extended-null','max',{'xhigh':None,'max':None},'high',True),
  ('S06-off-none','',{'off':'none'},'off',True),
  ('S06-off-null-minimal','',{'off':None},'minimal',True),
  ('S06-off-null-low','',{'off':None,'minimal':None},'low',True),
  ('S06-all-null','low',{level:None for level in levels},'off',True),
  ('S07-omitted','',{},'off',True),
  ('S08-non-reasoning','high',{},'off',False),
  ('S12-full-simple-wire','low',{'low':None,'medium':None,'high':'strong'},'high',True),
 ]
 profiles={level:{'top_k':i+1} for i,level in enumerate(levels)}
 # Each entry's wire value below is worked independently from its literal case.
 for name,requested,mapping,selected,reasoning in cases:
  for entry in ['stream','streamSimple']:
   sc=copy.deepcopy(base);sc['entry']=entry;sc['id']='thinking-sampling-'+name+'-'+entry
   sc['about']=name+'; expected canonical profile '+selected+'; synthetic marker top_k.'
   sc['modelPatch']={'reasoning':reasoning,'thinkingLevelMap':mapping,'samplingParams':{'temperature':0.9},'samplingParamsByThinkingLevel':profiles}
   sc['options']={'maxTokens':16,'temperature':0.2}
   if requested:sc['options']['reasoning' if entry=='streamSimple' else 'reasoningEffort']=requested
   body=sc['expect']['request']['body'];sc['expect']['request'].pop('method',None)
   body['temperature']=0.9;body['top_k']=levels.index(selected)+1
   if route in ['responses','deepseek-responses']:
    body['input'][0]['role']='developer' if reasoning else 'system';body['max_output_tokens']=16
    body.pop('reasoning',None);body.pop('include',None)
    wire=requested if entry=='stream' else (selected if requested and selected!='off' else '')
    if reasoning and wire:
     body['reasoning']={'effort':mapping.get(wire) or wire,'summary':'auto'}
     if route=='responses':body['include']=['reasoning.encrypted_content']
    elif reasoning and mapping.get('off','none') is not None:body['reasoning']={'effort':mapping.get('off','none')}
   else:
    body['messages'][0]['role']='developer' if route=='chat' and reasoning else 'system'
    body['max_tokens' if route=='deepseek-chat' else 'max_completion_tokens']=16
    body.pop('thinking',None);body.pop('reasoning_effort',None)
    wire=requested if entry=='stream' else (selected if requested and selected!='off' else '')
    if reasoning and wire:
     body['reasoning_effort']=mapping.get(wire) or wire
     if route=='deepseek-chat':body['thinking']={'type':'enabled'}
    elif reasoning:
     if route=='deepseek-chat' and mapping.get('off','unset') is not None:body['thinking']={'type':'disabled'}
     elif route=='chat' and mapping.get('off') is not None:body['reasoning_effort']=mapping['off']
   if route=='deepseek-responses':sc['pidiffSkip']='ADR-0014 extension route; no native pi differential.'
   scenarios.append(sc)
 # Sparse/empty profiles never remove general defaults, and a nested value is
 # replaced in full. These are explicit expected payloads, not recursive merge.
 for variant,general,profile,call,expected in [
  ('S09-missing',{'top_p':0.5},{'low':{'top_k':2}},{'temperature':0.4},{'top_p':0.5,'temperature':0.4}),
  ('S09-empty',{'top_p':0.5},{'high':{}},{'temperature':0.4},{'top_p':0.5,'temperature':0.4}),
  ('S09-no-profiles',{'top_p':0.5},None,{}, {'top_p':0.5,'temperature':0.2}),
  ('S09-no-maps',None,None,None,{'temperature':0.2}),
  ('S10-shallow',{'extra':{'a':1,'b':2},'array':[1,2]},{'high':{'extra':{'b':3},'array':[None,False]}},{},{'temperature':0.2,'extra':{'b':3},'array':[None,False]}),
  ('S11-null',{'temperature':0.9,'seed':1},{'high':{'temperature':None,'seed':False}},{},{'temperature':None,'seed':False}),
 ]:
  for template in scenarios[:2]:
   sc=copy.deepcopy(template);sc['id']='thinking-sampling-'+variant+'-'+sc['entry'];sc['about']=variant
   sc['modelPatch']['samplingParams']=general;sc['modelPatch']['samplingParamsByThinkingLevel']=profile;sc['options']['samplingParams']=call
   body=sc['expect']['request']['body']
   for key in ['temperature','top_p','top_k','extra','seed','array']:body.pop(key,None)
   body.update(expected);sc['options']['temperature']=0.2
   scenarios.append(sc)
 if route in ['responses','deepseek-responses']:
  for selected,mapping in [('medium',{}),('high',{'medium':None})]:
   sc=copy.deepcopy(scenarios[0]);sc['id']='thinking-sampling-S13-summary-'+selected;sc['about']='Summary-only wire medium, sampler '+selected
   sc['options']={'maxTokens':16,'reasoningSummary':'concise'}
   sc['modelPatch']={'reasoning':True,'thinkingLevelMap':mapping,'samplingParamsByThinkingLevel':profiles}
   body=sc['expect']['request']['body']
   for key in ['temperature','top_p','top_k','extra','seed','array']:body.pop(key,None)
   body['top_k']=levels.index(selected)+1;body['reasoning']={'effort':'medium','summary':'concise'}
   scenarios.append(sc)
 # A call-level reasoning override is applied AFTER profile selection.
 for template in scenarios[:2]:
  sc=copy.deepcopy(template);sc['id']='thinking-sampling-S16-reasoning-override-'+sc['entry']
  override={'reasoning':{'effort':'low'}} if route in ['responses','deepseek-responses'] else {'reasoning_effort':'low'}
  if route=='deepseek-chat':override['thinking']={'type':'disabled'}
  sc['options']['samplingParams'].update(override);sc['expect']['request']['body'].update(override)
  scenarios.append(sc)
 out={'case':'Canonical thinking-level sampling','source':'Synthetic worked examples from .scratch/barness-ai-thinking-sampling/acceptance.md; values are not vendor recommendations. Frozen baseline pi-ai 1.1.0 abe508e1b89912adde45528136c3221eb69acdd7.','scenarios':scenarios}
 (root/route/'thinking-sampling.json').write_text('{\n"case":'+json.dumps(out['case'])+',\n"source":'+json.dumps(out['source'])+',\n"scenarios":[\n'+',\n'.join(json.dumps(s,ensure_ascii=False,separators=(',',':')) for s in scenarios)+'\n]}\n')
