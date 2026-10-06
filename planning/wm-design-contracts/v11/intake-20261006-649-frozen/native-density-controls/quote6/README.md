# Native quote controls

Six exact font/size/leading pairs from the designer's `03fad66` quote scaling
rules were captured in local Microsoft PowerPoint. This packet retains the
source-pinned control manifest, untouched input PPTX, character and paragraph
measurements, native PDF, signed CLI export receipt, and PDF text-show origins.

The maintained Go fixture `TestWriteDensityQuoteNativeControls` generates only
previously unmeasured pairs. `TestDeriveDensityNativeCapture` matches the input
hash, text, frames, assigned font settings and paragraph leading before deriving
baseline offsets, terminal visible-character heights and observed line pitch.
The permanent calibration closure test independently rederives every anchor
from these files. No nearby font size or line spacing is treated as calibrated.

The modern supplement has 312 controls: the prior 306 plus these six. Sources
predating the designer's quote/shape-ink rules retain the original 306-control
supplement and its SHA. The original 17 base anchors remain unchanged.

Native assigned font names are observed; the exact installed font-file identity
is not verified. These measurements do not qualify arbitrary slide content.
Native visual qualification is tracked separately in the gallery intake.
