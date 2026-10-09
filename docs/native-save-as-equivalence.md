# PowerPoint Save As structural equivalence

Implementation/qualification work dated 2026-10-08. This is a narrow compatibility
rule for existing geometry reconciliation and explicitly mapped known block
copies, not an arbitrary native-object import feature.

Actual macOS PowerPoint Save As plus GUI duplication exposed three serialization
differences in an otherwise unchanged block subtree:

- Added Office 2014 `creationId` metadata beneath `cNvPr`.
- An empty text body added to a surface that previously had no text body.
- Removed paragraph-end properties identical to the single plain text run.

The analysis comparator recognizes only:

1. The exact creation-ID extension URI, namespace and element, with one valid
   braced GUID attribute and no other payload. Unknown extensions and extra
   attributes remain compared. Creation IDs never replace receipt lineage.
2. An attribute-free `txBody` with empty `bodyPr` and `lstStyle`, and one empty
   paragraph containing only a language-valued `endParaRPr`. Added text, other
   paragraph content, anchoring, sizing, or additional formatting stay visible.
3. Paragraph-end properties exactly equal to the existing single nonempty plain
   run's properties. A meaningful end style, multiple runs, fields, or breaks
   does not qualify for this normalization.

These are analysis views. Original PPTX files, exact packet inputs, raw hashes,
baseline receipts, source text, and retained predecessor bytes are preserved.
The general strict text-structure hash remains unchanged; the normalizations
apply to geometry/copy compatibility. Unsupported native style, attachment,
route, topology, ownership and payload changes remain subject to existing checks.
Copy rejection now identifies the differing XML location and category without
printing slide text or arbitrary XML values.

## Tests and qualification boundaries

Focused positive tests verify bidirectional equivalence and no input mutation.
Negative tests cover unknown/malformed metadata, additional properties, nonempty
surface text, changed fonts/sizes/colors/paint/presets, meaningful end properties,
and rich-run additions. Existing mapped-copy, inherited-tag, idempotence, bounds,
ownership and connector round-trip tests remain required.

A new development CLI changes the executable pin. Reconciliation intentionally
refuses an old GUI file against a migrated project's old receipt when the lock
differs. Qualification must generate a fresh baseline using the exact new CLI,
then save and edit that baseline in PowerPoint. Do not rewrite receipt hashes,
change executable pins to impersonate an old compiler, or relabel a controlled
XML replay as an actual GUI round trip.

Private evidence is retained under
`~/Documents/pptxgengo-qualification/gui-save-as-20261008/attempt-01/`.
This rule does not establish Windows or broad native-import qualification.

## Actual GUI copy/delete specimen

A fresh receipt-backed baseline was generated with a fixed development CLI
(SHA-256 `bbc93da76e51b1875a410149f27e3029b0f7ffeca0d9326cfafb9e0f28830ecc`).
PowerPoint GUI Save As, whole `node08` deletion, and `node06` duplication were
performed on its owned copy. Duplicate-name accessibility ambiguity was avoided
by coordinate selection and unique root naming; an exact-owned native API moved
the copy to 687,234 pt and saved it. The exact saved file was proposed, not an XML
replay or a file with rewritten receipt identities.

Proposal and adoption passed for the two complete structural changes. The packet
retains 25 manual findings: 23 package changes and two structure/format findings
for `node01`/`node01.bullets`. Those changes were not silently normalized or
claimed adopted. Actual PowerPoint PNG exports of the edited and rebuilt decks
were byte-identical, SHA-256
`18fc2dfd14571ed72c590ec177c6402b7a2711c6e09dc2d628882d66ee77f4ae`.
The rebuilt deck is `build-20261009T004825-9007f62af4e21389`; command records,
closed packet and raw inputs are in `copy-serialization-followup-01/`, with root
visual evidence in `copy-followup-visual-comparison.json`.

This qualifies this known block copy/delete workflow on macOS PowerPoint 16.113.4.
It does not qualify all package rewrites, arbitrary inserted objects, semantic
team/Gantt composition, other Office versions, or Windows behavior.
