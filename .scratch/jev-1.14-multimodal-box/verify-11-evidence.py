#!/usr/bin/env python3
"""Independently read one immutable candidate package; never import formal support.

No network, credentials, or working-source equality required for historical runs.
Use --require-live to reject controlled HTTP packages. NOT_RUN/FAIL exit nonzero.
"""
import argparse
import base64
import hashlib
import json
from pathlib import Path
from datetime import datetime

PINS = {
    "jev": "9ac5655199e22d45db2fd5b72af79ef597068a89f4298fbc47c539ed4f6426a4",
    "decisions": "d5cc54e4189c8075d031180fc6f9456e06306fa0ca7e8ae99d6dab776450712a",
}
LIMITS = dict(calls=12, attempts=12, questions=32, sources=24, references=32,
              sourceBytes=8 << 20, expandedBytes=16 << 20, stateBytes=4 << 20)


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def pairs(items):
    result = {}
    for key, value in items:
        require(key not in result, "duplicate JSON field")
        result[key] = value
    return result


def decode(data):
    return json.loads(data, object_pairs_hook=pairs)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def compact(value):
    return json.dumps(value, ensure_ascii=False, separators=(",", ":")).encode()


def expected(route):
    names = ["default-context", "explicit-context", "ordered-reuse"]
    if route == "jev":
        names += ["local-state-instructions", "visual-criteria", "box-existence"]
    return [f"{name}-{entry}" for entry in ("plain", "hooked") for name in names]


def validate_mapping(req, native, route, hooked, fixture_images):
    """Expected projection comes from public slots, not the final native tree."""
    public_images = req["images"]
    # Hook exchanges the two known fixture files, including single-image calls.
    def content(image):
        return (image["mimeType"], image["data"])
    def selected(index):
        return content(fixture_images[1-index] if hooked else fixture_images[index])
    refs = req.get("imageReferences") or [dict(imageIndex=i, placement="context")
                                           for i in range(len(public_images))]
    context = [selected(r["imageIndex"]) for r in refs if r["placement"] == "context"]
    if route == "decisions":
        blocks = native["input"][0]["content"]
        actual = []
        for b in blocks[1:]:
            head, data = b["image_url"].split(",", 1)
            actual.append((head[5:-7], data))
        require(actual == context, "public context order/reuse changed")
        require(blocks[0]["text"] == compact(req["state"]).decode(), "Decisions State changed")
        return
    projected = decode(compact(dict(state=req["state"], questions=req["questions"])))
    for ref in refs:
        if ref["placement"] != "value":
            continue
        parts = [p.replace("~1", "/").replace("~0", "~") for p in ref["pointer"].split("/")[1:]]
        cursor = projected
        for part in parts[:-1]:
            cursor = cursor[int(part)] if isinstance(cursor, list) else cursor[part]
        key = int(parts[-1]) if isinstance(cursor, list) else parts[-1]
        require(cursor[key] is None, "public image slot is not null")
        cursor[key] = selected(ref["imageIndex"])
    assets = native["assets"]
    def restore(value):
        if isinstance(value, str) and value in assets:
            a = assets[value]
            return ("image/"+a["format"].replace("jpg", "jpeg"), a["data_base64"])
        if isinstance(value, dict):
            return {k: restore(v) for k, v in value.items()}
        if isinstance(value, list):
            return [restore(v) for v in value]
        return value
    state = restore(native["state"])
    if context:
        require(state == dict(state=projected["state"], images=context), "Jev context mapping changed")
    else:
        require(state == projected["state"], "Jev State position changed")
    for name, q in projected["questions"].items():
        final = native["questions"][name]
        require(restore(final["instructions"]) == q["instructions"], "instructions scope changed")
        if "criteria" in q:
            require(restore(final["criteria"]) == q["criteria"], "visual criteria scope changed")


