"""Shared private fixed recovery publication mechanics; authority values come only from compiled contexts."""
from pathlib import Path

class Publication:

    def __init__(self, flow):
        self.flow = flow

    def roles(self, out):
        self.flow.s.guard(out)
        summary = self.flow.e.summarize()
        if summary['actualProcessCount'] != self.flow.r.ORIGINAL_PROCESS_COUNT + len(self.flow.r.jobs(out)) or len(summary['routes']) != len(self.flow.r.jobs(out)) or any((v['state'] != 'COMPLETED' or v['budget']['status'] != 'PASS' for v in summary['routes'])):
            raise RuntimeError('all new fixed full processes must pass; all historical records stay retained')
        selected = {}
        for combo in self.flow.r.REGIONAL:
            name = 'candidate-' + combo
            p = self.flow.r.c.common_root() / '.evidence/minimax-launches' / (self.flow.r.old.ROUND + '-' + name + '.json')
            value = self.flow.s.doc(p)
            if value.get('state') != 'COMPLETED' or value.get('exitCode') != 0 or value.get('budget', {}).get('status') != 'PASS' or (not value.get('nativeConsumerReceiptSha256')):
                raise RuntimeError('six genuinely promoted historical candidate roles required')
            selected[name] = dict(round=self.flow.r.old.ROUND, receiptSHA256=self.flow.s.pin(p), historical=True)
        if self.flow.r.RETAINED_GENERIC:
            selected.update(self.flow.retained()['roles'])
        else:
            import final_publication
            self.flow.h.preserved()
            for route in self.flow.r.ROUTES:
                receipt, sha = final_publication.require(route)
                selected['glm-' + route] = dict(round=self.flow.r.old.ROUND, receiptSHA256=sha, historical=True, measuredCommit=receipt['sourceCommit'])
        for name, job in self.flow.r.jobs(out).items():
            receipt = self.flow.s.doc(self.flow.e.marker(name))
            command = out / (name + '-command.json')
            if receipt.get('sourceCommit') != self.flow.s.doc(out / 'prepared.json')['sourceCommit'] or receipt.get('collectionRoot') != self.flow.r.c.redact(str(out)) or receipt.get('exitCode') != 0 or (receipt.get('commandReceiptSha256') != self.flow.s.pin(command)) or (self.flow.s.doc(command).get('launcherAudit') != dict(passed=True, findings=[])):
                raise RuntimeError('all own current recovery native process receipts required')
            self.flow.n.require_consumed(out, job, receipt)
            selected[name] = dict(round=self.flow.r.ROUND, receiptSHA256=self.flow.s.pin(self.flow.e.marker(name)), historical=False)
        if len(selected) != 33:
            raise RuntimeError('exact33 effective full-process roles required')
        return selected

    def own_base(self, out, name):
        import final_execute
        job = self.flow.r.jobs(out)[name]
        receipt = self.flow.s.doc(self.flow.e.marker(name))
        base = final_execute.report_path(out, job).parent
        if receipt.get('reportedFileSha256') != self.flow.s.pin(final_execute.report_path(out, job)):
            raise RuntimeError('own completed report changed')
        return base

    def bind_inputs(self, out, result):
        import final_publication
        import renewal_common as u
        selected = self.roles(out)
        live = result['liveInputs']
        effects = result['effects']
        for name, base in live.items():
            if name == 'typesafe-classifier-jev-1.14.0' or (name.startswith('zhipu-') and not self.flow.r.RETAINED_GENERIC):
                continue
            if name in self.flow.r.RETAINED_GENERIC:
                retained=self.flow.retained()['roles'][name]
                if Path(base)!=Path(retained['base']):raise RuntimeError('retained generic must equal exact closed4 canonical raw input')
                continue
            jobname = 'glm-'+name.removeprefix('zhipu-') if name.startswith('zhipu-') else 'formal-' + name if name in self.flow.r.REGIONAL else 'mm-global-formal-m3-' + name.removeprefix('minimax-') if name.startswith('minimax-') else name
            if Path(base) != self.own_base(out, jobname):
                raise RuntimeError('final input must be its new recovery canonical own capture')
        for name, base in effects.items():
            if Path(base) != self.own_base(out, name):
                raise RuntimeError('effect must be its current recovery canonical own capture')
        protocol = self.own_base(out, 'protocol114')
        visual = self.own_base(out, 'visual-jev')
        jev = Path(live['typesafe-classifier-jev-1.14.0'])
        for target, source in [(jev / 'protocol', protocol), (jev / 'effects', visual)]:
            if self.flow.files.tree(target) != self.flow.files.tree(source):
                raise RuntimeError('Jev derivation must retain exact current recovery protocol and visual trees')
        if not self.flow.r.RETAINED_GENERIC:
            index = self.flow.files.doc(self.flow.r.c.ROOT / 'ai/live/testdata/glm-formal/live-captures.json')
            views = self.flow.files.doc(self.flow.r.c.ROOT / 'ai/live/testdata/glm-formal/support-views.json')
            final_publication.validate(index, views)
            from repair_publication import zipped
            for route in self.flow.r.ROUTES:
                base = Path(live['zhipu-' + route])
                actual = self.flow.files.tree(base)
                archived, members = zipped(views['routes'][route], list(actual))
                if members != set(actual) or actual != archived:
                    raise RuntimeError('preserved GLM support input must equal indexed original derivation')
        return selected
