#!/usr/bin/env python3
"""Replay the removed retry through a temporary Go overlay; never edit runtime source."""
import datetime, hashlib, json, os, subprocess, tempfile
from pathlib import Path

root = Path(__file__).resolve().parents[4]
source = root / 'ai/internal/testkit/provider/write_trace.go'
text = source.read_text()
start = text.index('func (c *traceConn) Write(p []byte) (int, error) {')
end = text.index('\ntype traceContextKey', start)
legacy = r'''func (c *traceConn) Write(p []byte) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.sequence++
    until := time.Now().Add(time.Second)
    written := 0
    for {
        n, err := c.Conn.Write(p[written:])
        if c.trace {
            digest := sha256.Sum256(p[written:])
            w := SocketWrite{Sequence:c.sequence, Offset:c.offset, Bytes:len(p)-written, Written:n, SHA256:hex.EncodeToString(digest[:])}
            if err != nil { w.Error = err.Error() }
            c.writes = append(c.writes, w)
        }
        c.offset += int64(n)
        written += n
        if !errors.Is(err, syscall.ENOBUFS) || written == len(p) || !time.Now().Before(until) {
            return written, err
        }
        time.Sleep(time.Millisecond)
    }
}
'''
patched = text[:start] + legacy + text[end:]
patched = patched.replace('import (', 'import (\n"errors"\n"syscall"\n"time"', 1)
base = root / '.evidence/pressure-integrity' / ('tool-red-replay-' + datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ'))
base.mkdir(mode=0o700, exist_ok=False)
(base/'overlaid-write-trace.go.txt').write_text(patched)
env = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ}
env['BARNESS_AI_EVIDENCE_DIR'] = str(base/'bundles')
with tempfile.TemporaryDirectory() as temporary:
    folder = Path(temporary)
    overlay_source = folder/'write_trace.go'
    overlay_source.write_text(patched)
    overlay = folder/'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(source):str(overlay_source)}}))
    binary = folder/'red.test'
    build = subprocess.run(['go','test','-c','-overlay',str(overlay),'-o',str(binary),'./ai/e2e'], cwd=root, env=env, capture_output=True, timeout=120)
    if build.returncode: raise SystemExit('red overlay did not compile')
    binary_hash = hashlib.sha256(binary.read_bytes()).hexdigest()
    result = subprocess.run([str(binary),'-test.run=^TestPressureToolWriteIntegrity$/^(json-envelope|argument-limit)$','-test.timeout=60s'], cwd=root/'ai/e2e', env=env, capture_output=True, timeout=65)
    output = (result.stdout+result.stderr).replace(str(root).encode(),b'[REPO]').replace(str(Path.home()).encode(),b'[HOME]')
    (base/'execution.log').write_bytes(output)
receipts = []
for mode, offset, frame, frame_offset in [('tool-write-json-envelope',1580479,4099,262033),('tool-write-argument-limit',1318335,4098,262214)]:
    paths = list((base/'bundles').glob('*/E08-pressure-'+mode+'/stream-integrity.json'))
    if len(paths) != 1: raise SystemExit('missing independent byte evidence')
    trace = json.loads(paths[0].read_text())[0]
    difference = trace.get('firstDifference', {})
    if (difference.get('offset'),difference.get('frame'),difference.get('frameOffset')) != (offset,frame,frame_offset):
        raise SystemExit('red differed from the retained counterexample')
    receipts.append({'mode':mode,'offset':offset,'frame':frame,'frameOffset':frame_offset})
if result.returncode == 0: raise SystemExit('unsafe overlay unexpectedly passed')
receipt = {'intent':'intentional regression RED only; removed unsafe retry in a temporary overlay','exitCode':result.returncode,'binarySHA256':binary_hash,'overlaySourceSHA256':hashlib.sha256(patched.encode()).hexdigest(),'runtimeSourceUnchanged':source.read_text()==text,'verifiedDifferences':receipts}
(base/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({'redReproduced':True,'evidence':str(base.relative_to(root)),'runtimeSourceUnchanged':receipt['runtimeSourceUnchanged']}))
