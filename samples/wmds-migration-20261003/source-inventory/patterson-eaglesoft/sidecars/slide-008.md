---
schema_version: 1
source_file: "Patterson - Eaglesoft Modernization RFP Deck.pptx"
source_sha256: a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3
slide_number: 8
slide_id: "2147483646"
slide_xml: "ppt/slides/slide8.xml"
rendered_png: null
title_candidate: "Recommended target architecture for Eaglesoft user journeys"
---

# Slide 8: Recommended target architecture for Eaglesoft user journeys

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 66 |
| text shapes | 22 |
| pictures | 0 |
| tables | 1 |
| charts | 0 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

8

TARGET ARCHITECTURE

Recommended target architecture for Eaglesoft user journeys

A user journey is the end-to-end work a staff member or customer completes to achieve an outcome. Modernize complete journeys through a three-tier architecture that separates the user interface, API and domain services, and data access.

PRACTICE NETWORK

PowerBuilder during transition

Only for user journeys that have not moved

Complete journey moves

1  USER INTERFACE LAYER

Modern Eaglesoft application

React and TypeScript
Complete staff and customer journeys

2  API LAYER

3  DATA ACCESS LAYER

Centralized, governed PostgreSQL interaction

ON-PREMISES DATA AND INTEGRATIONS

PostgreSQL, documents, imaging, and approved practice integrations

CORE RECOMMENDATION
Identify end-to-end user journeys for PowerBuilder retirement. After a journey meets acceptance criteria, retire its screen and runtime, rebuild its UI logic in React, and route business rules and data access through the API and data access layers.

Initial phase: define journey boundaries and validate PostgreSQL readiness, practice server capacity, device and integration dependencies, and deployment support before selecting the first journey.

Domain services
Business rules, authorization, and transaction control

Shared engineering controls

Automated testing, performance evidence, telemetry, and release controls
```

## Speaker notes

```text
Architecture shown is a proposal recommendation. All deployment remains on premises. Confirm application boundaries, PostgreSQL readiness, practice server capacity, device integration, and deployment support during the initial phase.
Patterson public Eaglesoft product information: https://www.pattersondental.com/cp/software/dental-practice-management-software/eaglesoft-
Internal RFP understanding: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Deliverables/2026-09-23 UTC - Patterson Eaglesoft Our Understanding.md
Internal technical narrative: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Working/2026-09-23 UTC - Patterson Eaglesoft Technical Narrative Revision 2.md
Historical diligence, validation context only: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Context/2026-09-23 UTC - Patterson Eaglesoft Prior Deliverables Evidence Store.md
8
```
