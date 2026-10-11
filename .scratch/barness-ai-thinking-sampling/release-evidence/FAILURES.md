# Portable evidence checker: failure cases before implementation

This checker reads archived public E2E output. It never calls a vendor.

- A missing or changed archive, report, command record, or source hash must fail verification.
- An archive path escaping the temporary directory, a symlink/hardlink, duplicate name, or excessive expanded size must be rejected before extraction.
- PAX/global-PAX/GNU extension payloads and consecutive extension headers must be bounded before tarfile parses them; metadata must not bypass payload limits.
- Every extracted archive tree must receive a fresh redaction audit, including interrupted runs with no manifest; a synthetic verification manifest must not certify the original run.
- A missing case, missing artifact, invalid relative artifact name, or altered artifact must fail verification.
- An audit finding must fail verification even when the recorded test passed; a failed test is allowed only as a recorded failure.
- Different current sources or source inventory must prevent recompiling gate/audit tools against the recorded evidence.
- A different catalog, oracle, traceability or ledger must not silently reinterpret the recorded result.
- Nine live combinations and the independent Chinese evaluation must remain distinct and complete; a historical catalog report must not replace a current one.
- Recorded failing live/effect gates must remain FAIL on artifact replay. Command gates are NOT_RUN on replay and cannot claim fresh execution.
- Chinese evaluation `-verify` exits zero for a complete, honest FAIL report: integrity success must not be confused with business-result PASS.
- An interrupted initial gate must retain its failed offline bundle and log; its incomplete race output cannot certify a complete race run.
- A completed race case suite whose original artifact audit was killed must remain a command FAIL, even if a later independent audit succeeds.
- Python optimization must not disable integrity assertions.
- A live runner's output outside the repository or reuse of an existing output directory must fail before credentials are read or calls begin.

Boundary checks are executable in verify.py: --fault-tests deliberately supplies
unsafe archives and changed/missing files, then runs the normal complete replay.
The final verification receipt records actual return codes and evidence decisions.
