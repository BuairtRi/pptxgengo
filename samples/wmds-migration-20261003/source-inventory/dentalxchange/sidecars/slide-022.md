---
schema_version: 1
source_file: "DentalXChange_Product_RD_Operating_Model_SOW_Working_Deck_v4.pptx"
source_sha256: 4079af03947081cf446dce38025d441aa02c813c085b997b9464ce0de202d69c
slide_number: 22
slide_id: "318"
slide_xml: "ppt/slides/slide22.xml"
rendered_png: null
title_candidate: ""
---

# Slide 22: Untitled

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 20 |
| text shapes | 15 |
| pictures | 1 |
| tables | 0 |
| charts | 0 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
Engineering Priority: Test

Move quality checks earlier and remove the waits that slow validation.

ENGINEERING · TEST

STATUS TODAY

Core regression is manual, with limited QA capacity.
Reconcile AI's verification agents write Jest, Playwright, and Postman tests per story.
XConnect is ~100% automated; other pods are partially automated.
Teams wait for access, test data, and read-only diagnostics; Recon waited months for production-like data.
QA uses Claude to create tests and diagnose failures; access is limited to some full-time staff.

NOW   Pilot: Core, Reconcile AI, Rules Engine

Map Core's highest-risk regression workflows, using the Release root-cause categories.
Log who waits for access and test data, for what, and for how long.

NEXT   Automated quality

Automate Core's highest-risk workflows first.
Extend Recon's test agents to Eligibility AI and Pay Connect.
Masked, production-like test data and a read-only diagnostic access path.

FUTURE

Automated regression as a release entry criterion on every pod.
Self-service environments and access.

Agentic lens:  Recon agents write and trace tests; QA uses Claude to draft tests and diagnose failures, time-boxed and reviewed before every PR.

Owner: Quality leadership + DevOps/Infrastructure + Security  |  Measure: automated coverage of high-risk workflows, escaped defects, wait time

© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

22
```
