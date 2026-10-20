"""One pinned CN spend inventory; confers no current target/proof authority."""
import hashlib
import paid_core as c
from paid_capture import bounded

NAME='mm-research-m3-chat'
INTENT_SHA='cfac130565e33b804d22eb2bd7feb58e66d8d6baa12e2094edc4cdcf9f6dfff0'
PREFIX_SHA='222c68c07ce06819b6c84f6aa739b0a08824f1b70893bae7accd014bcae2b4f8'
PREFIX_BYTES=1804
SOURCE='9481a6ad8028025a7f35da19a3e488935c6cd897'


def inventory():
    # These immutable pins describe already consumed accounting only. Missing
    # history is unknown, including in a new clone; it never grants a reset.
    known=False
    try:
        root=c.common_root()/'.evidence'
        intent=bounded(root/'minimax-launches'/(NAME+'.json'),2<<20)
        ledger=bounded(root/'minimax-stages/candidate.jsonl',1<<20)
        known=(hashlib.sha256(intent).hexdigest()==INTENT_SHA and
               hashlib.sha256(ledger[:PREFIX_BYTES]).hexdigest()==PREFIX_SHA)
    except (OSError,RuntimeError):pass
    row=dict(route=NAME,provider='minimax-cn',stage='candidate',
        budgetSlot='chat/MiniMax-M3/standard',accountingOnly=True,
        state='REPORTED' if known else 'UNKNOWN',status='FAIL' if known else 'UNKNOWN',
        calls=2 if known else None,attempts=2 if known else None,
        currentReportBudget=dict(calls=2,attempts=2,requestedOutput=5120,reservedInput=4096) if known else {},
        sourceCommit=SOURCE,intentSha256=INTENT_SHA,ledgerPrefixSha256=PREFIX_SHA,
        maxCalls=20,maxAttempts=24,conservativeCalls=2 if known else 20,
        conservativeAttempts=2 if known else 24)
    return dict(accountingOnly=True,knownCalls=2 if known else None,
        knownAttempts=2 if known else None,routes=[row])


def require_known():
    if inventory()['knownCalls'] is None:
        raise RuntimeError('immutable historical accounting missing or changed; full slot consumed')
