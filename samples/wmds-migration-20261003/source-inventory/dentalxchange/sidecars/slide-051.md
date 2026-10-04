---
schema_version: 1
source_file: "DentalXChange_Product_RD_Operating_Model_SOW_Working_Deck_v4.pptx"
source_sha256: 4079af03947081cf446dce38025d441aa02c813c085b997b9464ce0de202d69c
slide_number: 51
slide_id: "300"
slide_xml: "ppt/slides/slide51.xml"
rendered_png: null
title_candidate: ""
---

# Slide 51: Untitled

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 13 |
| text shapes | 8 |
| pictures | 3 |
| tables | 0 |
| charts | 0 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 2 |

## Extracted slide text

```text
Environments, access, and test data

Remove the waits for access and realistic data that slow both build and validation.

ENGINEERING E3

WHAT SHOULD BE DONE

Make access and test-data wait times visible, by team and by type of request.
Provide masked, production-like test data on a refresh cadence, within HIPAA and PHI rules.
Give engineers, including offshore teams, approved read-only access for diagnostics.

HOW TO ACTIVATE IT

Now (Reconcile AI and Rules Engine): log who waits for access and test data, for what, and for how long.
Next: masked, production-like data for Recon and a read-only diagnostic access path.
Future: self-service environments and access provisioning on every pod.

Owner: DevOps/Infrastructure + Security | Dependency: data masking approach, security review, environment capacity | Measure: wait time for access and test data
```

## Speaker notes

```text
E3 sources: engineering survey free text (access delays, late test data, no read-only database access for troubleshooting), product survey (Reconcile AI waited months for production-like data), Core engineering working session (offshore access to production data), and the readout (HIPAA and PII approval constraints).
```

## Alt text

- How to activate it
- What should be done
