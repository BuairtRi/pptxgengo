---
schema_version: 1
source_file: "Patterson - Eaglesoft Modernization RFP Deck.pptx"
source_sha256: a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3
slide_number: 5
slide_id: "289"
slide_xml: "ppt/slides/slide5.xml"
rendered_png: null
title_candidate: "Eaglesoft Architecture – What We Know & Understand Today"
---

# Slide 5: Eaglesoft Architecture – What We Know & Understand Today

## Structural data

| Element | Count |
| --- | ---: |
| shapes | 31 |
| text shapes | 27 |
| pictures | 4 |
| tables | 1 |
| charts | 0 |
| groups | 2 |
| hyperlinks | 0 |
| alt text entries | 1 |

## Extracted slide text

```text
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.

5

Eaglesoft Architecture – What We Know & Understand Today

Our Understanding

Dental Office

Data Layer

App Layer

Patterson App Server
(.NET C#)

Server Status
(.NET C#)

Eaglesoft DB (SQL Anywhere)

Client (PowerBuilder)

DB Upgrade
(.NET C#)

Fast Check-In
(.NET )

Clinical
(PowerBuilder)

Eaglesoft API

External Interfaces
Generic Imaging Devices (Twain)
Digital Radiography Devices (Direct)
DICOM File
PACS

Server

Designed as a traditional client-server architecture, the application and database reside on a central server, typically located at a dental office, and interact with end-user desktops (clients). While this model works well for single-office setups, it struggles to support larger dental groups with more than 30 users
Frontend UI is built on C++ and PowerBuilder, technologies that are outdated and less commonly used in modern software development making resources in the market difficult to find to maintain and update
Backend business logic is written in C++, but modernization efforts have introduced .NET 4.6, with plans to upgrade to .NET 4.8 in version 24.20. but many areas of the application still rely on outdated 32-bit libraries (eg. LeadTools, TXTextControl, and CodeJock) posing challenges in transitioning to modern 64-bit systems
Application database is SQL Anywhere, a lightweight and self-contained database engine; however, support for this database is ending in January 2025, prompting an extensive migration to a new database
Business logic is embedded in the data layer, comprising the majority of the application's functionality across 1,500 stored procedures. This design limits scalability and makes the system unsuitable as the foundation for a web-based, multi-tenant application capable of supporting thousands of customers

1

2

ABOVE  PAR

MATURITY LEGEND

BELOW PAR/GAP

AT
PAR

3

4

5
```

## Alt text

- Computer with solid fill
