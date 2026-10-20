"""Fixed3 binding of shared reviewed recovery mechanics; old authority stays fixed3."""
from recovery_flow_state import PINNED

def path(value):
    from recovery_flow import fixed3
    return fixed3().s.path(value)

def doc(p):
    from recovery_flow import fixed3
    return fixed3().s.doc(p)

def pin(p):
    from recovery_flow import fixed3
    return fixed3().s.pin(p)

def prepare(out):
    from recovery_flow import fixed3
    return fixed3().s.prepare(out)

def guard(out):
    from recovery_flow import fixed3
    return fixed3().s.guard(out)

def approved(out, job):
    from recovery_flow import fixed3
    return fixed3().s.approved(out,job)

def checks(out):
    from recovery_flow import fixed3
    return fixed3().s.checks(out)
