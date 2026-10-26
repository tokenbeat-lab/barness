"""Forward the selected native producer's contract without acquiring authority."""
import os


def selected_environment(route):
    selectors = ('BARNESS_REGIONAL_ROUTE', 'BARNESS_MINIMAX_ROUTE', 'BARNESS_GLM_ROUTE',
                 'BARNESS_MM_ROUTE', 'BARNESS_AI_LIVE_COMBO')
    selected = [name for name in selectors if os.environ.get(name)]
    if len(selected) != 1:
        raise ValueError('one explicitly selected native live route required')
    selector = selected[0]
    value = os.environ[selector]
    switch = 'BARNESS_AI_LIVE'
    if selector == 'BARNESS_AI_LIVE_COMBO':
        identity = value
        key = 'BARNESS_AI_LIVE_KEY_' + value.upper().replace('-', '_')
        names = [selector, key] + ['BARNESS_AI_LIVE_' + suffix for suffix in
                 ('ACCOUNT_ALIAS', 'REGION_ALIAS', 'PERMISSIONS_FILE', 'ALLOW_UNKNOWN_ACCOUNT', 'PROXY')]
    elif selector == 'BARNESS_MM_ROUTE':
        switch = 'BARNESS_MM_LIVE'
        identity = 'typesafe-classifier-jev-1.14.0' if value == 'jev' and os.environ.get('BARNESS_MM_FORMAL') == '1' else 'multimodal-' + value
        key = 'BARNESS_MM_KEY'
        names = [selector, key, 'BARNESS_MM_FORMAL', 'BARNESS_MM_AUTHORIZATION', 'BARNESS_AI_LIVE_PROXY']
    else:
        prefix = selector.removesuffix('ROUTE')
        identity = {'BARNESS_REGIONAL_': value, 'BARNESS_GLM_': 'zhipu-' + value,
                    'BARNESS_MINIMAX_': 'minimax-' + value}[prefix]
        suffixes = {'BARNESS_REGIONAL_': ('SERVICE', 'GRANT', 'STAGE', 'PROXY'),
                    'BARNESS_GLM_': ('FORMAL', 'GRANT', 'ACCOUNT_ALIAS', 'REGION_ALIAS', 'PROXY'),
                    'BARNESS_MINIMAX_': ('MODEL', 'PURPOSE', 'FORMAL', 'RESEARCH', 'LEDGER', 'GRANT',
                                         'ACCOUNT_ALIAS', 'REGION_ALIAS', 'ENTITLEMENT', 'FORMAL_RENEWAL', 'PROXY')}[prefix]
        key = prefix + 'KEY'
        names = [selector, key] + [prefix + suffix for suffix in suffixes]
    if route != identity or os.environ.get(switch) != '1':
        raise ValueError('route or native live switch does not match selection')
    env = {name: os.environ[name] for name in names if name in os.environ}
    env[switch] = '1'
    return env, [env[key]] if env.get(key) else []
