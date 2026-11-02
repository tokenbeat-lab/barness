#!/usr/bin/env python3
import pathlib,json,hashlib,copy
root=pathlib.Path(__file__).resolve().parent
out=root/'definitions-v1';out.mkdir(exist_ok=False)
for region,provider in [('int','moonshot'),('cn','moonshot-cn')]:
 source=root/'inputs'/region
 def ref(name):
  p=source/(name+'.body');return str(p.relative_to(root))+'#sha256:'+hashlib.sha256(p.read_bytes()).hexdigest()
 quick=ref('quickstart');cache=ref('cache');pricing=ref('pricing')
 fx=ref('fx-2026-11-02') if region=='cn' else None
 rate={'input':2,'output':10,'cacheRead':0.2,'cacheWrite':2} if region=='int' else {'input':1.9360795878,'output':10.1271855360,'cacheRead':0.1936079588,'cacheWrite':1.9360795878}
 price={'provider':provider,'model':'kimi-k3.1-pro','currency':'USD' if region=='int' else 'CNY','unit':'per 1000000 tokens','native':{'input':2,'output':10,'cacheRead':0.2,'cacheWrite5m':2,'cacheWrite1h':4} if region=='int' else {'input':13,'output':68,'cacheRead':1.3,'cacheWrite5m':13,'cacheWrite1h':26},'catalogUSD':rate,'write1hUSD':4 if region=='int' else 3.8721591756,'pricingEvidence':pricing,'cacheEvidence':cache,'conversion':None if region=='int' else {'source':fx,'date':'2026-11-02','usdToCny':'6.7146','rounding':'10 decimal places, half-even; 1h is rounded input times two','note':'Frozen estimate, not USD vendor quotation; existing K3 rates unchanged'}}
 (out/(region+'-prices.json')).write_text(json.dumps(price,ensure_ascii=False,indent=2)+'\n')
 for api,n in [('openai-completions','chat'),('openai-responses','responses'),('anthropic-messages','messages')]:
  definition={'provider':provider,'api':api,'id':'kimi-k3.1-pro','name':f'Kimi K3.1 Pro {provider} {api}','reasoning':True,'input':['text'],'contextWindow':1048576,'maxTokens':131072,'thinkingLevelMap':{'off':None,'minimal':None,'low':'low','medium':None,'high':'high','xhigh':None,'max':'max'},'defaultThinkingLevel':'max','compat':{'requiresThinking':True},'cost':rate}
  prefix=region+'-'+n
  (out/(prefix+'-definition.json')).write_text(json.dumps(definition,ensure_ascii=False,indent=2)+'\n')
  apiRef=ref(n)
  decisions={}
  for key in ['provider','api','id','name']:
   decisions['/'+key]={'ownership':'manual','reason':'Independently identified official site, protocol and wire identity; this is not a cross-site inference.','evidence':apiRef}
  for key in ['reasoning','input','contextWindow','maxTokens','thinkingLevelMap','defaultThinkingLevel','compat/requiresThinking']:
   decisions['/'+key]={'ownership':'manual','reason':'Official model facts reviewed. Text and 131072 maximum are the existing implementation subset, not the vendor maximum; always-thinking low/high/max with max default.','evidence':quick}
  for key in rate:
   decisions['/cost/'+key]={'ownership':'manual','reason':'Standard per-million token rate; cacheWrite is 5m default, 1h tracked separately. '+('CNY converted at frozen 2026-11-02 CFETS 6.7146 with half-even ten places; see '+fx if fx else 'Official USD price; no cross-region substitution.'),'evidence':pricing}
  (out/(prefix+'-decisions.json')).write_text(json.dumps(decisions,ensure_ascii=False,indent=2)+'\n')
