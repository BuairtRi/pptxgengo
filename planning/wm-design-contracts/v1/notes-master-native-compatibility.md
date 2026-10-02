# Optional notes master list compatibility

## Schema and package evidence

The official Microsoft Open XML SDK metadata defines `p:notesMasterIdLst` as optional (zero or one occurrence) in `p:presentation`. When present, its position precedes `p:handoutMasterIdLst` and `p:sldIdLst`. The required `p:notesSz` remains present. Source: [SDK presentation schema metadata](https://github.com/dotnet/Open-XML-SDK/blob/main/data/schemas/schemas_openxmlformats_org_presentationml_2006_main.json).

Read-only comparison of `samples/wmds-library-20261002/review03/library-reference.pptx` and `isolate-full-no-notes-list/reference.pptx` confirms that the diagnostic modifies only `ppt/presentation.xml`, omitting the optional list. Both packages retain the presentation-to-notes-master relationship, all 97 notes parts, their 97 relationships to `ppt/notesMasters/notesMaster1.xml`, and identical notes text. No notes-master relationship target is missing.

The parent agent's native isolation established that the original full package triggers repair, while the package with only this omission opens without repair and displays all 97 slides with notes. The same behavior was isolated on a plain-title, empty-body single slide, excluding chart, table and custom geometry content as the trigger for this repair. The shared writer now omits the optional list and preserves notes parts and relationships.

## Acceptance boundaries

This is a narrow native compatibility resolution supported by package isolation. It does not identify the underlying PowerPoint implementation behavior, and the original list order agrees with the published schema. The notes relationship graph and content remain intact. Native acceptance applies to the observed source reference and bound fixture on the reviewed PowerPoint installation; it does not qualify every Office version or arbitrary caller content.

Native slide opening and PDF review do not establish notes-page formatting or print fidelity. The workaround does not remove notes, rename fonts, change text metrics, or flatten editable content. No source snapshot or calibration data changes are part of this resolution.
