# Wave 2 synthetic bio sidebar polish

The people-spec generator now gives every bullet in the upper work-area panel and lower responsibilities panel the same 23 pt minimum block height. This evens the nominal bullet pitch while leaving native measured text height free to grow if a line wraps. The lower panel now has five illustrative proposed-responsibility items, closer to UHG67’s five-item sidebar density. The added items describe planned facilitation and documentation tasks; they do not claim completed work, credentials, or outcomes.

The UHG44 source-control page and UHG67 crop/reference geometry are unchanged. The synthetic bio keeps the existing 9.5 pt Arial marker/text blocks; rich or emphasized bullets are not introduced because the current layout API rejects rich paragraphs paired with markers.

Expected native QA: confirm both 176 pt sidebar cells fit after the 12 pt top / 10 pt bottom padding. The current declared minimum stack is 16 pt heading + five 23 pt bullet blocks = 131 pt, leaving 23 pt nominal room within each 154 pt inner cell. Actual measured blocks may use more height when text wraps. Confirm the bullet baselines appear evenly spaced and that neither panel clips; do not reduce type size if a measured block exceeds its cell.
