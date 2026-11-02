#!/usr/bin/env python3
import pathlib,subprocess,json,sys
w=pathlib.Path('artifacts/catalog/work/20261102T023400Z-c9d70d4f')
for i,(reg,provider,api,n) in enumerate([('int','moonshot','openai-responses','responses'),('int','moonshot','anthropic-messages','messages'),('cn','moonshot-cn','openai-completions','chat'),('cn','moonshot-cn','openai-responses','responses'),('cn','moonshot-cn','anthropic-messages','messages')],5):
 meta=json.loads((w/'inputs'/reg/'pricing.json').read_text()); url=meta.get('requestedURL',meta.get('url')); stamp=meta.get('retrievedAtUTC',meta.get('fetchedAt'))
 args=['python3',str(w/'operate-v1.py'),'--label',f'{i:02d}-official-{reg}-{n}-preview','--',str(w/'bin/catalogctl-before-core'),'add','--store','artifacts/catalog/work/store','--operation','chat','--provider',provider,'--api',api,'--model','kimi-k3.1-pro','--source-kind','official','--source',str(w/'inputs'/reg/'pricing.body'),'--source-url',url,'--source-time',stamp,'--definition',str(w/'definitions-v1'/f'{reg}-{n}-definition.json'),'--decisions',str(w/'definitions-v1'/f'{reg}-{n}-decisions.json'),'--reason','Official K3.1 Pro text implementation subset with reviewed standard rates','--dry-run']
 p=subprocess.run(args); 
 if p.returncode: sys.exit(p.returncode)
