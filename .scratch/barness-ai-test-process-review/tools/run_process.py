"""Bounded process capture with group cancellation and credential redaction."""
import codecs
import os
import signal
import subprocess
import threading

ALLOW = ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'GOROOT', 'GOTOOLCHAIN',
         'GOOS', 'GOARCH', 'CGO_ENABLED', 'CC', 'CXX', 'SYSTEMROOT')
LIMIT = 64 << 20


def environment(stage):
    env = {key: os.environ[key] for key in ALLOW if key in os.environ}
    secrets = []
    if stage == 'pressure':
        env['BARNESS_AI_PRESSURE'] = '1'
        if 'BARNESS_AI_PRESSURE_LOOPS' in os.environ:
            value = int(os.environ['BARNESS_AI_PRESSURE_LOOPS'])
            if not 1 <= value <= 100:
                raise ValueError('pressure loop count outside bounds')
            env['BARNESS_AI_PRESSURE_LOOPS'] = str(value)
    if stage == 'live':
        # Only explicitly selected route variables cross the process boundary.
        families = ('BARNESS_REGIONAL_', 'BARNESS_MINIMAX_', 'BARNESS_GLM_', 'BARNESS_MM_')
        selected = [p for p in families if os.environ.get(p + 'ROUTE')]
        if len(selected) != 1 or os.environ.get('BARNESS_AI_LIVE') != '1':
            raise ValueError('one explicitly selected live route required')
        prefix = selected[0]
        allowed = ('ROUTE', 'SERVICE', 'GRANT', 'KEY', 'STAGE', 'FORMAL', 'MODEL', 'PURPOSE', 'FORMAL_RENEWAL', 'PROXY')
        for suffix in allowed:
            key = prefix + suffix
            if key in os.environ:
                env[key] = os.environ[key]
                if suffix == 'KEY':
                    secrets.append(env[key])
        env['BARNESS_AI_LIVE'] = '1'
    return env, secrets


def capture(argv, cwd, env, out, timeout, secrets):
    state = {'exitCode': None, 'status': 'START_FAILED', 'error': None}
    try:
        process = subprocess.Popen(argv, cwd=cwd, env=env, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, start_new_session=True)
    except OSError:
        state['error'] = 'process unavailable'
        return state
    errors = []
    def stop():
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    def pump(stream, target):
        decoder = codecs.getincrementaldecoder('utf-8')('replace')
        pending = ''
        total = 0
        hold = max([len(x) for x in secrets] or [1]) - 1
        try:
            with target.open('w') as log:
                while block := stream.read(65536):
                    total += len(block)
                    if total > LIMIT:
                        raise ValueError('capture limit exceeded')
                    pending += decoder.decode(block)
                    # Keep incomplete secret prefixes before writing a bounded chunk.
                    cut = max(0, len(pending) - hold)
                    for secret in secrets:
                        pos = pending.find(secret)
                        while pos >= 0:
                            if pos < cut < pos + len(secret):
                                cut = pos
                            pos = pending.find(secret, pos + len(secret))
                    ready, pending = pending[:cut], pending[cut:]
                    for secret in secrets:
                        ready = ready.replace(secret, '[REDACTED]')
                    log.write(ready)
                pending += decoder.decode(b'', final=True)
                for secret in secrets:
                    pending = pending.replace(secret, '[REDACTED]')
                log.write(pending)
                log.flush()
                os.fsync(log.fileno())
        except Exception:
            errors.append('process capture failed')
            stop()
        finally:
            stream.close()
    threads = [threading.Thread(target=pump, args=(process.stdout, out / 'stdout.jsonl')),
               threading.Thread(target=pump, args=(process.stderr, out / 'stderr.log'))]
    for thread in threads:
        thread.start()
    previous = signal.getsignal(signal.SIGTERM)
    def cancelled(_signum, _frame):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, cancelled)
    try:
        state['exitCode'] = process.wait(timeout=timeout)
        state['status'] = 'PASS' if state['exitCode'] == 0 else 'FAIL'
    except subprocess.TimeoutExpired:
        stop()
        state.update(status='TIMEOUT', exitCode=process.wait())
    except KeyboardInterrupt:
        stop()
        state.update(status='CANCELLED', exitCode=process.wait())
    finally:
        signal.signal(signal.SIGTERM, previous)
        # A descendant holding a pipe cannot extend the stage indefinitely.
        stop()
        for thread in threads:
            thread.join(timeout=5)
        if any(thread.is_alive() for thread in threads) or errors:
            state.update(status='CAPTURE_FAILED', error='process capture failed')
    return state
