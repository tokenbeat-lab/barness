# Ticket 02 — two-axis review

Fixed point: `7c1cd1a533a924dd06c57beaaa03a0977f130de2`.
Initial review command: `git diff 7c1cd1a533a924dd06c57beaaa03a0977f130de2...4d837da`.
Final implementation: `15b66fd6d9ffbbf1572ddc000890bebe3a73861b`.
The implement skill invokes code-review; Standards and Spec ran as independent parallel reviewers. Follow-up patches were reviewed on both axes. No unrelated preexisting document edits were part of the review or implementation commits.

## Standards

Initial report: “Standards: no actionable findings. The diff follows the documented boundaries, finite resource checks, immutable snapshots, precredential rejection, and E2E evidence conventions. No baseline smell warrants an actionable finding within this ticket’s scope.”

Follow-up finding, P2: “Bound raw documents before scanning projected text.” AGENTS.md principle 5 requires bounded resource use and cancellation. The new projection scan preceded the cheaper raw request/document length checks.

Resolution in `15b66fd`: reference count is bounded first; explicit declarations then check raw input and per-document sizes before projection. Cancellation is checked between bounded documents and occurrences; shorthand retains the existing raw accounting and error precedence.

Final reviewer report: “Resolved. Raw whole-input and per-document limits now precede projection scans, and cancellation is checked between bounded documents and image occurrences. No remaining actionable standards or resource-correctness findings in this patch.”

## Spec

Initial finding, P2: “Combine text and expanded-image sizes before copying or reading credentials.” R08 requires “禁止在复制 / 序列化后才检查已知超限”; R14 assigns known input limits to `resource_limit/scope`. State/questions with sources once and expanded images alone could each fit, while their known native total exceeded the request limit.

The public regression first failed with `resource_limit/request`, a credential read, and approximately 10 MB allocated. The repair reserves projected text lower bounds alongside every expanded image URL, before copying. Independent raw limits remain in force. Additional initially failing regressions ensure whitespace-padded and Unicode-escaped State can still succeed when both public and final native representations fit. These lower bounds avoid mixing raw and native representation costs.

Final reviewer report: “Resolved. The revised calculation combines conservative projected text sizes with expanded image URLs while retaining separate raw-input limits. It avoids the whitespace/Unicode-escape false rejection and closes the original pre-copy budget gap. The new public regressions cover both cases. No additional actionable issues found in this patch.”

Source budgets and reference budgets were also reviewed: native blocks are final occurrences, so applying source limits again would incorrectly reject explicitly authorized reuse beyond source capacities. Source collection bounds remain at the public boundary; final native blocks undergo reference, per-image, MIME, model, authorization and encoded-body checks.

Total findings: Standards 1 P2, resolved; Spec 1 P2, resolved. Outstanding findings: 0 on each axis.
