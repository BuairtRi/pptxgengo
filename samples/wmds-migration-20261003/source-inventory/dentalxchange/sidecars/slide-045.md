---
schema_version: 1
source_file: "DentalXChange_Product_RD_Operating_Model_SOW_Working_Deck_v4.pptx"
source_sha256: 4079af03947081cf446dce38025d441aa02c813c085b997b9464ce0de202d69c
slide_number: 45
slide_id: "303"
slide_xml: "ppt/slides/slide45.xml"
rendered_png: null
title_candidate: ""
---

# Slide 45: Untitled

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 17 |
| text shapes | 11 |
| pictures | 1 |
| tables | 0 |
| charts | 1 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
Release health baseline, 2024–2026

Production changes to the Core applications, as a starting baseline for Ship with proof and delivery measurement.

DELIVERABLE 01

50%

of 2026 production changes fell outside the biweekly cadence: 11 patches, 2 off-version releases, and 2 rollbacks out of 30.

5 of 11

2026 patches were credential or certificate changes; two expired Salesforce passwords blocked customer workflows.

2

rollbacks in 2026 (DEA in January, DCQ in August), plus a release item held after a failed checkout (DCQ, January).

Half of 2026 changes were unplanned: defects, rollbacks, and credential or certificate fixes. Each merits a root-cause review (E1).

Source: DXC Production Release Calendar, Jan 2024 – Sep 2026; Core applications only; 2024–25 did not label patches.
```

## Speaker notes

```text
Release health baseline. Counts come from the consolidated release calendar: 2024–2025 from 'New builds' release emails, 2026 from the Deployments Team channel. The calendar lists Core applications (CHS, DCI, DCQ, CUSTCARE, REG, CRED, PNT, DWS, DEA and others); new-platform releases are not included. 2024–2025 emails did not label patches separately, so zeros for those years reflect recording, not absence. Credential and certificate patches: Salesforce client password expirations on 05/30 and 08/28 (the second followed by the 09/01 restoration of outbound Salesforce calls); BMA (04/07), DIL (05/11), and SSO (08/20) certificates applied as patches, where the calendar does not say whether they had expired. Defects that reached production: an ambiguous column name in DCI (03/23), a slow Custcare credentialing search (05/15), and a file-delivery failure for MET (07/15). Rollbacks: DEA (01/23) and DCQ (08/28). Two March patches (03/26 and 03/27) have no recorded cause.
```
