---
schema_version: 1
source_file: "Patterson - Eaglesoft Modernization RFP Deck.pptx"
source_sha256: a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3
slide_number: 11
slide_id: "257"
slide_xml: "ppt/slides/slide11.xml"
rendered_png: null
title_candidate: ""
---

# Slide 11: Untitled

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 71 |
| text shapes | 25 |
| pictures | 0 |
| tables | 1 |
| charts | 0 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
TRANSITION ARCHITECTURE

React entry points during the PowerBuilder transition

West Monroe currently sees three potential paths forward for maintaining PowerBuilder screens in parallel with modernizing the user interface to be a modern web stack. We currently recommend redirecting users towards a browser based application; however, the exact approach will be validated during Phase 0.

PRACTICE NETWORK

Embedded PowerBuilder web view

Retains the Eaglesoft window. Use only when the installed PowerBuilder runtime supports the required behavior and the adoption value is clear.

CURRENT PILOT HYPOTHESIS

Browser application or PWA

Users open a practice network URL without installing or updating a desktop application. This is currently our preferred pilot for a bounded journey with limited native device needs.

Installed desktop companion

A React UI can run in Electron, Tauri, or a .NET WebView2 shell. Useful for printing and tighter device access, but it may require an added workstation install. Phase 0 will assess how much native integration is needed.

PHASE 0 VALIDATION

Work with Patterson to select one pilot route and test the assumptions that matter for that journey. Assess runtime behavior, identity, printing, imaging, device access, deployment, recovery, and user effort.

Current belief: browser or PWA offers the easiest user access.

Recommended modern web application

React + TypeScript is the recommendation

Phase 0 can assess Vue, Svelte, Angular, or Blazor/.NET if alternatives to React are considered desireable.

Eaglesoft API layer

Domain operations, authorization, and transaction control

Segregated data ecosystem

Governed PostgreSQL and related practice data

Phase 0. Work with Patterson to select one route to test. Use user effort, runtime compatibility, printing, device needs, deployment support, and recovery evidence to choose the pilot.

© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

11

Tyler – pull through carestream and prior modernization examples
```

## Speaker notes

```text
The browser application or PWA is the current pilot hypothesis because users can open a practice-network URL without installing or updating a desktop application. This reduces user effort and deployment friction.
In Phase 0, work with Patterson to select the appropriate pilot route and test that route with a real workflow. The intent is not to spike all three routes. Compare the relevant trade-offs and consider another option if the selected journey exposes a gap.
An installed desktop companion can render the same React application through Electron, Tauri, or a .NET WebView2 host. It can support printing and tighter device access, but may require users or practice IT to install and update software on each workstation. Assess how much native integration the selected journey actually needs.
React and TypeScript remain the recommendation. Phase 0 can assess Vue, Svelte, Angular, or Blazor and .NET when organizational standards or language alignment justify the trade-offs.
All modern entry routes call the Eaglesoft API layer. The API layer uses the segregated, governed practice data ecosystem. All deployment remains on premises.
Patterson public Eaglesoft product information: https://www.pattersondental.com/cp/software/dental-practice-management-software/eaglesoft-
Appeon PowerBuilder runtime notes: https://docs.appeon.com/pb/release_bulletin_for_pb/Build-2797-MR-UpgradeNotes.html
Appeon WebBrowser packaging reference: https://docs.appeon.com/pb2025/pbug/Packaging_WebBrowser.html
Internal RFP understanding: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Deliverables/2026-09-23 UTC - Patterson Eaglesoft Our Understanding.md
Internal technical narrative: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Working/2026-09-23 UTC - Patterson Eaglesoft Technical Narrative Revision 2.md
Historical diligence, validation context only: /Users/rscott/Documents/clients/private equity/patient square capital/patterson companies/Context/2026-09-23 UTC - Patterson Eaglesoft Prior Deliverables Evidence Store.md
11
```
