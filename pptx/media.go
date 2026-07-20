// media.go ports src/gen-media.ts: reading/encoding slide media relationships
// (images/audio/video) into base64, plus image dimension sniffing.
//
// Only the Node.js code paths of the TS source are ported (per PORTING.md);
// the browser FileReader/XMLHttpRequest/<canvas> paths are skipped since Go
// has no browser runtime.
//
// Known deviations from the TS Node path (deliberate; each is also called
// out at its point of use below):
//   - mediaHTTPTimeout: TS's https.get has no timeout at all; Go bounds the
//     fetch at 30s so a library call can never hang forever.
//   - fetchMediaHTTP treats any non-2xx HTTP response as a failure. TS's
//     https.get only rejects on transport-level errors and would happily
//     base64-encode a 404/500 error page as if it were valid image bytes.
//   - Redirects: Go's http.Client (the zero-value CheckRedirect policy)
//     follows HTTP 3xx redirects automatically and embeds the FINAL
//     response's body — i.e. the actual image at the redirect target. TS's
//     https.get does NOT follow redirects on its own; given a 3xx response
//     it has no manual redirect-following logic, so real PptxGenJS embeds
//     the redirect response's own (typically tiny, non-image) body as-is.
//     Go's behavior is treated as an intentional improvement here, not
//     bug-for-bug fidelity with the TS source.
package pptx

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// mediaHTTPTimeout bounds the http(s).Get used to fetch remote media. The TS
// Node path (https.get) has no timeout at all; we add a "sensible" one since
// a library call should never be able to hang forever. This is a deliberate,
// documented deviation from the TS source.
const mediaHTTPTimeout = 30 * time.Second

// maxMediaBytes prevents local devices, FIFOs, oversized files, and remote
// responses from turning a presentation build into an unbounded read. One
// hundred MiB is intentionally generous for slide media while still bounding
// memory use.
const maxMediaBytes int64 = 100 << 20

// mediaHTTPClient is overridable by tests; production code uses the default.
var mediaHTTPClient = &http.Client{Timeout: mediaHTTPTimeout}

// ---------------------------------------------------------------------------
// resolveSlideMediaRels (ports encodeSlideMediaRels's Node code path)
// ---------------------------------------------------------------------------

// resolveSlideMediaRels reads/encodes every media relationship on layout
// (local files via os.ReadFile-backed I/O, remote http(s) URLs via
// mediaHTTPClient) and fills each SlideRelMedia.Data with a base64 string,
// mirroring encodeSlideMediaRels in gen-media.ts.
//
// It takes *SlideBaseProps (rather than *PresSlide) because the TS function
// accepts `PresSlide | SlideLayout`, and SlideBaseProps is the struct both Go
// types embed (and where RelsMedia actually lives) — this lets callers pass
// either &presSlide.SlideBaseProps or &slideLayout.SlideBaseProps.
//
// Unlike the TS original (which fires every read as an independent Promise),
// this runs strictly sequentially: PresSlide/SlideLayout are shared mutable
// state, so a caller wanting parallelism should fan this out itself (e.g. one
// goroutine per slide, not per relationship) once results are needed.
//
// On a read/fetch failure the offending rel's Data is set to IMG_BROKEN
// (matching the TS fallback) and the failure is recorded in the returned
// error slice; processing continues for the remaining rels.
//
// SVG note: see the long comment on svgPngNodeFallback below — the observed
// Node behavior for the isSvgPng "PNG preview" companion rel is subtler than
// a simple "always broken", and this function reproduces that exact quirk.
func resolveSlideMediaRels(layout *SlideBaseProps) []error {
	var errs []error
	rels := layout.RelsMedia

	// Snapshot which isSvgPng rels already carry non-empty data *before* we
	// touch anything. See svgPngNodeFallback for why this matters.
	preEncodedSvgPng := svgPngNodeFallback(rels)

	// STEP A (candidateRels): media that needs reading — not "online" (those
	// already carry data:'dummy' set at push time), not already having data,
	// and whose path (if any) doesn't contain "preencoded".
	var candidates []int
	for i := range rels {
		r := &rels[i]
		if r.Type == "online" {
			continue
		}
		if dataString(r.Data) != "" {
			continue
		}
		if r.Path != "" && strings.Contains(r.Path, "preencoded") {
			continue
		}
		candidates = append(candidates, i)
	}

	// STEP B (dedupe): mark repeat `path` values as duplicates, same as the
	// TS `unqPaths` scan (first occurrence wins, order-preserving).
	seenPaths := make(map[string]bool, len(candidates))
	for _, ci := range candidates {
		p := rels[ci].Path
		if !seenPaths[p] {
			seenPaths[p] = true
			rels[ci].IsDuplicate = ptr(false)
		} else {
			rels[ci].IsDuplicate = ptr(true)
		}
	}

	// STEP C: read/encode each unique candidate, then copy its result onto
	// every duplicate sharing the same path (mirrors the TS `.forEach(dupe
	// => dupe.data = rel.data)` propagation).
	for _, ci := range candidates {
		if rels[ci].IsDuplicate != nil && *rels[ci].IsDuplicate {
			continue
		}
		r := &rels[ci]
		data, err := readMediaSource(r.Path)
		if err != nil {
			r.Data = IMG_BROKEN
			errs = append(errs, err)
		} else {
			r.Data = data
		}
		for _, di := range candidates {
			if di == ci {
				continue
			}
			if rels[di].IsDuplicate != nil && *rels[di].IsDuplicate && rels[di].Path == r.Path {
				rels[di].Data = r.Data
			}
		}
	}

	// STEP D: Node "SVG not supported" fallback — see svgPngNodeFallback.
	for _, i := range preEncodedSvgPng {
		rels[i].Data = IMG_BROKEN
	}

	return errs
}

