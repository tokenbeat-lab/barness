#!/usr/bin/env python3
"""Private public-wrapper timing fixture. Only the controlled native argv is intercepted."""
import json
import os
from pathlib import Path
import subprocess
import sys
import time

config = json.loads(Path(__file__).with_name('config.json').read_text())
args = sys.argv[1:]
controlled = args[:4] == ['test', '-tags', 'live', './ai/live']
if not controlled:
    os.execv(config['go'], [config['go'], *args])
Path(config['started']).write_text('controlled-native-start\n')
if config['mode'] == 'slow-start':
    time.sleep(11)
    os.execv(config['go'], [config['go'], *args])
# The heartbeat inherits both pipes and the process group, just like a native
# test executable under `go test`. It must stop before its wrapper returns.
child = subprocess.Popen([sys.executable, '-c', '''
from pathlib import Path
import sys, time
path = Path(sys.argv[1])
while True:
    with path.open('a') as out:
        out.write('heartbeat\\n')
    time.sleep(.03)
''', config['heartbeat']])
Path(config['pids']).write_text(json.dumps([os.getpid(), child.pid]))
child.wait()
