# Layout deduplication — first structural pass

Algorithm: `native-layout-dedup-4`. Generated from the four hash-verified registered sources.

| Measure | Count |
| --- | ---: |
| source slides | 369 |
| native content groups | 369 |
| duplicate native content occurrences | 0 |
| multi member native groups | 0 |
| geometry candidate buckets | 336 |
| near candidate pairs before queue limit | 0 |
| review queue pairs | 33 |
| unique semantic layout count | Pending review |

Native-content groups use normalized slide XML and a conservative resource closure,
including inherited layout/master/theme bytes and presentation defaults. They are
not pixel comparisons. The geometry queue abstracts literal text and image identity;
every proposed layout merge remains pending style, role and visual review.

## Native-content duplicate groups

No exact native-content groups found by this conservative algorithm.

## First candidate pairs

| Left | Right | Geometry score | Basis |
| --- | --- | ---: | --- |
| enablecomp:003 | enablecomp:008 | 1.0000 | same_rounded_geometry_candidate |
| enablecomp:004 | enablecomp:025 | 1.0000 | same_rounded_geometry_candidate |
| enablecomp:005 | enablecomp:034 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:033 | graphics-and-layouts:034 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:035 | graphics-and-layouts:036 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:055 | graphics-and-layouts:056 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:074 | graphics-and-layouts:075 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:076 | graphics-and-layouts:077 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:094 | graphics-and-layouts:098 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:111 | graphics-and-layouts:112 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:117 | graphics-and-layouts:118 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:148 | graphics-and-layouts:149 | 1.0000 | same_rounded_geometry_candidate |
| graphics-and-layouts:153 | software-modernization:007 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:002 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:013 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:021 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:028 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:032 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:036 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:039 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:040 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:041 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:043 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:051 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:076 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:001 | software-modernization:083 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:042 | software-modernization:077 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:042 | software-modernization:080 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:052 | software-modernization:053 | 1.0000 | same_rounded_geometry_candidate |
| software-modernization:052 | software-modernization:054 | 1.0000 | same_rounded_geometry_candidate |

## Limitations and next gate

- Identical native layout names or masters never establish a match.
- Incidental dependency XML differences can hide real duplicates; no tolerance is silently used to confirm them.
- Nested geometry remains local; near-geometry comparison skips nested groups.
- Candidate edges do not form automatic transitive clusters.
- All source occurrences and their own content/fit responsibilities remain intact.
- Render representatives and ambiguous pairs; record explicit decisions before sharing semantic layout contracts.
- The 250–300 layout estimate remains unverified. No reduction target was imposed.

## Visual review decisions

Native PowerPoint previews were inspected for all 51 candidate occurrences.
The authoritative decisions are in [layout-decisions.json](layout-decisions.json).

| Measure | Count |
|---|---:|
| source slides | 369 |
| reviewed previews | 51 |
| shared visual families | 17 |
| reviewed distinct items | 1 |
| classification work units | 336 |
| repeated classification units avoided | 33 |
| unreviewed singletons | 318 |
| queue pairs reviewed | 33 |
| pairs kept separate | 2 |
| final unique layout count | Not established |

This reduces the current classification worklist to 336 units: 17 shared visual
families, one reviewed blank item and 318 unreviewed singletons. It does not
establish 336 unique reusable layouts. Broader matching remains necessary for
different group encodings, inherited geometry and component cardinality. The
250–300 estimate remains unverified.

The structural queue has 33 pairs. Two blank-versus-title pairs are kept separate;
77/80 are grouped after direct visual comparison. UHG 3/10/21 and 59/63 are joined
as one five-member arrangement family after root review, retaining narrower
underline and closing-subtitle variants. All 369 occurrences remain addressable.

Representative selection currently follows source order, with stock designs
favored where available. It does not declare the representative the best design.
Every family has classification-only readiness and an unrated design preference.
