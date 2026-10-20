"""Private assembly for the two actual fixed recovery workflows, never a registry."""
from recovery_flow_files import Artifacts
from recovery_flow_state import State
from recovery_flow_execute import Execution
from recovery_flow_native import NativeConsumer
from recovery_flow_publication import Publication
from recovery_flow_finalize import Finalizer

class _FixedFlow:
    def __init__(self,contract,history,authority):
        self.r,self.h,self.authority=contract,history,authority
        self.files=Artifacts(contract.c)
        self.s=State(self);self.e=Execution(self);self.n=NativeConsumer(self)
        self.p=Publication(self);self.f=Finalizer(self)

def fixed3():
    import recovery_contract,recovery_history,recovery_native
    return _FixedFlow(recovery_contract,recovery_history,recovery_native)

def fixed4():
    import recovery4_contract,recovery4_history,recovery4_native
    return _FixedFlow(recovery4_contract,recovery4_history,recovery4_native)
