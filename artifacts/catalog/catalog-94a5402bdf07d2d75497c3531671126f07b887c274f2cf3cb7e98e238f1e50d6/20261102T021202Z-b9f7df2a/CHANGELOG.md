# GPT-6.1 Terra catalog delivery

Catalog.Version: catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6.

The approved directory contains 119 identities (113 chat, 3 image, 3 classifier).
Only chat/openai/openai-responses/gpt-6.1-terra is added; the prior 118 definitions
and their relative order are unchanged. OpenAI content version is 2.

Runtime Go code and dependencies are unchanged between 37a283ee4ef743f93d3a12132d1bc2c3211fb1e5 and 4831e3544945f667d9a817be8dfb92dd08856575.
The only non-test Go change is the ChineseEvaluation example default config
selection from v3 to v4. Its previous configs and snapshots remain immutable.

External JSON: the exact snapshot passed code compatibility on 37a283ee4ef743f93d3a12132d1bc2c3211fb1e5.
A Host with external load/injection can distribute JSON and create a new Client;
this does not establish compatibility with arbitrary older releases.
Builtin default: rebuild and distribute the program to replace embedded data.

28 required native cases passed; full ordinary verification, generate, build and
vet exited 0. All four bundle audits have zero findings. Live/race/pressure are
not applicable for this scope. No vendor qualification or software release is claimed.

This batch applies the Catalog CLI delivery SOP: versioned immutable runtime JSON,
exact review inventory, changelog, code/distribution decisions and verification
references are collected together. Original capsules remain at the locations
listed in verification/originals.json.
