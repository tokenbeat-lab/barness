"""Fixed3 binding of shared reviewed recovery mechanics; old authority stays fixed3."""

def roles(out):
    from recovery_flow import fixed3
    return fixed3().p.roles(out)

def own_base(out, name):
    from recovery_flow import fixed3
    return fixed3().p.own_base(out,name)

def bind_inputs(out, result):
    from recovery_flow import fixed3
    return fixed3().p.bind_inputs(out,result)
