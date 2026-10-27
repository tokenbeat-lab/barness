"""Bounded process capture with group cancellation and credential redaction."""
import codecs
import json
import os
import shutil
import signal
import subprocess
import selectors
import time
from live_environment import selected_environment

ALLOW = ('PATH', 'HOME', 'TMPDIR', 'GOCACHE', 'GOPATH', 'GOROOT', 'GOTOOLCHAIN',
         'GOOS', 'GOARCH', 'CGO_ENABLED', 'CC', 'CXX', 'SYSTEMROOT')
LIMIT = 64 << 20


def environment(stage, route=None):
    env = {key: os.environ[key] for key in ALLOW if key in os.environ}
    env.update(GOENV='off', GOFLAGS='')
    secrets = []
    if stage == 'pressure':
        env['BARNESS_AI_PRESSURE'] = '1'
        if 'BARNESS_AI_PRESSURE_LOOPS' in os.environ:
            value = int(os.environ['BARNESS_AI_PRESSURE_LOOPS'])
            if not 1 <= value <= 100:
                raise ValueError('pressure loop count outside bounds')
            env['BARNESS_AI_PRESSURE_LOOPS'] = str(value)
    if stage == 'live':
        selected, secrets = selected_environment(route)
        env.update(selected)
    return env, secrets


def isolate_go_home(env, stage, out):
    if stage == 'live':
        return
    # Legacy child launchers retain HOME but may rebuild their environment.
    # An empty stage-owned home prevents them reopening the user's Go config.
    go = shutil.which('go', path=env.get('PATH'))
    if go:
        caches = json.loads(subprocess.check_output(
            [go, 'env', '-json', 'GOCACHE', 'GOPATH'], env=env, timeout=10))
        for name in ('GOCACHE', 'GOPATH'):
            env.setdefault(name, caches[name])
    home = out / 'home'
    home.mkdir(mode=0o700)
    env['HOME'] = str(home)


def capture(argv, cwd, env, out, timeout, secrets, shutdown_grace=5):
    if not 0 < shutdown_grace <= 5:
        raise ValueError('shutdown grace outside bounds')
    state = {'exitCode': None, 'status': 'START_FAILED', 'error': None}
    try:
        process = subprocess.Popen(argv, cwd=cwd, env=env, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, start_new_session=True)
    except OSError:
        state['error'] = 'process unavailable'
        return state
    deadline = time.monotonic() + timeout
    selector = selectors.DefaultSelector()
    streams = []
    hold = max([len(x) for x in secrets] or [1]) - 1

    def signal_group(signum):
        try:
            os.killpg(process.pid, signum)
            return True
        except ProcessLookupError:
            return False
        except PermissionError:
            # Darwin can report EPERM for a reaped group; never target another PID.
            if process.poll() is None:
                process.send_signal(signum)
            return False

    def stop():
        # Nested capture owners must receive cancellation before their parent
        # group is killed: they own separate native groups and finalize those
        # in their SIGTERM handler. Probe only this owned group, never a global
        # process tree. A non-cooperative member still has a finite grace. Nested owners
        # use a shorter grace than their enclosing collector so force-killing
        # the leaf group always precedes the outer owner deadline.
        if not signal_group(signal.SIGTERM):
            return
        until = time.monotonic() + shutdown_grace
        while time.monotonic() < until:
            process.poll()  # Reap the direct child before probing its group.
            try:
                os.killpg(process.pid, 0)
            except ProcessLookupError:
                return
            except PermissionError:
                if process.poll() is not None:
                    return
                break
            time.sleep(.01)
        signal_group(signal.SIGKILL)

    def write(stream, block, final=False):
        stream['total'] += len(block)
        if stream['total'] > LIMIT:
            raise ValueError('capture limit exceeded')
        pending = stream['pending'] + stream['decoder'].decode(block, final=final)
        cut = len(pending) if final else max(0, len(pending) - hold)
        for secret in secrets:
            pos = pending.find(secret)
            while pos >= 0:
                if pos < cut < pos + len(secret):
                    cut = pos
                pos = pending.find(secret, pos + len(secret))
        ready, stream['pending'] = pending[:cut], pending[cut:]
        for secret in secrets:
            ready = ready.replace(secret, '[REDACTED]')
        stream['log'].write(ready)
        stream['log'].flush()

    previous = signal.getsignal(signal.SIGTERM)
    def cancelled(_signum, _frame):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, cancelled)
    try:
        for pipe, name in ((process.stdout, 'stdout.jsonl'), (process.stderr, 'stderr.log')):
            stream = {'pipe': pipe, 'log': (out / name).open('w'), 'total': 0,
                      'pending': '', 'decoder': codecs.getincrementaldecoder('utf-8')('replace')}
            streams.append(stream)
            os.set_blocking(pipe.fileno(), False)
            selector.register(pipe, selectors.EVENT_READ, stream)
        while selector.get_map() or process.poll() is None:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                if process.poll() is None:
                    state['status'] = 'TIMEOUT'
                else:
                    state.update(status='CAPTURE_FAILED', error='pipe remained open after child exit')
                break
            for key, _ in selector.select(min(.05, remaining)):
                stream = key.data
                block = os.read(key.fd, 65536)
                write(stream, block, final=not block)
                if not block:
                    selector.unregister(key.fileobj)
                    key.fileobj.close()
        else:
            state['status'] = 'PASS' if process.returncode == 0 else 'FAIL'
    except KeyboardInterrupt:
        state['status'] = 'CANCELLED'
    except (OSError, ValueError):
        state.update(status='CAPTURE_FAILED', error='process capture failed')
    finally:
        # The deadline covers both process lifetime and inherited pipe handles.
        # No blocking reader threads survive finalization, even for detached children.
        try:
            stop()
            state['exitCode'] = process.wait(timeout=5)
        except (OSError, subprocess.TimeoutExpired):
            state.update(status='CAPTURE_FAILED', exitCode=process.poll(), error='process wait failed')
        selector.close()
        for stream in streams:
            stream['pipe'].close()
            try:
                write(stream, b'', final=True)
                os.fsync(stream['log'].fileno())
            except (OSError, ValueError):
                state.update(status='CAPTURE_FAILED', error='capture finalization failed')
            finally:
                stream['log'].close()
        signal.signal(signal.SIGTERM, previous)
    return state
