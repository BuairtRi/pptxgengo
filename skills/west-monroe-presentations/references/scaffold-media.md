# Scaffold with explicit schematic media

Use normal `project scaffold` when the project has its required verified artwork.
For a private, synthetic layout prototype, explicitly request schematic media:

```sh
pptxgengo design project scaffold --bundle v11 \
  --template team-curve/build-together \
  --reason "Prototype the composition using clearly schematic media" \
  --placeholder-media --out scaffold.json
```

Match the project's toolchain lock. This mode requires v4.3.0 or later. It is a
local derivative; `--stock` retains the original shared template and cannot use
this flag. Review the output's `policy` and `placeholder_media` receipt. Genuine
registered icons and arrows remain genuine; other registered media can become
conspicuous gray schematic placeholders. This does not acquire or redistribute
branding originals or photographs.

Placeholder envelopes follow the mark's purpose: title underscores remain thin
horizontal accents, and highlight brushes retain a horizontal aspect. These are
schematic shapes with distinct hashes, not replicas of the original artwork.

## Install the returned assets with the derivative

The JSON contains the adapted `template`, `synthetic_source_values`, `assets`,
`asset_payloads` and `placeholder_media`. Install the local template and values
through the project's usual local-template/slide source structure. Also:

1. Merge each returned `assets` descriptor into the deck's global asset registry.
2. Decode each base64 `asset_payloads` value to its declared project-relative
   path. The API returns raw bytes; only the JSON encoding uses base64.
3. Verify the payload SHA matches the descriptor and path. Keep its
   `placeholder_for` and schematic description; they explicitly declare the
   source media being substituted. Do not mark it as a registered original.
4. Run project check/build from that project, inspect the deck and retain the
   receipt with its decisions. Preview/edit only a copy of the generated deck.

Paths use `assets/objects/sha256/HASH`, so repeated placeholder bytes share one
file across versions. Preserve the asset registry and objects when snapshotting,
sharing or moving the project. The descriptor's explicit alias also covers inline
marks; do not replace matching words in copy or edit renderer strings manually.

## Review or replace before delivery

Schematic images are intentional missing-content markers. Replace/review them
before delivery; template previews are not branding fidelity evidence. Replacing
an alias with a real project-owned graphic requires its real content hash and the
normal asset registration/composition workflow, with the placeholder alias
removed. Retain the previous snapshot and decision receipt.

Normal scaffolding has no silent placeholder fallback. Unknown aliases, competing
aliases/original claims, original-byte hash claims, protected icon/arrow aliases,
missing hashes, unsafe paths, symlinks and payload drift are refused. Never weaken
those checks to make a missing branding path succeed.
# Original frame-container clearance

Scaffolding does not silently shrink source geometry. For a pinned original
container/frame whose height leaves insufficient source-caption/footer
clearance, explicitly use `--source-container-clearance-fit` with your adaptation
reason. Only an original envelope within six points of the body bottom is
eligible. The result retains its true parent and `_source_geometry`, reports
`geometry_adjustments`, and uses a fixed `source_container` allocation with
six-point caption/footer clearance. V11 `architecture/layer-map` changes height
294 to 291. Review this intentional geometry change before delivery.

