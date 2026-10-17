#!/usr/bin/env python3
"""Diagnostic-only historical retry overlay. No injected socket errors or gate verdicts."""
import datetime, hashlib, json, os, subprocess, sys, tempfile, signal
from pathlib import Path

root = Path.cwd()
env = {k:os.environ[k] for k in ('PATH','HOME','TMPDIR','GOCACHE','GOPATH','GOTOOLCHAIN') if k in os.environ}
fixed = "--fixed" in sys.argv
if fixed: sys.argv.remove("--fixed")
base = json.loads((Path(__file__).parent/"source-base.json").read_text())
actual = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/"ai").rglob("*.go")}
if actual != base: raise SystemExit("diagnostic replay requires the recorded Go source revision")
stamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%S.%fZ')
dest = root / '.evidence/pressure-integrity' / ('native-' + stamp)
dest.mkdir(mode=0o700, parents=True)
condition = len(sys.argv)>2 and sys.argv[2]=='condition'
if condition: (dest/'condition').mkdir()
sources = {}

write = root / 'ai/internal/testkit/provider/write_trace.go'
text = write.read_text()
if not fixed:
    text = text.replace("\t// Bound the controlled fixture's kernel queue before any HTTP writes.\n\t// Autotuned send buffers across 32 design-load streams can exhaust Darwin\n\t// mbufs; unlike ordinary TCP backpressure, ENOBUFS may hide consumed bytes\n\t// (#07). Request a small send window so writers wait for readers instead.\n\t// Custom socket-fault listeners retain their own connection semantics.\n\tif tcp, ok := c.(*net.TCPConn); ok {\n\t\tif err := tcp.SetWriteBuffer(64 << 10); err != nil {\n\t\t\t_ = c.Close()\n\t\t\treturn nil, err\n\t\t}\n\t}\n", '')
text = text.replace('import (', 'import (\n"errors"\n"syscall"\n"time"\n"os"', 1)
text = text.replace('Sequence int', 'Payload []byte `json:"payload,omitempty"`\nResume bool `json:"resume,omitempty"`\nSequence int', 1)
text = text.replace('trace    bool', 'nativeTool bool\nclosed chan struct{}\ncloseOnce sync.Once\ntrace    bool', 1)
text = text.replace('id: l.next.Add(1)', 'id: l.next.Add(1), closed: make(chan struct{})')
a = text.index('func (c *traceConn) Write(')
b = text.index('\ntype traceContextKey', a)
if fixed:
    text = text[:a] + 'func (c *traceConn) Write(p []byte) (int, error) {\n\tc.mu.Lock()\n\tdefer c.mu.Unlock()\n\tc.sequence++\n\tn, err := c.Conn.Write(p)\nif c.nativeTool && errors.Is(err, syscall.ENOBUFS) { if dir:=os.Getenv("BARNESS_NATIVE_CONDITION_DIR"); dir!="" { _=os.WriteFile(dir+"/release", []byte("native ENOBUFS"),0600) } }\n\tif c.trace && err != nil {\n\t\tdigest := sha256.Sum256(p)\n\t\tc.writes = append(c.writes, SocketWrite{Sequence: c.sequence, Offset: c.offset, Bytes: len(p), Written: n,\n\t\t\tSHA256: hex.EncodeToString(digest[:]), Error: err.Error()})\n\t}\n\tc.offset += int64(n)\n\treturn n, err\n}\n\nfunc (c *traceConn) Close() error {\n    c.closeOnce.Do(func(){ close(c.closed) })\n    return c.Conn.Close()\n}\n' + text[b:]
    text = text.replace('\n"time"', '', 1)
else:
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
                if err != nil { w.Error = err.Error(); if len(c.writes)==0 { w.Payload = append([]byte(nil), pending...) } }
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

