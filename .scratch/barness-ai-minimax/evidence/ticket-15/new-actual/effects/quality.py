"""Independent finite recomputation of the fixed twelve synthetic visual samples.

This historical frontdoor shares no mutable evaluation state with the producer.
Only observed native gold/result bodies are compared; these are not population
or billing estimates. Full original metrics remain in the immutable report.
"""
import collections,math

def equal(a,b):
 if isinstance(a,(int,float)) and not isinstance(a,bool) and isinstance(b,(int,float)):return math.isclose(a,b,rel_tol=1e-12,abs_tol=1e-12)
 if isinstance(a,dict) and isinstance(b,dict):return a.keys()==b.keys() and all(equal(a[k],b[k]) for k in a)
 if isinstance(a,list) and isinstance(b,list):return len(a)==len(b) and all(equal(x,y) for x,y in zip(a,b))
 return a==b

def group(samples,gold,threshold):
 m=dict(total=len(samples),classScored=0,classCorrect=0,existScored=0,existCorrect=0,falsePositive=0,falseNegative=0,classAccuracy=None,existAccuracy=None,confusion={},statuses={},errors=[],confidenceBins=[dict(count=0,correct=0,sum=0) for _ in range(10)],existenceBrier=None);brier=0
 for s,g in zip(samples,gold):
  m['statuses'][s['status']]=m['statuses'].get(s['status'],0)+1;a=s['view']['kind'];pred=a.get('choice') or a.get('status') or s['status'];row=m['confusion'].setdefault(g['label'],{});row[pred]=row.get(pred,0)+1;correct=a['status']=='answered' and a.get('choice')==g['label'];m['classScored']+=a['status']=='answered';m['classCorrect']+=correct
  confidence=a.get('confidence')
  if confidence is not None:
   cell=m['confidenceBins'][min(9,int(confidence*10))];cell['count']+=1;cell['sum']+=confidence;cell['correct']+=correct
  e=s['view']['exists'].get('probability');ec=False
  if e is not None:
   prediction=e>=threshold;ec=prediction==g['exists'];m['existScored']+=1;m['existCorrect']+=ec;m['falsePositive']+=prediction and not g['exists'];m['falseNegative']+=not prediction and g['exists'];brier+=(e-int(g['exists']))**2
  if not correct or not ec:m['errors'].append(s['id'])
 m['classAccuracy']=m['classCorrect']/len(samples);m['existAccuracy']=m['existCorrect']/len(samples)
 if m['existScored']:m['existenceBrier']=brier/m['existScored']
 return m

def iou(a,b):
 overlap=max(0,min(a['bottom'],b['bottom'])-max(a['top'],b['top']))*max(0,min(a['right'],b['right'])-max(a['left'],b['left']));area=lambda x:(x['bottom']-x['top'])*(x['right']-x['left']);return overlap/(area(a)+area(b)-overlap)

def metrics(report,gold,experiment):
 samples=report['samples'];threshold=experiment['existenceThreshold'];route=report['route'];local='local' if route=='jev' else 'supplementalContext';out=dict(common=group(samples[:8],gold[:8],threshold),local=None,supplementalContext=None,box=None);out[local]=group(samples[8:],gold[8:],threshold)
 if route=='jev':
  box=dict(present=0,absent=0,iouCount=0,meanIoU=None,presentDenominatorMeanIoU=None,consumed=0,absentFalsePositive=0,absentRawBoxes=0,lowExistenceBoxes=0,missingExistenceBoxes=0,missingPresentBoxes=0);total=0
  for s,g in zip(samples[8:],gold[8:]):
   box['present' if g['exists'] else 'absent']+=1;a=s['view'].get('location',{}).get('box')
   if a is None:
    box['missingPresentBoxes']+=g['exists'];continue
   if g['exists']:total+=iou(a,g['box']);box['iouCount']+=1
   else:box['absentRawBoxes']+=1
   e=s['view']['exists'].get('probability')
   if e is None:box['missingExistenceBoxes']+=1;continue
   if e<threshold:box['lowExistenceBoxes']+=1;continue
   box['consumed' if g['exists'] else 'absentFalsePositive']+=1
  if box['iouCount']:box['meanIoU']=total/box['iouCount']
  box['presentDenominatorMeanIoU']=total/box['present'];out['box']=box
 if not equal(out,report['metrics']):raise ValueError('independent visual quality metrics differ')
 return out
