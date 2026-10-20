"""Private assembly for the actual fixed recovery workflows, never a registry."""
from recovery_flow_files import Artifacts
from recovery_flow_state import State
from recovery_flow_execute import Execution
from recovery_flow_native import NativeConsumer
from recovery_flow_publication import Publication
from recovery_flow_finalize import Finalizer

class _FixedFlow:
    def __init__(self,contract,history,authority,retained):
        self.r,self.h,self.authority=contract,history,authority
        self.retained=retained
        self.files=Artifacts(contract.c)
        self.s=State(self);self.e=Execution(self);self.n=NativeConsumer(self)
        self.p=Publication(self);self.f=Finalizer(self)

def fixed3():
    import recovery_contract,recovery_history,recovery_native
    return _FixedFlow(recovery_contract,recovery_history,recovery_native,recovery_history.preserved)

def fixed4():
    import recovery4_contract,recovery4_history,recovery4_native
    return _FixedFlow(recovery4_contract,recovery4_history,recovery4_native,recovery4_history.preserved)

def fixed5():
    import recovery5_contract,recovery5_history,recovery5_native
    return _FixedFlow(recovery5_contract,recovery5_history,recovery5_native,recovery5_history.retained)

def fixed6():
    import recovery6_contract,recovery6_history,recovery6_native
    return _FixedFlow(recovery6_contract,recovery6_history,recovery6_native,recovery6_history.retained)
