"""Fixed3 binding of shared reviewed recovery mechanics; old authority stays fixed3."""

def marker(name):
    from recovery_flow import fixed3
    return fixed3().e.marker(name)

def authority():
    from recovery_flow import fixed3
    return fixed3().e.authority()

def sync_dir(path):
    from recovery_flow import fixed3
    return fixed3().e.sync_dir(path)

def locked():
    from recovery_flow import fixed3
    return fixed3().e.locked()

def summarize():
    from recovery_flow import fixed3
    return fixed3().e.summarize()

def require_checks(out):
    from recovery_flow import fixed3
    return fixed3().e.require_checks(out)

def claim(out, job, prepared, auth):
    from recovery_flow import fixed3
    return fixed3().e.claim(out,job,prepared,auth)

def run_selected(out, job):
    from recovery_flow import fixed3
    return fixed3().e.run_selected(out,job)