def verify(root, live):
    root = root.resolve()

    def read(name):
        path = root / name
        require(path.is_file() and not path.is_symlink(), "missing/nonregular artifact")
        require(path.stat().st_size <= 16 << 20, "oversized artifact")
        return path.read_bytes()

    inventory = decode(read("multimodal-integrity.json"))
    require(inventory["schema"] == 1, "inventory schema")
    files = inventory["files"]
    actual_files = set()
    for path in root.rglob("*"):
        require(not path.is_symlink(), "symlink in package")
        if path.is_file() and path.name != "multimodal-integrity.json":
            actual_files.add(path.relative_to(root).as_posix())
    require(set(files) == actual_files, "inventory missing/extra artifacts")
    for name, digest in files.items():
        require(not Path(name).is_absolute() and ".." not in Path(name).parts,
                "inventory path escapes package")
        require(sha(read(name)) == digest, "artifact checksum mismatch")
    report = decode(read("multimodal-report.json"))
    route = report["route"]
    require(route in PINS and report["schema"] == 1 and
            report["kind"] == "multimodal-candidate-only", "wrong candidate identity")
    require(report["mode"] in ("controlled", "live"), "unknown provenance")
    require(not live or report["mode"] == "live", "controlled package is not live evidence")
    require(report["outcome"] == "PASS" and report["complete"] is True and
            report["auditPassed"] is True, "package has no accepting verdict")
    require(decode(read("audit.json")) == {"findings": []}, "audit failed")
    manifest = decode(read("manifest.json"))
    cases = {case["id"]: case for case in manifest["cases"]}
    require(len(cases) == len(manifest["cases"]), "duplicate manifest case")
    require(report["expected"] == expected(route) == report["scenarios"], "incomplete pipeline")
    require(set(cases) == {"MM11-configuration", *["MM11-"+s for s in expected(route)]},
            "manifest scenario set differs")
    for case in cases.values():
        require(case["status"] == "PASS", "manifest contains non-PASS case")
        prefix = case["dir"] + "/"
        for name, digest in case["artifacts"].items():
            require(files.get(prefix+name) == digest, "manifest artifacts differ")
        assertions = decode(read(prefix+"assertions.json"))
        require(assertions["status"] == "PASS", "assertions not PASS")
        if case["id"] != "MM11-configuration":
            require(assertions["assertions"] and all(a["pass"] is True for a in assertions["assertions"]),
                    "missing/failed assertions")
        for name, digest in case["fixtures"].items():
            require(sha(read(prefix+name)) == digest, "fixture hash differs")
    catalog_bytes = read("MM11-configuration/candidate.json")
    require(sha(catalog_bytes) == PINS[route] == report["catalogFileHash"], "wrong pinned candidate")
    cat = decode(catalog_bytes)
    require(cat["version"] == report["catalogVersion"], "catalog version differs")
    config = decode(read("MM11-configuration/host-configuration.json"))
    require(all(v == 0 for v in decode(read("MM11-configuration/resources.json")).values()),
            "resources retained")
    require(config["catalog"]["version"] == report["catalogVersion"], "Host catalog differs")
    # The typed catalog JSON, not the pretty source file, defines Catalog.Hash.
    require("sha256:"+sha(compact(config["catalog"])) == report["catalogHash"], "catalog hash differs")
    require(manifest["versions"]["model_catalog_hash"] == report["catalogHash"], "manifest catalog differs")
    model = "jev-1.14.0" if route == "jev" else "gpt-6-luna"
    url = "https://api.typesafe.ai/v1/systemone" if route == "jev" else "https://api.openai.com/v1/decisions"
    proof = decode(read("MM11-configuration/authorization.json"))
    require(proof == report["authorization"] and sha(compact(proof)) == report["authorizationHash"], "authorization differs")
    require(proof["model"] == model and proof["service"] == url and proof["route"] == route and
            proof["dedicatedCredential"] is True and proof["maxCalls"] == proof["maxAttempts"] == 12,
            "authorization mismatch")
    b = report["budget"]
    require(b["limits"] == LIMITS and b["retries"] == 0 and b["processSeconds"] == 180 and
            b["callSeconds"] == 30 and b["finalizationReserveSeconds"] == 10 and
            b["maxNativeBodyBytes"] == 4 << 20, "wrong budget")
    start, finish = [datetime.fromisoformat(report[key].replace("Z", "+00:00"))
                     for key in ("startedAt", "finishedAt")]
    require(0 <= (finish-start).total_seconds() <= 180, "process time budget")
    totals = dict.fromkeys(LIMITS, 0)
    reserved = dict.fromkeys(LIMITS, 0)
    tokens_in = tokens_out = 0
    request_ids = set()
    for scenario in expected(route):
        prefix = "MM11-"+scenario+"/"
        req = decode(read(prefix+"public-request.json"))
        reservation = decode(read(prefix+"reservation.json"))
        for key in LIMITS:
            reserved[key] += reservation[key]
        exchanges = decode(read(prefix+"exchanges.json"))
        result_doc = decode(read(prefix+"result.json"))
        res = result_doc["result"]
        obs = decode(read(prefix+"observations.json"))
        stats = decode(read(prefix+"observer-stats.json"))
        require(stats["failed"] == stats["dropped"] == 0, "Observer delivery failed")
        require(0 <= decode(read(prefix+"duration.json"))["milliseconds"] <= 30000,
                "call time budget exceeded")
        require(result_doc["error"] == "" and res["stopReason"] == "stop", "call failed")
        require(len(exchanges) == 1, "unexpected HTTP attempt")
        ex = exchanges[0]
        require(ex["url"] == url and ex["method"] == "POST" and ex["status"] == 200 and
                not ex.get("truncated") and len(ex["requestBody"].encode()) <= 4 << 20, "wrong final HTTP")
        native, response = decode(ex["requestBody"]), decode(ex["responseBody"])
        fixtures = decode(read("MM11-ordered-reuse-plain/public-request.json"))["images"]
        validate_mapping(req, native, route, scenario.endswith("hooked"), fixtures)
        meta = res["metadata"]
        require(native["model"] == model == meta["modelId"] and meta["catalogHash"] == report["catalogHash"], "wrong final model/catalog")
        require(meta["requestId"] not in request_ids, "reused RequestID")
        request_ids.add(meta["requestId"])
        require(res.get("responseModel") == response["model"], "actual response model differs")
        require(len(meta["attempts"]) == 1, "retry/absent Attempt")
        attempt = meta["attempts"][0]
        require(attempt["attemptId"] == meta["requestId"]+"#1" and not attempt.get("retryDelay") and
                attempt["httpStatus"] == 200 and attempt["usage"] == res["usage"], "Attempt differs")
        require(len(obs) == 4 and obs[-1]["call"] == meta and obs[-1]["usage"] == res["usage"], "Observer differs")
        require(res["usage"]["input"] == response["usage"]["input_tokens"] and
                res["usage"]["output"] == response["usage"]["output_tokens"], "usage differs")
        tokens_in += res["usage"]["input"]
        tokens_out += res["usage"]["output"]
        totals["calls"] += 1
        totals["attempts"] += 1
        totals["questions"] += len(req["questions"])
        images, seen = [], set()

        def image(key, mime, data):
            raw = base64.b64decode(data, validate=True)
            identity = dict(mimeType=mime, sha256=sha(raw))
            totals["references"] += 1
            totals["expandedBytes"] += len(raw)
            if key not in seen:
                totals["sources"] += 1
                totals["sourceBytes"] += len(raw)
            if route == "decisions" or key not in seen:
                images.append(identity)
            seen.add(key)

        if route == "decisions":
            blocks = native["input"][0]["content"]
            require(native["input"][0]["role"] == "user" and blocks[0]["type"] == "input_text", "input order")
            totals["stateBytes"] += len(blocks[0]["text"].encode())
            for block in blocks[1:]:
                head, data = block["image_url"].split(",", 1)
                image(block["image_url"], head[5:-7], data)
        else:
            assets = native["assets"]

            def walk(value):
                if isinstance(value, str) and value in assets:
                    asset = assets[value]
                    mime = "image/"+asset["format"].replace("jpg", "jpeg")
                    image(value, mime, asset["data_base64"])
                elif isinstance(value, dict):
                    for key in sorted(value):
                        walk(value[key])
                elif isinstance(value, list):
                    for part in value:
                        walk(part)

            walk(native["state"])
            totals["stateBytes"] += len(compact(native["state"]).replace(b"<", b"\\u003c").replace(b">", b"\\u003e").replace(b"&", b"\\u0026"))
            for key in sorted(native["questions"]):
                q = native["questions"][key]
                walk(q.get("instructions"))
                walk(q.get("criteria"))
            require(ex["requestHeaders"]["typesafe-danger-zone-images"] == "true", "image opt-in absent")
            if scenario.startswith("box-existence"):
                require(res["answers"]["location"] == response["answers"]["location"], "Box changed")
                require("exists" in res["answers"], "independent existence absent")
            if scenario.startswith("visual-criteria"):
                answer = res["answers"]["score"]
                require(answer["legend"] == {"0": {"example": None}, "1": {"example": None}},
                        "public visual legend changed")
                refs = []
                for i, level in enumerate(native["questions"]["score"]["criteria"]):
                    asset = assets[level["example"]]
                    digest = sha(base64.b64decode(asset["data_base64"], validate=True))
                    index = next(n for n, identity in enumerate(images) if identity["sha256"] == digest)
                    refs.append(dict(imageIndex=index, placement="value", pointer=f"/{i}/example"))
                    require(response["answers"]["score"]["legend"][str(i)] == level,
                            "native legend differs from final criteria")
                require(answer["legendImageReferences"] == refs, "legend final image index differs")
        require(images == res["imageInputs"], "final image identities differ")
    require(totals == b["actual"] and reserved == b["reserved"], f"budget counters differ: actual={totals}, reserved={reserved}")
    require(b["callsUsed"] == totals["calls"] and b["httpAttempts"] == totals["attempts"] and
            b["reportedInputTokens"] == tokens_in and b["reportedOutputTokens"] == tokens_out, "usage counters differ")
    for key, limit in LIMITS.items():
        require(0 <= totals[key] <= reserved[key] <= limit, "cumulative budget exceeded")
    require(report["sourceHashes"] and report["unknown"], "source seal/unknowns absent")
    print(f"{route}: {report['mode']} candidate PASS; {len(files)} hashes verified")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--require-live", action="store_true")
    parser.add_argument("bundle", type=Path)
    args = parser.parse_args()
    try:
        verify(args.bundle, args.require_live)
    except (ValueError, KeyError, TypeError, OSError, IndexError) as error:
        print(f"REJECTED: {error}")
        raise SystemExit(1)
