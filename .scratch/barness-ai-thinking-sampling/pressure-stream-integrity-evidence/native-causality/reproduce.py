#!/usr/bin/env python3
"""Diagnostic-only historical retry overlay. No injected socket errors or gate verdicts."""
import datetime, hashlib, json, os, subprocess, sys, tempfile, signal
from pathlib import Path

root = Path(__file__).resolve().parents[4]
stamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ')
dest = root / '.evidence/pressure-integrity' / ('native-' + stamp)
dest.mkdir(mode=0o700)
condition = len(sys.argv)>2 and sys.argv[2]=='condition'
if condition: (dest/'condition').mkdir()
sources = {}

write = root / 'ai/internal/testkit/provider/write_trace.go'
text = write.read_text().replace('import (', 'import (\n"errors"\n"syscall"\n"time"\n"os"', 1)
text = text.replace('Sequence int', 'Payload []byte `json:"payload,omitempty"`\nResume bool `json:"resume,omitempty"`\nSequence int', 1)
text = text.replace('trace    bool', 'nativeTool bool\nclosed chan struct{}\ncloseOnce sync.Once\ntrace    bool', 1)
text = text.replace('id: l.next.Add(1)', 'id: l.next.Add(1), closed: make(chan struct{})')
a = text.index('func (c *traceConn) Write(')
b = text.index('\ntype traceContextKey', a)
text = text[:a] + r'''// DEBUG-native07: exact removed bounded retry, observing only native failures/resumes.
func (c *traceConn) Write(p []byte) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.sequence++
    until := time.Now().Add(time.Second)
    written := 0
    resume := false
    for {
        pending := p[written:]
        n, err := c.Conn.Write(pending)
        if c.nativeTool && errors.Is(err,syscall.ENOBUFS) { if dir:=os.Getenv("BARNESS_NATIVE_CONDITION_DIR");dir!="" { _=os.WriteFile(dir+"/release",[]byte("native ENOBUFS"),0600) } }
        if c.trace && (err != nil || resume) {
            digest := sha256.Sum256(pending)
            w := SocketWrite{Sequence:c.sequence, Offset:c.offset, Bytes:len(pending), Written:n, SHA256:hex.EncodeToString(digest[:]), Resume:resume}
            if err != nil { w.Error = err.Error(); w.Payload = append([]byte(nil), pending...) }
            c.writes = append(c.writes, w)
        }
        c.offset += int64(n)
        written += n
        if !errors.Is(err, syscall.ENOBUFS) || written == len(p) || !time.Now().Before(until) { return written, err }
        resume = true
        timer := time.NewTimer(min(time.Millisecond, time.Until(until)))
        select {
        case <-timer.C:
        case <-c.closed:
            timer.Stop()
            return written, net.ErrClosed
        }
    }
}
func (c *traceConn) Close() error {
    c.closeOnce.Do(func(){ close(c.closed) })
    return c.Conn.Close()
}
''' + text[b:]
text += '\nfunc startNativeTrace(ctx context.Context, tool bool) int64 { c:=ctx.Value(traceContextKey{}).(*traceConn); c.nativeTool=tool; return startWriteTrace(ctx) }\n'
sources[write] = text
if condition:
    provider_source = root/'ai/internal/testkit/provider/provider.go'
    sources[provider_source] = provider_source.read_text().replace('startWriteTrace(r.Context())','startNativeTrace(r.Context(), len(reply.Chunks)==65)')

integrity = root / 'ai/e2e/pressure_integrity_test.go'
text = integrity.read_text()
text = text.replace('report        pressureIntegrityReport', 'captured []byte\nobservedBytes int64\nobservedReads int\nreport        pressureIntegrityReport')
text = text.replace('f.mu.Lock()\n\tdefer f.mu.Unlock()\n\tif f.streams', '''// DEBUG-native07: copy only tool bodies; no hash/comparison/lock on the read path.
    if trace.report.ExpectedBytes < 2<<20 { trace.captured = make([]byte, 0, int(trace.report.ExpectedBytes)+(2<<20)) }
    f.mu.Lock()
    defer f.mu.Unlock()
    if f.streams''', 1)
