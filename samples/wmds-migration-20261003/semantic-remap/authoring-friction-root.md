# Observations from semantic case authoring

## Implemented reusable operations

The semantic pass used the frozen Go CLI for shared-template discovery, geometry/binding inspection, actual-parent local scaffolds, validation and build. No new Python conversion file was added. Human/agent work chose the argument and reshaped copy; local definitions retain shared ancestry and a reason.

## Observed friction

- Pull quote fit: compiler fit passed, but the native quote mark touched the card edge at padding9. Increasing padding to18 produced a clean inset. Quote marks should participate in the card's measured top reserve.
- Metric stacks: wrapped caption lines can abut the next large metric even when each node fits independently. Short captions and explicit spacing repaired the actual native pages. A future advisory measurement should report inter-node overlap/insufficient reserve as well as intra-node fit.
- Source screens: an original case contained separate device frames and screen captures. Choosing the frame assets alone produced empty devices in a valid build. Asset inspection and native review selected the actual screen captures instead. A reusable source-media inventory with preview/crop/constituent provenance would make this less error-prone.
- Local panel bounds: a text box placed outside its declared parent zone failed strict validation. The error correctly enforced the contract; moving the text into the correct zone repaired the definition.
- Panel padding: an object supplied for padding failed because the schema currently expects a numeric value. CLI schema/examples should keep this visible.
- Brief editing: new editing documentation described inline text, while the established schema requires a relative path to a page brief. Short text accidentally passed as a nonexistent path; longer text failed with a filesystem error. The correction keeps path semantics and provides clearer documentation and contextual validation errors. Delivery runtimes remain frozen so the follow-up cannot alter already-qualified artifacts.
- Native review remains necessary: build/check success did not catch caption-to-next-metric crowding or semantically empty device frames.

## Qualification

The case fragment had five actual native layout/media findings (original58/60/67/69/70) plus optional hierarchy/padding refinements. Compiler preflight rejections are recorded separately and are not counted as native defects. The integrated83 deck received a full individual native pass and two repair batches; all83 now have native acceptance. Source preservation and artifact equivalence are separate qualification checks.

The integrated review exposed12 missing headlines from one repeated unbound `role: title` mistake. These authored definitions now use `slide-title`, and the repository loader rejects unused non-frame zones with a contextual diagnostic. Custom roles remain extensible when explicitly bound. All83 Software briefs now use real project-relative Markdown paths. These source fixes preserve the frozen delivery runtime.

Reusable Go operations still needed for future migrations are recorded in [AUTHORING-GO-API-GAPS.md](../../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/repair-wave/AUTHORING-GO-API-GAPS.md): source inventory intake, transactional scaffold installation, qualification comparison and minimal delivery dependency closure. They are proposals, not implemented CLI verbs.
