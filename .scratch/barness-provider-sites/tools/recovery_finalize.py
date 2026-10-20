"""Fixed3 binding of shared reviewed recovery mechanics; old authority stays fixed3."""

def plan(out, selected):
    from recovery_flow import fixed3
    return fixed3().f.plan(out,selected)

def finalize(out, selected):
    from recovery_flow import fixed3
    return fixed3().f.finalize(out,selected)