text = text.replace('func (b *pressureIntegrity) read(p []byte) {', '''func (b *pressureIntegrity) read(p []byte) {
    b.observedBytes += int64(len(p))
    b.observedReads++
    if b.captured != nil && len(b.captured)+len(p) <= cap(b.captured) { b.captured = append(b.captured,p...) }
}
func (b *pressureIntegrity) compare(p []byte) {''', 1)
text = text.replace('trace.mu.Lock()\n\t\tr := trace.report', '''if trace.captured != nil {
            trace.compare(trace.captured)
        } else {
            trace.report.ReadBytes = trace.observedBytes
        }
        trace.mu.Lock()
        r := trace.report
        r.ReadCalls = trace.observedReads''', 1)
text += r'''
// DEBUG-native07: retain raw synthetic tool bodies only when their connection had a native error.
func (f *pressureNetworkErrors) nativeBodies(outcomes []provider.ReplyOutcome) map[string]any {
    result := map[string]any{}
    for _, o := range outcomes {
        trace := f.streams[o.RequestID]
        if trace == nil || trace.captured == nil || len(o.Writes)==0 { continue }
        result[o.RequestID] = map[string]any{"expected":bytes.Join(trace.chunks,nil),"actual":trace.captured,"observedBytes":trace.observedBytes,"captureComplete":int64(len(trace.captured))==trace.observedBytes}
    }
    return result
}
'''
text = text.replace('"sync"', '"sync"\n"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"')
sources[integrity] = text

pressure = root / 'ai/e2e/policy_pressure_test.go'
text = pressure.read_text().replace('ev.Record("stream-integrity", traces)', 'ev.Record("stream-integrity", traces)\n ev.Record("native-bodies", pw.failures.nativeBodies(outcomes))')
if condition:
    text = text.replace('for i := range l.Calls {\n\t\treply := pressureReply(textChunks)', 'var finalTools atomic.Int64\n finalReady := make(chan struct{})\n for i := range l.Calls {\n\t\treply := pressureReply(textChunks)',1)
    text = text.replace('reply.OnReceive = barrier', '\n'.join([
        'reply.OnReceive = barrier',
        '// DEBUG-native07: only the competing-buffer precondition; no stream bytes or errors injected.',
        'if isTool(i) { reply.AfterChunk = func(index int) {',
        'if index != len(toolChunks)-2 { return }',
        'if finalTools.Add(1) == int64(l.Calls/4) {',
        'dir:=os.Getenv("BARNESS_NATIVE_CONDITION_DIR")',
        '_=os.WriteFile(dir+"/start",[]byte("all tool prefixes arrived"),0600)',
        'until:=time.Now().Add(10*time.Second)',
        'for { if _,err:=os.Stat(dir+"/ready");err==nil { break };if time.Now().After(until) { break };time.Sleep(time.Millisecond) }',
        'close(finalReady)',
        '}',
        'select { case <-finalReady: case <-ctx.Done(): }',
        '} }']),1)
sources[pressure] = text

env = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH') if k in os.environ}
env.update(BARNESS_AI_EVIDENCE_DIR=str(dest/'bundles'), BARNESS_AI_PRESSURE='1', BARNESS_AI_PRESSURE_LOOPS=sys.argv[1] if len(sys.argv)>1 else '10')
if condition:
    env['BARNESS_NATIVE_CONDITION_DIR']=str(dest/'condition')
    env['BARNESS_NATIVE_RELEASE_DELAY_MS']=sys.argv[3] if len(sys.argv)>3 else '0'
    env['BARNESS_NATIVE_PAIRS']=sys.argv[4] if len(sys.argv)>4 else '128'
    env['BARNESS_NATIVE_REFILL']=sys.argv[5] if len(sys.argv)>5 else '0'
    env['BARNESS_NATIVE_DRAIN']=sys.argv[6] if len(sys.argv)>6 else '0'
    env['BARNESS_NATIVE_KIND']=sys.argv[7] if len(sys.argv)>7 else 'tcp'
    env['BARNESS_NATIVE_CLOSE_PULSE']=sys.argv[8] if len(sys.argv)>8 else '0'
