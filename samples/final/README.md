# Final converted decks

| Deck | PowerPoint | Editable YAML | Visual reference | Offline package |
| --- | --- | --- | --- | --- |
| Software Modernization · 83 slides | [deck.pptx](software-modernization/deck.pptx) | [deck.yaml](software-modernization/deck.yaml) | [PDF](software-modernization/visual-reference.pdf) | [ZIP](software-modernization/offline.zip) |
| Patterson Eaglesoft · 39 slides | [deck.pptx](patterson-eaglesoft/deck.pptx) | [deck.yaml](patterson-eaglesoft/deck.yaml) | [PDF](patterson-eaglesoft/visual-reference.pdf) | [ZIP](patterson-eaglesoft/offline.zip) |
| DentalXChange · 75 slides | [deck.pptx](dentalxchange/deck.pptx) | [deck.yaml](dentalxchange/deck.yaml) | [PDF](dentalxchange/visual-reference.pdf) | [ZIP](dentalxchange/offline.zip) |

Each folder includes required source dependencies and a toolchain lock. The
offline ZIP includes the frozen compiler and library used for the qualified build.

DentalXChange now has a 125-line deck descriptor, one editable YAML file per slide,
separate Markdown notes and three local exception definitions. It uses 71 genuine
stock templates (94.7%); all 75 slides have native visual acceptance and a relocated
offline rebuild produces the same PowerPoint. See its [editing guide](dentalxchange/README.md)
and [designer gaps](dentalxchange/TEMPLATE-MAPPING.md). Software and Patterson are unchanged.

Original input decks remain in the parent `samples` folder. Template discovery
and all 616 native template previews live in
[`library/wm-design-system/v7`](../../library/wm-design-system/v7/).
