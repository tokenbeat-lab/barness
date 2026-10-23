# Retained GLM captures

These six ZIPs preserve the exact raw/support-view bytes already pinned by
../live-captures.json and ../support-views.json. Their SHA-256 filenames separate
immutable content identity from capture-time local .evidence locators. Native
index bytes, pins, report hashes, measured sources and original derivations stay
unchanged. The public importer reads these tracked archives directly, with no
host-directory fallback or runtime restoration.

Historical views still reject current source drift. Shipping the inputs does not
renew live qualification, update a support matrix or authorize a Provider call.
provenance.json records the original locations, sizes, report paths and hashes.