receipt = {'controlledAmbientCondition':condition, 'releaseDelayMS':env.get('BARNESS_NATIVE_RELEASE_DELAY_MS'), 'ambientPairs':env.get('BARNESS_NATIVE_PAIRS'), 'ambientRefill':env.get('BARNESS_NATIVE_REFILL'), 'ambientDrain':env.get('BARNESS_NATIVE_DRAIN'), 'ambientKind':env.get('BARNESS_NATIVE_KIND'), 'ambientClosePulse':env.get('BARNESS_NATIVE_CLOSE_PULSE'), 'intent':'natural OS errors only; historical retry in temporary diagnostic overlay', 'sourceCommit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(), 'captureScope':'tool bodies below 2 MiB; text bytes counted only', 'load':'original 32 concurrency, 24 text 64K tokens, 8 tool 256KiB, original budgets', 'proxy':'loopback direct as historical command; system external proxy 127.0.0.1:7890', 'sources':{}}
with tempfile.TemporaryDirectory() as temporary:
    folder = Path(temporary)
    replaces = {}
    for i,(source,patched) in enumerate(sources.items()):
        target = folder/(str(i)+'.go')
        target.write_text(patched)
        formatted = subprocess.run(['gofmt','-w',str(target)],capture_output=True)
        if formatted.returncode: raise SystemExit(formatted.stderr.decode())
        patched = target.read_text()
        (dest/(source.name+'.txt')).write_text(patched)
        replaces[str(source)] = str(target)
        receipt['sources'][str(source.relative_to(root))] = {'originalSHA256':hashlib.sha256(source.read_bytes()).hexdigest(),'overlaySHA256':hashlib.sha256(patched.encode()).hexdigest()}
    overlay = folder/'overlay.json'
    overlay.write_text(json.dumps({'Replace':replaces}))
    binary = folder/'native.test'
    build = subprocess.run(['go','test','-c','-overlay',str(overlay),'-o',str(binary),'./ai/e2e'],cwd=root,env=env,capture_output=True,timeout=120)
    (dest/'build.log').write_bytes(build.stdout+build.stderr)
    if build.returncode: raise SystemExit('diagnostic overlay did not compile: '+str(dest))
    ambient=None
    if condition:
        ambient_source=folder/'ambient.go'
        ambient_source.write_bytes((Path(__file__).parent/'ambient.go.txt').read_bytes())
        ambient_binary=folder/'ambient'
        subprocess.run(['go','build','-o',str(ambient_binary),str(ambient_source)],check=True,capture_output=True)
        (dest/'ambient-source.go.txt').write_bytes(ambient_source.read_bytes())
        receipt['ambientSHA256']=hashlib.sha256(ambient_binary.read_bytes()).hexdigest()
        ambient_log=(dest/'ambient.json').open('wb')
        ambient=subprocess.Popen([str(ambient_binary)],env=env,stdout=ambient_log,stderr=subprocess.STDOUT)
    receipt['binarySHA256'] = hashlib.sha256(binary.read_bytes()).hexdigest()
    command = [str(binary),'-test.run=^TestPressureLoopback$','-test.timeout=15m','-test.v']
    receipt['command'] = ['[OVERLAY_BINARY]'] + command[1:]
    (dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
    print(str(dest),flush=True)
    started = datetime.datetime.now(datetime.timezone.utc)
    with (dest/'execution.log').open('wb') as log:
        result = subprocess.run(command,cwd=root/'ai/e2e',env=env,stdout=log,stderr=subprocess.STDOUT,timeout=920)
    if ambient is not None:
        ambient.send_signal(signal.SIGTERM)
        ambient.wait(timeout=10)
        ambient_log.close()
    receipt.update(exitCode=result.returncode,elapsedSeconds=(datetime.datetime.now(datetime.timezone.utc)-started).total_seconds(),runtimeSourceUnchanged=all(hashlib.sha256(p.read_bytes()).hexdigest()==receipt['sources'][str(p.relative_to(root))]['originalSHA256'] for p in sources))
    (dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
    print(json.dumps({'exitCode':result.returncode,'evidence':str(dest)}),flush=True)
