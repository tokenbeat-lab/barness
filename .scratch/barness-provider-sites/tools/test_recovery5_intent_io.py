"""Public canonical settlement FIFO cases; no selected credential is supplied."""
import json
import os
import signal
import stat
import subprocess
import sys
from test_recovery5_round import Recovery5Round
from test_recovery5_fixture import approve

INJECT=r'''
import os,pathlib,runpy,sys
cli,mode=sys.argv[1:];sys.path.insert(0,str(pathlib.Path(cli).parent))
import final_process
victim=pathlib.Path('.evidence/minimax-launches/provider-sites-final-acceptance-5-formal-zai-responses.json').resolve()
armed=False;fired=False;selected=final_process.selected
def replace():
 global fired
 fired=True;victim.unlink();os.mkfifo(victim);print('FIFO_CANONICAL_INTENT_'+mode,flush=True)
def refused(*args):
 global armed
 try:return selected(*args)
 finally:
  armed=True
  if mode=='static':replace()
final_process.selected=refused
opened=pathlib.Path.open
def path_open(self,*args,**kw):
 if armed and not fired and self==victim and (args[0] if args else kw.get('mode','r'))=='r':replace()
 return opened(self,*args,**kw)
pathlib.Path.open=path_open
native_open=os.open
def fd_open(path,flags,*args,**kw):
 if armed and not fired and pathlib.Path(path)==victim and flags&os.O_ACCMODE==os.O_RDONLY:replace()
 return native_open(path,flags,*args,**kw)
os.open=fd_open
sys.argv=[cli,'run-selected','.evidence/provider-sites/recovery5','formal-zai-responses'];runpy.run_path(cli,run_name='__main__')
'''

class Recovery5IntentIO(Recovery5Round):
    def boundary(self,mode):
        base=self.ready();approve(self,base)
        argv=[sys.executable,'-B','-c',INJECT,str(self.tools/'recovery5_round.py'),mode]
        process=subprocess.Popen(argv,cwd=self.root,env=self.env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
        timed_out=False
        try:stdout,stderr=process.communicate(timeout=12)
        except subprocess.TimeoutExpired:
            timed_out=True;os.killpg(process.pid,signal.SIGKILL);stdout,stderr=process.communicate()
        intent=self.root/'.evidence/minimax-launches/provider-sites-final-acceptance-5-formal-zai-responses.json'
        fifo=stat.S_ISFIFO(intent.lstat().st_mode)
        self.journal.append(dict(argv=argv,state='TIMEOUT' if timed_out else 'COMPLETED',exitCode=process.returncode,stdout=stdout,stderr=stderr,selectedCredentialSupplied=False,vendorChildStarted=(base/'selected-child-started').exists(),refusedNonregular=dict(path=str(intent.relative_to(self.root)),kind='FIFO' if fifo else 'unexpected',copied=False)))
        self.assertIn('FIFO_CANONICAL_INTENT_'+mode,stdout)
        self.assertTrue(fifo)
        self.assertTrue((self.root/'.env').is_dir())
        self.assertFalse((base/'selected-child-started').exists())
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-5/grant.json').exists())
        self.assertFalse(timed_out,'canonical intent blocked on FIFO')
        self.assertEqual(process.returncode,1,stdout+stderr);self.assertIn('REFUSED',stderr)
        command=json.loads((base/'formal-zai-responses-command.json').read_text())
        self.assertEqual(command['state'],'ERROR');self.assertEqual(command['exitCode'],1)
        self.runcli('run-selected','.evidence/provider-sites/recovery5','mm-global-formal-m3-chat',ok=False)
        self.assertTrue((self.root/'.env').is_dir())

    def test_intent_static_fifo_after_key_refusal_is_bounded(self):self.boundary('static')
    def test_intent_open_fifo_swap_after_key_refusal_is_bounded(self):self.boundary('opening')

def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(Recovery5IntentIO(n) for n in loader.getTestCaseNames(Recovery5IntentIO) if n.startswith('test_intent_'))
