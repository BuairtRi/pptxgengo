---
schema_version: 1
source_file: "Patterson - Eaglesoft Modernization RFP Deck.pptx"
source_sha256: a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3
slide_number: 9
slide_id: "266"
slide_xml: "ppt/slides/slide9.xml"
rendered_png: null
title_candidate: "Recommended target architecture for Eaglesoft user journeys"
---

# Slide 9: Recommended target architecture for Eaglesoft user journeys

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 30 |
| text shapes | 19 |
| pictures | 2 |
| tables | 1 |
| charts | 0 |
| groups | 2 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

9

Recommended target architecture for Eaglesoft user journeys

CORE RECOMMENDATION
Retire PowerBuilder journey by journey
Identify end-to-end user journeys and define acceptance criteria
Retire each screen and runtime only after its journey is accepted
Rebuild UI logic in React + TypeScript
Route business rules through API and domain services
Centralize PostgreSQL access and preserve on-premises documents and imaging

TARGET ARCHITECTURE

DATA ACCESS LAYER
PostgreSQL access

API LAYER
Domain services

USER INTERFACE LAYER
React + TypeScript

ON-PREMISES DATA
Documents + imaging

PARTNER INTEGRATIONS
Continuous support
API + data access paths

MODERNIZATION PRINCIPLES

JOURNEY-BY-JOURNEY

Modernize one complete user journey at a time; retire its PowerBuilder screen only after cutover.

TRANSITION PLAN

Sequence coexistence, testing, deployment, and support so legacy and modern experiences can run safely in parallel.

PARTNER CONTINUITY

Continuously support approved partner integrations throughout migration, whether they use APIs or data access.

KEY DESIGN DECISION

For each journey, decide whether to reimagine the experience or migrate it largely as is.
```

## Speaker notes

```text
Our Differentiators:
- Easier technology due diligence in future transactions
- Improved SOX readiness through:
Clear data lineage
Modern controls
Documented architecture
- Platform extensibility:
Multi-entity support
Future acquisitions
- Reduced dependency on tribal knowledge
9
```
