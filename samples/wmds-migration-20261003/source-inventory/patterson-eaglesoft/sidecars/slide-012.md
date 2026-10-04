---
schema_version: 1
source_file: "Patterson - Eaglesoft Modernization RFP Deck.pptx"
source_sha256: a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3
slide_number: 12
slide_id: "2147483647"
slide_xml: "ppt/slides/slide12.xml"
rendered_png: null
title_candidate: ""
---

# Slide 12: Untitled

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 44 |
| text shapes | 27 |
| pictures | 0 |
| tables | 1 |
| charts | 0 |
| groups | 0 |
| hyperlinks | 0 |
| alt text entries | 0 |

## Extracted slide text

```text
MODERNIZATION METHOD

Journey based PowerBuilder migration and retirement

Modernize a complete staff or customer journey, then retire the PowerBuilder components that served it.

JOURNEY MIGRATION PATH

1  Map the journey

Document the customer or staff outcome and its current execution path.

2  Expose the dependency chain

Link screens, DataWindows, event logic, SQL, stored procedures, integrations, devices, and existing test coverage to the journey.

3  Choose the experience

Make the product decision first. Preserve an experience that still works. Improve workflows that need material change. Sequence the journey later when a dependency blocks a responsible move.

4  Build the modern journey

Build the chosen experience, not just new interface technology. React uses shared business operations for authorization, transaction behavior, and data access.

5  Prove and retire

Integrate migration with automated behavior and performance tests. Use high-confidence evidence to confirm supportability, then remove the PowerBuilder dependencies the journey replaces.

Acceptance follows the journey: correct business outcome, dependable performance, and a supportable release.

1

2

3

4

5

✓

OPEN PRODUCT QUESTION

Should this journey preserve today’s experience, materially improve it, or move later?

RETIREMENT RULE

PowerBuilder exit occurs journey by journey, with no unowned legacy dependency left behind.

© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

12
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
12
```