// svgPngNodeFallback returns the indices of rels that the real Node
// implementation's STEP-5 filter (`rel.isSvgPng && rel.data`) would catch.
//
// In the TS source that filter is a *synchronous* Array.filter call sitting
// in the body of encodeSlideMediaRels, evaluated immediately when the
// function is invoked — i.e. before any of the STEP 1-4 `async () => {...}`
// file-read closures have had a chance to run (they haven't even reached
// their first `await` yet when the synchronous part of the function
// returns). So the filter only ever sees whatever `rel.data` was *before*
// resolution:
//   - Caller-supplied pre-encoded SVG data (`data: 'image/svg+xml;base64,...'`):
//     already truthy at call time → caught → set to IMG_BROKEN ("SVG is not
//     supported in Node").
//   - Path-based SVGs (opt.path, no opt.data): both the isSvgPng "PNG
//     preview" rel and the real SVG rel start with `data: ”` (falsy) →
//     NOT caught by the filter. They proceed through the normal STEP 1-4
//     read, and because addImageDefinition gives them the *same* `path`
//     (the source .svg file), STEP 4's dedupe-copy logic ends up copying the
//     base64 of the raw SVG file bytes onto the PNG-labeled companion rel
//     too. The IMG_BROKEN fallback never applies to it. This is a latent
//     quirk/bug in the upstream JS library, not a deliberate design — we
//     reproduce it here for output fidelity rather than "fixing" it.
func svgPngNodeFallback(rels []SlideRelMedia) []int {
	var out []int
	for i := range rels {
		if boolDeref(rels[i].IsSvgPng) && dataString(rels[i].Data) != "" {
			out = append(out, i)
		}
	}
	return out
}

// readMediaSource reads and base64-encodes media from a local file path or an
// http(s) URL, mirroring the two Node branches in encodeSlideMediaRels
// (fs.readFileSync / https.get).
func readMediaSource(path string) (string, error) {
	if strings.HasPrefix(path, "http") {
		return fetchMediaHTTP(path)
	}
	return readMediaFile(path)
}

func readMediaFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("unable to read media: %q: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to inspect media: %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("unable to read media: %q is not a regular file", path)
	}
	if info.Size() > maxMediaBytes {
		return "", fmt.Errorf("unable to read media: %q exceeds %d bytes", path, maxMediaBytes)
	}
	b, err := io.ReadAll(io.LimitReader(f, maxMediaBytes+1))
	if err != nil {
		return "", fmt.Errorf("unable to read media: %q: %w", path, err)
	}
	if int64(len(b)) > maxMediaBytes {
		return "", fmt.Errorf("unable to read media: %q exceeds %d bytes", path, maxMediaBytes)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// fetchMediaHTTP performs the http(s) fetch. Unlike TS's https.get (which
// only treats *transport-level* failures as errors and would happily
// base64-encode a 404 error page as if it were valid image bytes), a non-2xx
// response here is treated as a failure too — a deliberate, documented
// deviation: bug-for-bug fidelity on "silently embed the 404 page as image
// bytes" is not useful and isn't covered by any golden-file fixture.
func fetchMediaHTTP(url string) (string, error) {
	resp, err := mediaHTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("unable to load image (https.get): %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unable to load image (https.get): %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMediaBytes+1))
	if err != nil {
		return "", fmt.Errorf("unable to load image (https.get): %s: %w", url, err)
	}
	if int64(len(body)) > maxMediaBytes {
		return "", fmt.Errorf("unable to load image (https.get): %s exceeds %d bytes", url, maxMediaBytes)
	}
	return base64.StdEncoding.EncodeToString(body), nil
}

// dataString reads a SlideRelMedia.Data value (declared `any` in types.go to
// allow string|[]byte per ISlideRelMedia) as a string. resolveSlideMediaRels
// always writes strings (base64), matching the JS `Buffer...toString('base64')`
// output, but reads defensively in case a caller pre-populated []byte.
func dataString(d any) string {
	switch v := d.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

// boolDeref safely reads a *bool, treating nil as false.
func boolDeref(b *bool) bool {
	return b != nil && *b
}

// ---------------------------------------------------------------------------
// getSizeFromImage (net-new: ports the *intent* of gen-media.ts's dead code)
// ---------------------------------------------------------------------------
//
// NOTE ON TS FIDELITY: `getSizeFromImage` in gen-media.ts (lines ~199-236) is
// entirely commented out and unreachable — it's dead code that delegated to
// the Node `sizeof` npm package (browser fallback used the DOM `Image`
// element). Its only caller, addImageDefinition in gen-objects.ts, has the
// invocation commented out too (see gen-objects.ts:445-446: "FIXME: Measure
// actual image when no intWidth/intHeight params passed... this is an async
// process"). Real PptxGenJS output therefore NEVER measures actual image
// pixel dimensions; when w/h are omitted, gen-objects.ts defaults them to `1`
// (inch). There is no TS regex or byte-parsing logic to port faithfully.
//
// This is implemented anyway per this task's spec, as a net-new, stdlib-only
// Go utility other owners (e.g. a future gen-objects.go) may opt into for
// real EMU auto-sizing. It is currently unused by any other file in this
// worktree.

// ImageSize holds pixel dimensions sniffed from raster/vector image data.
type ImageSize struct {
	Width  int
	Height int
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}

// getSizeFromImage sniffs the pixel width/height of image data by inspecting
// its header/magic bytes. Supports PNG, JPEG, GIF, BMP, and SVG (via
// width/height or viewBox attributes on the root <svg> element).
func getSizeFromImage(data []byte) (ImageSize, error) {
	switch {
	case bytes.HasPrefix(data, pngSignature):
		return parsePNGSize(data)
	case len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8:
		return parseJPEGSize(data)
	case len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a"):
		return parseGIFSize(data)
	case len(data) >= 2 && data[0] == 'B' && data[1] == 'M':
		return parseBMPSize(data)
	case svgTagRe.Match(data):
		return parseSVGSize(data)
	default:
		return ImageSize{}, errors.New("getSizeFromImage: unrecognized image format")
	}
}

// parsePNGSize reads width/height from the mandatory-first IHDR chunk:
// 8-byte signature, then [4-byte length]["IHDR"][4-byte width][4-byte height]...
func parsePNGSize(data []byte) (ImageSize, error) {
	if len(data) < 24 {
		return ImageSize{}, errors.New("getSizeFromImage: truncated PNG (< 24 bytes)")
	}
	if string(data[12:16]) != "IHDR" {
		return ImageSize{}, errors.New("getSizeFromImage: PNG missing leading IHDR chunk")
	}
	w := binary.BigEndian.Uint32(data[16:20])
	h := binary.BigEndian.Uint32(data[20:24])
	return ImageSize{Width: int(w), Height: int(h)}, nil
}

// parseJPEGSize scans marker segments for the first SOF0-SOF15 (excluding
// DHT/JPG/DAC, i.e. 0xC4/0xC8/0xCC) segment, which carries height then width
// as big-endian uint16 starting 3 bytes into the segment payload.
func parseJPEGSize(data []byte) (ImageSize, error) {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return ImageSize{}, errors.New("getSizeFromImage: not a JPEG (bad SOI)")
	}
	i := 2
	for i+1 < len(data) {
		if data[i] != 0xFF {
			i++
			continue
		}
		marker := data[i+1]
		if marker == 0xFF { // fill byte
			i++
			continue
		}
		// Markers with no payload: SOI/EOI, RSTn (0xD0-0xD7), TEM (0x01).
		if marker == 0xD8 || marker == 0xD9 || marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			i += 2
			continue
		}
		if i+4 > len(data) {
			break
		}
		segLen := int(binary.BigEndian.Uint16(data[i+2 : i+4]))
		isSOF := marker >= 0xC0 && marker <= 0xCF && marker != 0xC4 && marker != 0xC8 && marker != 0xCC
		if isSOF {
			if i+9 > len(data) {
				return ImageSize{}, errors.New("getSizeFromImage: truncated JPEG SOF segment")
			}
			h := binary.BigEndian.Uint16(data[i+5 : i+7])
			w := binary.BigEndian.Uint16(data[i+7 : i+9])
			return ImageSize{Width: int(w), Height: int(h)}, nil
		}
		if marker == 0xDA { // start of scan: no more header segments follow
			break
		}
		if segLen < 2 {
			return ImageSize{}, errors.New("getSizeFromImage: malformed JPEG segment length")
		}
		i += 2 + segLen
	}
	return ImageSize{}, errors.New("getSizeFromImage: no SOF marker found in JPEG")
}

// parseGIFSize reads the 6-byte GIF87a/GIF89a signature followed by a
// little-endian uint16 width then height (logical screen descriptor).
func parseGIFSize(data []byte) (ImageSize, error) {
	if len(data) < 10 {
		return ImageSize{}, errors.New("getSizeFromImage: truncated GIF (< 10 bytes)")
	}
	w := binary.LittleEndian.Uint16(data[6:8])
	h := binary.LittleEndian.Uint16(data[8:10])
	return ImageSize{Width: int(w), Height: int(h)}, nil
}

// parseBMPSize reads width/height from a BMP's DIB header, dispatching on
// the DIB header-size field (a little-endian uint32 at offset 14, right
// after the 14-byte BITMAPFILEHEADER):
//   - 12 (BITMAPCOREHEADER / OS/2 v1): width/height are 2-byte LE fields at
//     offsets 18/20. This header predates BITMAPINFOHEADER and uses a
//     narrower, differently-laid-out struct; blindly reading it with the
//     BITMAPINFOHEADER offsets/widths (as an earlier version of this
//     function did) silently produces garbage dimensions.
//   - >= 40 (BITMAPINFOHEADER and its supersets — BITMAPV4HEADER (108),
//     BITMAPV5HEADER (124), OS/2 BITMAPCOREHEADER2 (64), etc.): all of
//     these begin with the same first 40 bytes as BITMAPINFOHEADER, so
//     width/height are 4-byte LE int32 fields at offsets 18/22. A negative
//     height indicates a top-down bitmap; we report the absolute value
//     since callers want a magnitude, not orientation.
//   - any other value: an unrecognized/unsupported DIB header layout. We
//     error out rather than guessing, since guessing is exactly how this
//     function used to misread BITMAPCOREHEADER BMPs.
func parseBMPSize(data []byte) (ImageSize, error) {
	if len(data) < 26 {
		return ImageSize{}, errors.New("getSizeFromImage: truncated BMP (< 26 bytes)")
	}
	headerSize := binary.LittleEndian.Uint32(data[14:18])
	switch {
	case headerSize == 12:
		w := binary.LittleEndian.Uint16(data[18:20])
		h := binary.LittleEndian.Uint16(data[20:22])
		return ImageSize{Width: int(w), Height: int(h)}, nil
	case headerSize >= 40:
		w := int32(binary.LittleEndian.Uint32(data[18:22]))
		h := int32(binary.LittleEndian.Uint32(data[22:26]))
		if h < 0 {
			h = -h
		}
		if w < 0 {
			w = -w
		}
		return ImageSize{Width: int(w), Height: int(h)}, nil
	default:
		return ImageSize{}, fmt.Errorf("getSizeFromImage: unsupported BMP DIB header size (%d bytes)", headerSize)
	}
}

var (
	svgTagRe = regexp.MustCompile(`(?is)<svg\b[^>]*>`)
	// Attribute regexes support both double- and single-quoted values (both
	// are valid XML/SVG: width="120" and width='120'); exactly one of the
	// two capture groups will be non-nil on a match — see svgAttrValue.
	svgWidthAttrRe  = regexp.MustCompile(`(?i)\bwidth\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	svgHeightAttrRe = regexp.MustCompile(`(?i)\bheight\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	svgViewBoxRe    = regexp.MustCompile(`(?i)\bviewBox\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	// svgNumericRe accepts a plain number or a "px"-suffixed number; it
	// rejects percentages and other CSS units (em, cm, pt, ...), which must
	// fall back to viewBox since they aren't absolute pixel dimensions.
	svgNumericRe = regexp.MustCompile(`^[0-9]*\.?[0-9]+(px)?$`)
)

// svgAttrValue matches re against tag and returns the attribute value,
// reading whichever of the two quote-style capture groups participated in
// the match (the other is nil since regexp alternation only matches one
// branch). Returns ok=false if re did not match at all.
func svgAttrValue(tag []byte, re *regexp.Regexp) (value string, ok bool) {
	m := re.FindSubmatch(tag)
	if m == nil {
		return "", false
	}
	if m[1] != nil {
		return string(m[1]), true
	}
	if m[2] != nil {
		return string(m[2]), true
	}
	return "", true
}

// parseSVGSize reads pixel dimensions from the root <svg> element's
// width/height attributes (when both are present and given in absolute
// units), falling back to the `viewBox` width/height when width/height are
// missing or use a non-pixel unit (e.g. "100%").
func parseSVGSize(data []byte) (ImageSize, error) {
	tag := svgTagRe.Find(data)
	if tag == nil {
		return ImageSize{}, errors.New("getSizeFromImage: no <svg> root element found")
	}

	wStr, wFound := svgAttrValue(tag, svgWidthAttrRe)
	hStr, hFound := svgAttrValue(tag, svgHeightAttrRe)
	if wFound && hFound {
		wStr = strings.TrimSpace(wStr)
		hStr = strings.TrimSpace(hStr)
		if svgNumericRe.MatchString(wStr) && svgNumericRe.MatchString(hStr) {
			w, wErr := strconv.ParseFloat(strings.TrimSuffix(wStr, "px"), 64)
			h, hErr := strconv.ParseFloat(strings.TrimSuffix(hStr, "px"), 64)
			if wErr == nil && hErr == nil {
				return ImageSize{Width: int(jsRound(w)), Height: int(jsRound(h))}, nil
			}
		}
	}

	if vbStr, ok := svgAttrValue(tag, svgViewBoxRe); ok {
		// SVG's viewBox list-of-numbers syntax separates values with
		// "comma-wsp" (whitespace, a comma, or a comma surrounded by
		// whitespace) — e.g. "0,0,300,150" and "0, 0, 300, 150" are both
		// valid alongside the plain-whitespace form "0 0 300 150".
		parts := strings.FieldsFunc(vbStr, func(r rune) bool {
			return unicode.IsSpace(r) || r == ','
		})
		if len(parts) == 4 {
			w, wErr := strconv.ParseFloat(parts[2], 64)
			h, hErr := strconv.ParseFloat(parts[3], 64)
			if wErr == nil && hErr == nil {
				return ImageSize{Width: int(jsRound(w)), Height: int(jsRound(h))}, nil
			}
		}
	}

	return ImageSize{}, errors.New("getSizeFromImage: SVG has no usable pixel width/height or viewBox")
}