pressure = root / 'ai/e2e/policy_pressure_test.go'
text = pressure.read_text().replace('ev.Record("stream-integrity", traces)', 'ev.Record("stream-integrity", traces)\n    // Diagnostic metadata only: no response body copies or verdict changes.\n    terminals := []map[string]any{}\n    for _, ended := range results {\n        item := map[string]any{"requestId":ended.res.Metadata.RequestID,"attempts":len(ended.res.Metadata.Attempts),"stopReason":ended.res.Message.StopReason}\n        var failure *ai.Error\n        if errors.As(ended.err,&failure) { item["code"]=failure.Code; item["phase"]=failure.Phase }\n        terminals=append(terminals,item)\n    }\n    ev.Record("native-call-terminals",terminals)')
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

env.update(BARNESS_AI_EVIDENCE_DIR=str(dest/'bundles'), BARNESS_AI_PRESSURE='1', BARNESS_AI_PRESSURE_LOOPS=sys.argv[1] if len(sys.argv)>1 else '10')
if condition:
    env['BARNESS_NATIVE_CONDITION_DIR']=str(dest/'condition')
    env['BARNESS_NATIVE_RELEASE_DELAY_MS']=sys.argv[3] if len(sys.argv)>3 else '0'
    env['BARNESS_NATIVE_PAIRS']=sys.argv[4] if len(sys.argv)>4 else '128'
    env['BARNESS_NATIVE_REFILL']=sys.argv[5] if len(sys.argv)>5 else '0'
    env['BARNESS_NATIVE_DRAIN']=sys.argv[6] if len(sys.argv)>6 else '0'
    env['BARNESS_NATIVE_KIND']=sys.argv[7] if len(sys.argv)>7 else 'tcp'
    env['BARNESS_NATIVE_CLOSE_PULSE']=sys.argv[8] if len(sys.argv)>8 else '0'
receipt = {'controlledAmbientCondition':condition, 'releaseDelayMS':env.get('BARNESS_NATIVE_RELEASE_DELAY_MS'), 'ambientPairs':env.get('BARNESS_NATIVE_PAIRS'), 'ambientRefill':env.get('BARNESS_NATIVE_REFILL'), 'ambientDrain':env.get('BARNESS_NATIVE_DRAIN'), 'ambientKind':env.get('BARNESS_NATIVE_KIND'), 'ambientClosePulse':env.get('BARNESS_NATIVE_CLOSE_PULSE'), 'intent':('current safe boundary under native contention; diagnostic only' if fixed else 'intentional historical retry RED; diagnostic only'), 'sourceCommit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,env=env,text=True).strip(), 'captureScope':'full streaming byte comparison; bounded first difference and failed wire write; no whole response copies', 'load':'original 32 concurrency, 24 text 64K tokens, 8 tool 256KiB, original budgets', 'transport':'direct controlled loopback; no external provider requests', 'sources':{}}
with tempfile.TemporaryDirectory() as temporary:
    folder = Path(temporary)
    replaces = {}
    for i,(source,patched) in enumerate(sources.items()):
        target = folder/(str(i)+'.go')
        target.write_text(patched)
        formatted = subprocess.run(['gofmt','-w',str(target)],env=env,capture_output=True,timeout=30)
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
        subprocess.run(['go','build','-o',str(ambient_binary),str(ambient_source)],env=env,check=True,capture_output=True,timeout=120)
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
    try:
        with (dest/'execution.log').open('wb') as log:
            result = subprocess.run(command,cwd=root/'ai/e2e',env=env,stdout=log,stderr=subprocess.STDOUT,timeout=920)
    finally:
        if ambient is not None:
            ambient.terminate()
            try:
                ambient.wait(timeout=10)
            except subprocess.TimeoutExpired:
                ambient.kill()
                ambient.wait()
            ambient_log.close()
    receipt.update(exitCode=result.returncode,elapsedSeconds=(datetime.datetime.now(datetime.timezone.utc)-started).total_seconds(),runtimeSourceUnchanged=all(hashlib.sha256(p.read_bytes()).hexdigest()==receipt['sources'][str(p.relative_to(root))]['originalSHA256'] for p in sources))
    (dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
    print(json.dumps({'exitCode':result.returncode,'evidence':str(dest)}),flush=True)
