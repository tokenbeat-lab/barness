"""Public CLI FIFO swap regressions at selected-key, audit and settlement boundaries."""
import subprocess
import sys
from test_recovery5_round import Recovery5Round
from test_recovery5_fixture import approve,keys

INJECT=r'''
import pathlib,os,runpy,sys
cli=sys.argv[1];mode=sys.argv[2];sys.path.insert(0,str(pathlib.Path(cli).parent))
import final_process
victim=pathlib.Path('.evidence/provider-sites/recovery5/formal-zai-responses/controlled-session/regional-report.json').resolve()
original=pathlib.Path.is_file;fired=False;armed=mode=='audit'
def swapped(self,*a,**kw):
 global fired
 result=original(self,*a,**kw)
 if armed and not fired and self==victim and result:
  fired=True;self.unlink();os.mkfifo(self);print('FIFO_SWAP_'+mode,flush=True)
 return result
pathlib.Path.is_file=swapped
original_lstat=pathlib.Path.lstat
def swapped_lstat(self,*a,**kw):
 global fired
 info=original_lstat(self,*a,**kw)
 if armed and not fired and self==victim and __import__('stat').S_ISREG(info.st_mode):
  fired=True;self.unlink();os.mkfifo(self);print('FIFO_SWAP_'+mode,flush=True)
 return info
pathlib.Path.lstat=swapped_lstat
run=final_process.run
def selected(*a,**kw):
 global armed
 result=run(*a,**kw);armed=True;return result
final_process.run=selected
if mode=='credential':
 p=pathlib.Path('.env');p.unlink();os.mkfifo(p);print('FIFO_CREDENTIAL',flush=True)
sys.argv=[cli,'run-selected','.evidence/provider-sites/recovery5','formal-zai-responses'];runpy.run_path(cli,run_name='__main__')
'''

class Recovery5IO(Recovery5Round):
    def boundary(self,mode):
        base=self.ready();approve(self,base);keys(self)
        argv=[sys.executable,'-B','-c',INJECT,str(self.tools/'recovery5_round.py'),mode]
        try:p=subprocess.run(argv,cwd=self.root,env=self.env,capture_output=True,text=True,timeout=20)
        except subprocess.TimeoutExpired:
            self.journal.append(dict(argv=argv,state='TIMEOUT',exitCode=None));self.fail('public selected process blocked on '+mode+' FIFO')
        self.journal.append(dict(argv=argv,exitCode=p.returncode,stdout=p.stdout,stderr=p.stderr));self.assertEqual(p.returncode,1,p.stdout+p.stderr);self.assertIn('REFUSED',p.stderr)
        self.assertIn('FIFO_CREDENTIAL' if mode=='credential' else 'FIFO_SWAP_'+mode,p.stdout)
        self.assertFalse((self.root/'.evidence/minimax-renewals/provider-sites-final-acceptance-5/grant.json').exists())
        self.runcli('run-selected','.evidence/provider-sites/recovery5','mm-global-formal-m3-chat',ok=False)

    def test_io_settlement_open_swap_is_bounded(self):self.boundary('settle')
    def test_io_artifact_audit_open_swap_is_bounded(self):self.boundary('audit')
    def test_io_selected_credential_fifo_is_bounded(self):self.boundary('credential')

def load_tests(loader,tests,pattern):
    import unittest
    return unittest.TestSuite(Recovery5IO(n) for n in loader.getTestCaseNames(Recovery5IO) if n.startswith('test_io_'))
