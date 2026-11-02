#!/usr/bin/env python3
"""Retain actual international official responses and their provenance."""
import datetime
import hashlib
import json
import pathlib
import urllib.request

DEST = pathlib.Path(__file__).parent/'inputs'/'int'
DEST.mkdir(exist_ok=False)
PAGES = {
    'pricing':'https://platform.kimi.ai/docs/pricing/chat',
    'quickstart':'https://platform.kimi.ai/docs/guide/kimi-k3-1-pro-quickstart.md',
    'chat':'https://platform.kimi.ai/docs/api/chat.md',
    'responses':'https://platform.kimi.ai/docs/api/responses.md',
    'messages':'https://platform.kimi.ai/docs/api/messages.md',
    'cache':'https://platform.kimi.ai/docs/guide/context-caching.md',
}
for key,url in PAGES.items():
    meta={'url':url,'startedAt':datetime.datetime.now(datetime.timezone.utc).isoformat()}
    try:
        req=urllib.request.Request(url,headers={'User-Agent':'barness-catalog-maintenance/1.0'})
        with urllib.request.urlopen(req,timeout=25) as response:
            raw=response.read(16*1024*1024+1)
            if len(raw)>16*1024*1024:
                raise ValueError('response exceeds 16 MiB')
            meta.update(status=response.status,finalUrl=response.url,headers=dict(response.headers.items()))
        p=DEST/(key+'.body')
        with p.open('xb') as f:
            f.write(raw)
        meta.update(bytes=len(raw),sha256=hashlib.sha256(raw).hexdigest(),file=p.name)
    except Exception as error:
        meta.update(status='FETCH_FAILED',error=str(error))
    meta['fetchedAt']=datetime.datetime.now(datetime.timezone.utc).isoformat()
    with (DEST/(key+'.json')).open('x') as f:
        json.dump(meta,f,indent=2)
        f.write('\n')
    print(json.dumps({'page':key,**{k:meta.get(k) for k in ['status','bytes','sha256']}}),flush=True)
    if meta['status'] != 200:
        raise SystemExit(1)
