// xml.go ports src/gen-xml.ts: the XML string builders that turn the internal
// slide/presentation model into the OOXML parts of a .pptx package.
//
// Fidelity goal: byte-identical output to PptxGenJS. TS template-literal
// whitespace (literal tabs/newlines) is reproduced verbatim; number→string
// interpolation uses ftoa (JS String()) and jsRound (JS Math.round).
package pptx

import (
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Time hook — mirrors `new Date().toISOString()` in makeXmlCore. Overridable in
// tests so timestamps are deterministic.
// ---------------------------------------------------------------------------

// xmlNowFunc returns the current time; tests may override it.
var xmlNowFunc = time.Now

// ---------------------------------------------------------------------------
// Small helpers: itoa/boolDeref/strOr are shared with charts.go and media.go.
// ---------------------------------------------------------------------------
// Image sizing helpers (ImageSizingXml in TS)
// ---------------------------------------------------------------------------

type imgSizeDim struct{ w, h float64 }
type imgBoxDim struct{ w, h, x, y float64 }

func imageSizingCover(imgSize imgSizeDim, boxDim imgBoxDim) string {
	imgRatio := imgSize.h / imgSize.w
	boxRatio := boxDim.h / boxDim.w
	isBoxBased := boxRatio > imgRatio
	var width, height float64
	if isBoxBased {
		width = boxDim.h / imgRatio
		height = boxDim.h
	} else {
		width = boxDim.w
		height = boxDim.w * imgRatio
	}
	hzPerc := int(jsRound(1e5 * 0.5 * (1 - boxDim.w/width)))
	vzPerc := int(jsRound(1e5 * 0.5 * (1 - boxDim.h/height)))
	return `<a:srcRect l="` + itoa(hzPerc) + `" r="` + itoa(hzPerc) + `" t="` + itoa(vzPerc) + `" b="` + itoa(vzPerc) + `"/><a:stretch/>`
}

func imageSizingContain(imgSize imgSizeDim, boxDim imgBoxDim) string {
	imgRatio := imgSize.h / imgSize.w
	boxRatio := boxDim.h / boxDim.w
	widthBased := boxRatio > imgRatio
	var width, height float64
	if widthBased {
		width = boxDim.w
		height = boxDim.w * imgRatio
	} else {
		width = boxDim.h / imgRatio
		height = boxDim.h
	}
	hzPerc := int(jsRound(1e5 * 0.5 * (1 - boxDim.w/width)))
	vzPerc := int(jsRound(1e5 * 0.5 * (1 - boxDim.h/height)))
	return `<a:srcRect l="` + itoa(hzPerc) + `" r="` + itoa(hzPerc) + `" t="` + itoa(vzPerc) + `" b="` + itoa(vzPerc) + `"/><a:stretch/>`
}

func imageSizingCrop(imgSize imgSizeDim, boxDim imgBoxDim) string {
	l := boxDim.x
	r := imgSize.w - (boxDim.x + boxDim.w)
	t := boxDim.y
	b := imgSize.h - (boxDim.y + boxDim.h)
	lPerc := int(jsRound(1e5 * (l / imgSize.w)))
	rPerc := int(jsRound(1e5 * (r / imgSize.w)))
	tPerc := int(jsRound(1e5 * (t / imgSize.h)))
	bPerc := int(jsRound(1e5 * (b / imgSize.h)))
	return `<a:srcRect l="` + itoa(lPerc) + `" r="` + itoa(rPerc) + `" t="` + itoa(tPerc) + `" b="` + itoa(bPerc) + `"/><a:stretch/>`
}

func imageSizingXml(kind string, imgSize imgSizeDim, boxDim imgBoxDim) string {
	switch kind {
	case "cover":
		return imageSizingCover(imgSize, boxDim)
	case "contain":
		return imageSizingContain(imgSize, boxDim)
	case "crop":
		return imageSizingCrop(imgSize, boxDim)
	}
	return ""
}

// ---------------------------------------------------------------------------
// slideObjectToXml — transforms a slide / slideLayout / master to <p:cSld> XML.
//
// TS takes a `PresSlide | SlideLayout`. In Go we pass the shared SlideBaseProps
// plus the PresSlide's `_slideLayout` (nil for layouts/masters) used only for
// placeholder position inheritance.
// ---------------------------------------------------------------------------

func slideObjectToXml(slide *SlideBaseProps, slideLayout *SlideLayout) string {
	var strSlideXml string
	if slide.Name != "" {
		strSlideXml = `<p:cSld name="` + slide.Name + `">`
	} else {
		strSlideXml = "<p:cSld>"
	}
	intTableNum := 1

	// STEP 1: Background color/image
	if slide.BkgdImgRid != 0 {
		strSlideXml += `<p:bg><p:bgPr><a:blipFill dpi="0" rotWithShape="1"><a:blip r:embed="rId` + itoa(slide.BkgdImgRid) + `"><a:lum/></a:blip><a:srcRect/><a:stretch><a:fillRect/></a:stretch></a:blipFill><a:effectLst/></p:bgPr></p:bg>`
	} else if slide.Background != nil && slide.Background.Color != "" {
		strSlideXml += `<p:bg><p:bgPr>` + genXmlColorSelection(&slide.Background.ShapeFillProps) + `</p:bgPr></p:bg>`
	} else if slide.Bkgd == nil && slide.Name != "" && slide.Name == DEF_PRES_LAYOUT_NAME {
		strSlideXml += `<p:bg><p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef></p:bg>`
	}

	// STEP 2: spTree
	strSlideXml += "<p:spTree>"
	strSlideXml += `<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>`
	strSlideXml += `<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/>`
	strSlideXml += `<a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>`

	// STEP 3: Loop over slide objects
	for idx := range slide.SlideObjects {
		slideItemObj := &slide.SlideObjects[idx]

		x := 0
		y := 0
		cx := getSmartParseNumber(Percent(75), "X", slide.PresLayout)
		cy := 0
		var placeholderObj *SlideObject
		locationAttr := ""

		var sizing *ImageSizing
		var rounding *bool
		if slideItemObj.Options != nil {
			sizing = slideItemObj.Options.Sizing
			rounding = slideItemObj.Options.Rounding
		}

		if slideLayout != nil && slideLayout.SlideObjects != nil &&
			slideItemObj.Options != nil && slideItemObj.Options.Placeholder != "" {
			for li := range slideLayout.SlideObjects {
				lo := &slideLayout.SlideObjects[li]
				if lo.Options != nil && lo.Options.Placeholder == slideItemObj.Options.Placeholder {
					placeholderObj = lo
					break
				}
			}
		}

		// A: option vars
		if slideItemObj.Options == nil {
			slideItemObj.Options = &ObjectOptions{}
		}
		opts := slideItemObj.Options

		if opts.X != nil {
			x = getSmartParseNumber(*opts.X, "X", slide.PresLayout)
		}
		if opts.Y != nil {
			y = getSmartParseNumber(*opts.Y, "Y", slide.PresLayout)
		}
		if opts.W != nil {
			cx = getSmartParseNumber(*opts.W, "X", slide.PresLayout)
		}
		if opts.H != nil {
			cy = getSmartParseNumber(*opts.H, "Y", slide.PresLayout)
		}

		imgWidth := cx
		imgHeight := cy

		if placeholderObj != nil && placeholderObj.Options != nil {
			po := placeholderObj.Options
			if po.X != nil {
				x = getSmartParseNumber(*po.X, "X", slide.PresLayout)
			}
			if po.Y != nil {
				y = getSmartParseNumber(*po.Y, "Y", slide.PresLayout)
			}
			if po.W != nil {
				cx = getSmartParseNumber(*po.W, "X", slide.PresLayout)
			}
			if po.H != nil {
				cy = getSmartParseNumber(*po.H, "Y", slide.PresLayout)
			}
		}

		if boolDeref(opts.FlipH) {
			locationAttr += ` flipH="1"`
		}
		if boolDeref(opts.FlipV) {
			locationAttr += ` flipV="1"`
		}
		if opts.Rotate != 0 {
			locationAttr += ` rot="` + itoa(convertRotationDegrees(opts.Rotate)) + `"`
		}

		switch slideItemObj.Type {
		case SlideObjectTypeTable:
			strSlideXml += slideObjectTableToXml(slideItemObj, slide, intTableNum, x, y, cx, cy)
			intTableNum++

		case SlideObjectTypeText, SlideObjectTypePlaceholder:
			// Lines can have zero cy, but text should not
			if opts.Line == nil && cy == 0 {
				cy = int(jsRound(float64(EMU) * 0.3))
			}

			// Margin → _bodyProp insets
			if opts.BodyProp == nil {
				opts.BodyProp = &BodyProps{}
			}
			if opts.Margin != nil && len(opts.Margin) == 4 {
				opts.BodyProp.LIns = float64(valToPts(marginAt(opts.Margin, 0)))
				opts.BodyProp.RIns = float64(valToPts(marginAt(opts.Margin, 1)))
				opts.BodyProp.BIns = float64(valToPts(marginAt(opts.Margin, 2)))
				opts.BodyProp.TIns = float64(valToPts(marginAt(opts.Margin, 3)))
			} else if opts.Margin != nil && len(opts.Margin) == 1 {
				m := valToPts(opts.Margin[0])
				opts.BodyProp.LIns = float64(m)
				opts.BodyProp.RIns = float64(m)
				opts.BodyProp.BIns = float64(m)
				opts.BodyProp.TIns = float64(m)
			}

			// A: Start SHAPE
			strSlideXml += "<p:sp>"
			strSlideXml += `<p:nvSpPr><p:cNvPr id="` + itoa(idx+2) + `" name="` + opts.ObjectName + `">`
			if opts.Hyperlink != nil && opts.Hyperlink.URL != "" {
				strSlideXml += `<a:hlinkClick r:id="rId` + itoa(opts.Hyperlink.RID) + `" tooltip="` + tooltipVal(opts.Hyperlink) + `"/>`
			}
			if opts.Hyperlink != nil && opts.Hyperlink.Slide != 0 {
				strSlideXml += `<a:hlinkClick r:id="rId` + itoa(opts.Hyperlink.RID) + `" tooltip="` + tooltipVal(opts.Hyperlink) + `" action="ppaction://hlinksldjump"/>`
			}
			strSlideXml += "</p:cNvPr>"
			if boolDeref(opts.IsTextBox) {
				strSlideXml += `<p:cNvSpPr txBox="1"/>`
			} else {
				strSlideXml += `<p:cNvSpPr/>`
			}
			if slideItemObj.Type == SlideObjectTypePlaceholder {
				strSlideXml += `<p:nvPr>` + genXmlPlaceholder(slideItemObj) + `</p:nvPr>`
			} else {
				strSlideXml += `<p:nvPr>` + genXmlPlaceholder(placeholderObj) + `</p:nvPr>`
			}
			strSlideXml += "</p:nvSpPr><p:spPr>"
			strSlideXml += `<a:xfrm` + locationAttr + `>`
			strSlideXml += `<a:off x="` + itoa(x) + `" y="` + itoa(y) + `"/>`
			strSlideXml += `<a:ext cx="` + itoa(cx) + `" cy="` + itoa(cy) + `"/></a:xfrm>`

			if slideItemObj.Shape == ShapeTypeCustGeom {
				strSlideXml += `<a:custGeom><a:avLst />`
				strSlideXml += `<a:gdLst>`
				strSlideXml += `</a:gdLst>`
				strSlideXml += `<a:ahLst />`
				strSlideXml += `<a:cxnLst>`
				strSlideXml += `</a:cxnLst>`
				strSlideXml += `<a:rect l="l" t="t" r="r" b="b" />`
				strSlideXml += `<a:pathLst>`
				strSlideXml += `<a:path w="` + itoa(cx) + `" h="` + itoa(cy) + `">`
				for i := range opts.Points {
					point := opts.Points[i]
					if point.Curve != nil {
						switch point.Curve.Type {
						case "arc":
							strSlideXml += `<a:arcTo hR="` + itoa(getSmartParseNumber(point.Curve.HR, "Y", slide.PresLayout)) + `" wR="` + itoa(getSmartParseNumber(point.Curve.WR, "X", slide.PresLayout)) + `" stAng="` + itoa(convertRotationDegrees(point.Curve.StAng)) + `" swAng="` + itoa(convertRotationDegrees(point.Curve.SwAng)) + `" />`
						case "cubic":
							strSlideXml += `<a:cubicBezTo>
									<a:pt x="` + itoa(getSmartParseNumber(point.Curve.X1, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Curve.Y1, "Y", slide.PresLayout)) + `" />
									<a:pt x="` + itoa(getSmartParseNumber(point.Curve.X2, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Curve.Y2, "Y", slide.PresLayout)) + `" />
									<a:pt x="` + itoa(getSmartParseNumber(point.X, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Y, "Y", slide.PresLayout)) + `" />
									</a:cubicBezTo>`
						case "quadratic":
							strSlideXml += `<a:quadBezTo>
									<a:pt x="` + itoa(getSmartParseNumber(point.Curve.X1, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Curve.Y1, "Y", slide.PresLayout)) + `" />
									<a:pt x="` + itoa(getSmartParseNumber(point.X, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Y, "Y", slide.PresLayout)) + `" />
									</a:quadBezTo>`
						}
					} else if point.Close {
						strSlideXml += `<a:close />`
					} else if boolDeref(point.MoveTo) || i == 0 {
						strSlideXml += `<a:moveTo><a:pt x="` + itoa(getSmartParseNumber(point.X, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Y, "Y", slide.PresLayout)) + `" /></a:moveTo>`
					} else {
						strSlideXml += `<a:lnTo><a:pt x="` + itoa(getSmartParseNumber(point.X, "X", slide.PresLayout)) + `" y="` + itoa(getSmartParseNumber(point.Y, "Y", slide.PresLayout)) + `" /></a:lnTo>`
					}
				}
				strSlideXml += `</a:path>`
				strSlideXml += `</a:pathLst>`
				strSlideXml += `</a:custGeom>`
			} else {
				strSlideXml += `<a:prstGeom prst="` + string(slideItemObj.Shape) + `"><a:avLst>`
				if opts.RectRadius != 0 {
					minDim := cx
					if cy < minDim {
						minDim = cy
					}
					strSlideXml += `<a:gd name="adj" fmla="val ` + itoa(int(jsRound((opts.RectRadius*float64(EMU)*100000)/float64(minDim)))) + `"/>`
				} else if opts.AngleRange != nil {
					for i := 0; i < 2; i++ {
						angle := opts.AngleRange[i]
						strSlideXml += `<a:gd name="adj` + itoa(i+1) + `" fmla="val ` + itoa(convertRotationDegrees(angle)) + `" />`
					}
					if opts.ArcThicknessRatio != 0 {
						strSlideXml += `<a:gd name="adj3" fmla="val ` + itoa(int(jsRound(opts.ArcThicknessRatio*50000))) + `" />`
					}
				}
				strSlideXml += `</a:avLst></a:prstGeom>`
			}

			// Option: FILL
			if opts.Fill != nil {
				strSlideXml += genXmlColorSelection(opts.Fill)
			} else {
				strSlideXml += `<a:noFill/>`
			}

			// LINE
			if opts.Line != nil {
				if opts.Line.Width != 0 {
					strSlideXml += `<a:ln w="` + itoa(valToPts(opts.Line.Width)) + `">`
				} else {
					strSlideXml += `<a:ln>`
				}
				if opts.Line.Color != "" {
					strSlideXml += genXmlColorSelection(&opts.Line.ShapeFillProps)
				}
				if opts.Line.DashType != "" {
					strSlideXml += `<a:prstDash val="` + opts.Line.DashType + `"/>`
				}
				if opts.Line.BeginArrowType != "" {
					strSlideXml += `<a:headEnd type="` + opts.Line.BeginArrowType + `"/>`
				}
				if opts.Line.EndArrowType != "" {
					strSlideXml += `<a:tailEnd type="` + opts.Line.EndArrowType + `"/>`
				}
				strSlideXml += `</a:ln>`
			}

			// SHADOW
			if opts.Shadow != nil && opts.Shadow.Type != "none" {
				strSlideXml += shadowXml(opts.Shadow)
			}

			strSlideXml += "</p:spPr>"

			// Text body
			strSlideXml += genXmlTextBody(slideItemObj)

			strSlideXml += "</p:sp>"

		case SlideObjectTypeImage:
			strSlideXml += slideObjectImageToXml(slideItemObj, slide, placeholderObj, x, y, cx, cy, imgWidth, imgHeight, sizing, rounding, locationAttr)

		case SlideObjectTypeMedia:
			strSlideXml += slideObjectMediaToXml(slideItemObj, x, y, cx, cy, locationAttr)

		case SlideObjectTypeChart:
			strSlideXml += `<p:graphicFrame>`
			strSlideXml += ` <p:nvGraphicFramePr>`
			strSlideXml += `   <p:cNvPr id="` + itoa(idx+2) + `" name="` + opts.ObjectName + `" descr="` + encodeXmlEntities(opts.AltText) + `"/>`
			strSlideXml += `   <p:cNvGraphicFramePr/>`
			strSlideXml += `   <p:nvPr>` + genXmlPlaceholder(placeholderObj) + `</p:nvPr>`
			strSlideXml += ` </p:nvGraphicFramePr>`
			strSlideXml += ` <p:xfrm><a:off x="` + itoa(x) + `" y="` + itoa(y) + `"/><a:ext cx="` + itoa(cx) + `" cy="` + itoa(cy) + `"/></p:xfrm>`
			strSlideXml += ` <a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`
			strSlideXml += `  <a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/chart">`
			strSlideXml += `   <c:chart r:id="rId` + itoa(slideItemObj.ChartRID) + `" xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart"/>`
			strSlideXml += `  </a:graphicData>`
			strSlideXml += ` </a:graphic>`
			strSlideXml += `</p:graphicFrame>`

		default:
			strSlideXml += ""
		}
	}

	// STEP 4: Slide numbers
	if slide.SlideNumberProps != nil {
		snp := slide.SlideNumberProps
		if snp.Align == "" {
			snp.Align = "left"
		}

		strSlideXml += "<p:sp>"
		strSlideXml += " <p:nvSpPr>"
		strSlideXml += `  <p:cNvPr id="25" name="Slide Number Placeholder 0"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>`
		strSlideXml += `  <p:nvPr><p:ph type="sldNum" sz="quarter" idx="4294967295"/></p:nvPr>`
		strSlideXml += " </p:nvSpPr>"
		strSlideXml += " <p:spPr>"
		snpX := 0
		if snp.X != nil {
			snpX = getSmartParseNumber(*snp.X, "X", slide.PresLayout)
		}
		snpY := 0
		if snp.Y != nil {
			snpY = getSmartParseNumber(*snp.Y, "Y", slide.PresLayout)
		}
		wStr := "800000"
		if snp.W != nil {
			wStr = itoa(getSmartParseNumber(*snp.W, "X", slide.PresLayout))
		}
		hStr := "300000"
		if snp.H != nil {
			hStr = itoa(getSmartParseNumber(*snp.H, "Y", slide.PresLayout))
		}
		strSlideXml += `<a:xfrm>` +
			`<a:off x="` + itoa(snpX) + `" y="` + itoa(snpY) + `"/>` +
			`<a:ext cx="` + wStr + `" cy="` + hStr + `"/>` +
			`</a:xfrm>` +
			` <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>` +
			` <a:extLst><a:ext uri="{C572A759-6A51-4108-AA02-DFA0A04FC94B}"><ma14:wrappingTextBoxFlag val="0" xmlns:ma14="http://schemas.microsoft.com/office/mac/drawingml/2011/main"/></a:ext></a:extLst>` +
			`</p:spPr>`
		strSlideXml += "<p:txBody>"
		strSlideXml += "<a:bodyPr"
		if snp.Margin != nil && len(snp.Margin) == 4 {
			strSlideXml += ` lIns="` + itoa(valToPts(marginAt(snp.Margin, 3))) + `"`
			strSlideXml += ` tIns="` + itoa(valToPts(marginAt(snp.Margin, 0))) + `"`
			strSlideXml += ` rIns="` + itoa(valToPts(marginAt(snp.Margin, 1))) + `"`
			strSlideXml += ` bIns="` + itoa(valToPts(marginAt(snp.Margin, 2))) + `"`
		} else if snp.Margin != nil && len(snp.Margin) == 1 {
			m := valToPts(snp.Margin[0])
			strSlideXml += ` lIns="` + itoa(m) + `"`
			strSlideXml += ` tIns="` + itoa(m) + `"`
			strSlideXml += ` rIns="` + itoa(m) + `"`
			strSlideXml += ` bIns="` + itoa(m) + `"`
		}
		if snp.Valign != "" {
			v := snp.Valign
			v = strings.ReplaceAll(v, "top", "t")
			v = strings.ReplaceAll(v, "middle", "ctr")
			v = strings.ReplaceAll(v, "bottom", "b")
			strSlideXml += ` anchor="` + v + `"`
		}
		strSlideXml += "/>"
		strSlideXml += `  <a:lstStyle><a:lvl1pPr>`
		if snp.FontFace != "" || snp.FontSize != 0 || snp.Color != "" {
			fs := snp.FontSize
			if fs == 0 {
				fs = 12
			}
			strSlideXml += `<a:defRPr sz="` + itoa(int(jsRound(fs*100))) + `">`
			if snp.Color != "" {
				strSlideXml += genXmlColorSelection(&ShapeFillProps{Color: snp.Color})
			}
			if snp.FontFace != "" {
				strSlideXml += `<a:latin typeface="` + snp.FontFace + `"/><a:ea typeface="` + snp.FontFace + `"/><a:cs typeface="` + snp.FontFace + `"/>`
			}
			strSlideXml += `</a:defRPr>`
		}
		strSlideXml += `</a:lvl1pPr></a:lstStyle>`
		strSlideXml += "<a:p>"
		if strings.HasPrefix(snp.Align, "l") {
			strSlideXml += `<a:pPr algn="l"/>`
		} else if strings.HasPrefix(snp.Align, "c") {
			strSlideXml += `<a:pPr algn="ctr"/>`
		} else if strings.HasPrefix(snp.Align, "r") {
			strSlideXml += `<a:pPr algn="r"/>`
		} else {
			strSlideXml += `<a:pPr algn="l"/>`
		}
		boldVal := "0"
		if boolDeref(snp.Bold) {
			boldVal = "1"
		}
		strSlideXml += `<a:fld id="` + SLDNUMFLDID + `" type="slidenum"><a:rPr b="` + boldVal + `" lang="en-US"/>`
		strSlideXml += `<a:t>` + itoa(slide.SlideNum) + `</a:t></a:fld><a:endParaRPr lang="en-US"/></a:p>`
		strSlideXml += "</p:txBody></p:sp>"
	}

	// STEP 5: Close
	strSlideXml += "</p:spTree>"
	strSlideXml += "</p:cSld>"

	return strSlideXml
}

// tooltipVal encodes the hyperlink tooltip (empty when unset), mirroring
// `hyperlink.tooltip ? encodeXmlEntities(hyperlink.tooltip) : ”`.
func tooltipVal(h *HyperlinkProps) string {
	if h != nil && h.Tooltip != "" {
		return encodeXmlEntities(h.Tooltip)
	}
	return ""
}

// marginAt returns m[i] or 0 (mirrors `margin[i] || 0`).
func marginAt(m Margin, i int) float64 {
	if i < len(m) {
		return m[i]
	}
	return 0
}

// shadowXml builds the <a:effectLst> shadow for text/placeholder/image shapes.
// It mutates the shadow struct exactly as the TS does before rendering.
func shadowXml(sh *ShadowProps) string {
	sh.Type = strOr(sh.Type, "outer")
	blur := sh.Blur
	if blur == 0 {
		blur = 8
	}
	blurV := valToPts(blur)
	offset := sh.Offset
	if offset == 0 {
		offset = 4
	}
	offsetV := valToPts(offset)
	angle := sh.Angle
	if angle == 0 {
		angle = 270
	}
	angleV := int(jsRound(angle * 60000))
	opacity := sh.Opacity
	if opacity == 0 {
		opacity = 0.75
	}
	opacityV := int(jsRound(opacity * 100000))
	color := strOr(sh.Color, DEF_TEXT_SHADOW.Color)

	var b strings.Builder
	b.WriteString("<a:effectLst>")
	shdwAttrs := ""
	if sh.Type == "outer" {
		shdwAttrs = `sx="100000" sy="100000" kx="0" ky="0" algn="bl" rotWithShape="0"`
	}
	b.WriteString(` <a:` + sh.Type + `Shdw ` + shdwAttrs + ` blurRad="` + itoa(blurV) + `" dist="` + itoa(offsetV) + `" dir="` + itoa(angleV) + `">`)
	b.WriteString(` <a:srgbClr val="` + color + `">`)
	b.WriteString(` <a:alpha val="` + itoa(opacityV) + `"/></a:srgbClr>`)
	b.WriteString(` </a:outerShdw>`)
	b.WriteString("</a:effectLst>")
	return b.String()
}

// shadowImageXml mirrors the (slightly different) shadow block in the image case.
func shadowImageXml(sh *ShadowProps) string {
	sh.Type = strOr(sh.Type, "outer")
	blur := sh.Blur
	if blur == 0 {
		blur = 8
	}
	blurV := valToPts(blur)
	offset := sh.Offset
	if offset == 0 {
		offset = 4
	}
	offsetV := valToPts(offset)
	angle := sh.Angle
	if angle == 0 {
		angle = 270
	}
	angleV := int(jsRound(angle * 60000))
	opacity := sh.Opacity
	if opacity == 0 {
		opacity = 0.75
	}
	opacityV := int(jsRound(opacity * 100000))
	color := strOr(sh.Color, DEF_TEXT_SHADOW.Color)

	var b strings.Builder
	b.WriteString("<a:effectLst>")
	shdwAttrs := ""
	if sh.Type == "outer" {
		shdwAttrs = `sx="100000" sy="100000" kx="0" ky="0" algn="bl" rotWithShape="0"`
	}
	b.WriteString(`<a:` + sh.Type + `Shdw ` + shdwAttrs + ` blurRad="` + itoa(blurV) + `" dist="` + itoa(offsetV) + `" dir="` + itoa(angleV) + `">`)
	b.WriteString(`<a:srgbClr val="` + color + `">`)
	b.WriteString(`<a:alpha val="` + itoa(opacityV) + `"/></a:srgbClr>`)
	b.WriteString(`</a:` + sh.Type + `Shdw>`)
	b.WriteString("</a:effectLst>")
	return b.String()
}

// slideObjectImageToXml renders the SLIDE_OBJECT_TYPES.image case.
func slideObjectImageToXml(slideItemObj *SlideObject, slide *SlideBaseProps, placeholderObj *SlideObject, x, y, cx, cy, imgWidth, imgHeight int, sizing *ImageSizing, rounding *bool, locationAttr string) string {
	opts := slideItemObj.Options
	var s strings.Builder
	s.WriteString("<p:pic>")
	s.WriteString("  <p:nvPicPr>")
	descr := strOr(opts.AltText, slideItemObj.Image)
	s.WriteString(`<p:cNvPr id="` + itoa(imageObjIdx(slide, slideItemObj)) + `" name="` + opts.ObjectName + `" descr="` + encodeXmlEntities(descr) + `">`)
	if slideItemObj.Hyperlink != nil && slideItemObj.Hyperlink.URL != "" {
		s.WriteString(`<a:hlinkClick r:id="rId` + itoa(slideItemObj.Hyperlink.RID) + `" tooltip="` + tooltipVal(slideItemObj.Hyperlink) + `"/>`)
	}
	if slideItemObj.Hyperlink != nil && slideItemObj.Hyperlink.Slide != 0 {
		s.WriteString(`<a:hlinkClick r:id="rId` + itoa(slideItemObj.Hyperlink.RID) + `" tooltip="` + tooltipVal(slideItemObj.Hyperlink) + `" action="ppaction://hlinksldjump"/>`)
	}
	s.WriteString("    </p:cNvPr>")
	s.WriteString(`    <p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr>`)
	s.WriteString("    <p:nvPr>" + genXmlPlaceholder(placeholderObj) + "</p:nvPr>")
	s.WriteString("  </p:nvPicPr>")
	s.WriteString("<p:blipFill>")

	isSvg := false
	for i := range slide.RelsMedia {
		if slide.RelsMedia[i].RID == slideItemObj.ImageRID {
			isSvg = slide.RelsMedia[i].Extn == "svg"
			break
		}
	}
	if isSvg {
		s.WriteString(`<a:blip r:embed="rId` + itoa(slideItemObj.ImageRID-1) + `">`)
		if opts.Transparency != 0 {
			s.WriteString(` <a:alphaModFix amt="` + itoa(int(jsRound((100-opts.Transparency)*1000))) + `"/>`)
		}
		s.WriteString(" <a:extLst>")
		s.WriteString("  <a:ext uri=\"{96DAC541-7B7A-43D3-8B79-37D633B846F1}\">")
		s.WriteString(`   <asvg:svgBlip xmlns:asvg="http://schemas.microsoft.com/office/drawing/2016/SVG/main" r:embed="rId` + itoa(slideItemObj.ImageRID) + `"/>`)
		s.WriteString("  </a:ext>")
		s.WriteString(" </a:extLst>")
		s.WriteString("</a:blip>")
	} else {
		s.WriteString(`<a:blip r:embed="rId` + itoa(slideItemObj.ImageRID) + `">`)
		if opts.Transparency != 0 {
			s.WriteString(`<a:alphaModFix amt="` + itoa(int(jsRound((100-opts.Transparency)*1000))) + `"/>`)
		}
		s.WriteString("</a:blip>")
	}

	if sizing != nil && sizing.Type != "" {
		boxW := cx
		if sizing.W.Val != 0 || sizing.W.IsPct {
			boxW = getSmartParseNumber(sizing.W, "X", slide.PresLayout)
		}
		boxH := cy
		if sizing.H.Val != 0 || sizing.H.IsPct {
			boxH = getSmartParseNumber(sizing.H, "Y", slide.PresLayout)
		}
		boxX := 0
		if sizing.X != nil {
			boxX = getSmartParseNumber(*sizing.X, "X", slide.PresLayout)
		}
		boxY := 0
		if sizing.Y != nil {
			boxY = getSmartParseNumber(*sizing.Y, "Y", slide.PresLayout)
		}
		s.WriteString(imageSizingXml(sizing.Type, imgSizeDim{w: float64(imgWidth), h: float64(imgHeight)}, imgBoxDim{w: float64(boxW), h: float64(boxH), x: float64(boxX), y: float64(boxY)}))
		imgWidth = boxW
		imgHeight = boxH
	} else {
		s.WriteString("  <a:stretch><a:fillRect/></a:stretch>")
	}
	s.WriteString("</p:blipFill>")
	s.WriteString("<p:spPr>")
	s.WriteString(" <a:xfrm" + locationAttr + ">")
	s.WriteString(`  <a:off x="` + itoa(x) + `" y="` + itoa(y) + `"/>`)
	s.WriteString(`  <a:ext cx="` + itoa(imgWidth) + `" cy="` + itoa(imgHeight) + `"/>`)
	s.WriteString(" </a:xfrm>")
	prst := "rect"
	if boolDeref(rounding) {
		prst = "ellipse"
	}
	s.WriteString(` <a:prstGeom prst="` + prst + `"><a:avLst/></a:prstGeom>`)

	if opts.Shadow != nil && opts.Shadow.Type != "none" {
		s.WriteString(shadowImageXml(opts.Shadow))
	}
	s.WriteString("</p:spPr>")
	s.WriteString("</p:pic>")
	return s.String()
}

// imageObjIdx computes the image cNvPr id (`idx + 2`). We recompute the object's
// index within the slide since the image branch is factored out of the loop.
func imageObjIdx(slide *SlideBaseProps, target *SlideObject) int {
	for i := range slide.SlideObjects {
		if &slide.SlideObjects[i] == target {
			return i + 2
		}
	}
	return 2
}

// slideObjectMediaToXml renders the SLIDE_OBJECT_TYPES.media case.
func slideObjectMediaToXml(slideItemObj *SlideObject, x, y, cx, cy int, locationAttr string) string {
	opts := slideItemObj.Options
	var s strings.Builder
	if slideItemObj.Mtype == "online" {
		s.WriteString("<p:pic>")
		s.WriteString(" <p:nvPicPr>")
		s.WriteString(`<p:cNvPr id="` + itoa(slideItemObj.MediaRID+2) + `" name="` + opts.ObjectName + `"/>`)
		s.WriteString(" <p:cNvPicPr/>")
		s.WriteString(" <p:nvPr>")
		s.WriteString(`  <a:videoFile r:link="rId` + itoa(slideItemObj.MediaRID) + `"/>`)
		s.WriteString(" </p:nvPr>")
		s.WriteString(" </p:nvPicPr>")
		s.WriteString(` <p:blipFill><a:blip r:embed="rId` + itoa(slideItemObj.MediaRID+1) + `"/><a:stretch><a:fillRect/></a:stretch></p:blipFill>`)
		s.WriteString(" <p:spPr>")
		s.WriteString(`  <a:xfrm` + locationAttr + `><a:off x="` + itoa(x) + `" y="` + itoa(y) + `"/><a:ext cx="` + itoa(cx) + `" cy="` + itoa(cy) + `"/></a:xfrm>`)
		s.WriteString(`  <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`)
		s.WriteString(" </p:spPr>")
		s.WriteString("</p:pic>")
	} else {
		s.WriteString("<p:pic>")
		s.WriteString(" <p:nvPicPr>")
		s.WriteString(`<p:cNvPr id="` + itoa(slideItemObj.MediaRID+2) + `" name="` + opts.ObjectName + `"><a:hlinkClick r:id="" action="ppaction://media"/></p:cNvPr>`)
		s.WriteString(` <p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr>`)
		s.WriteString(" <p:nvPr>")
		s.WriteString(`  <a:videoFile r:link="rId` + itoa(slideItemObj.MediaRID) + `"/>`)
		s.WriteString("  <p:extLst>")
		s.WriteString(`   <p:ext uri="{DAA4B4D4-6D71-4841-9C94-3DE7FCFB9230}">`)
		s.WriteString(`    <p14:media xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" r:embed="rId` + itoa(slideItemObj.MediaRID+1) + `"/>`)
		s.WriteString("   </p:ext>")
		s.WriteString("  </p:extLst>")
		s.WriteString(" </p:nvPr>")
		s.WriteString(" </p:nvPicPr>")
		s.WriteString(` <p:blipFill><a:blip r:embed="rId` + itoa(slideItemObj.MediaRID+2) + `"/><a:stretch><a:fillRect/></a:stretch></p:blipFill>`)
		s.WriteString(" <p:spPr>")
		s.WriteString(`  <a:xfrm` + locationAttr + `><a:off x="` + itoa(x) + `" y="` + itoa(y) + `"/><a:ext cx="` + itoa(cx) + `" cy="` + itoa(cy) + `"/></a:xfrm>`)
		s.WriteString(`  <a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`)
		s.WriteString(" </p:spPr>")
		s.WriteString("</p:pic>")
	}
	return s.String()
}

// ---------------------------------------------------------------------------
// slideObjectTableToXml — the SLIDE_OBJECT_TYPES.table case of slideObjectToXml.
//
// NOTE: like the TS, this mutates arrTabRows in place (inserting hmerge/vmerge
// dummy cells). It is invoked once per table during a single render pass.
// ---------------------------------------------------------------------------

func slideObjectTableToXml(slideItemObj *SlideObject, slide *SlideBaseProps, intTableNum, x, y, cx, cy int) string {
	opts := slideItemObj.Options
	arrTabRows := slideItemObj.ArrTabRows
	objTabOpts := opts

	// Calc number of columns
	intColCnt := 0
	if len(arrTabRows) > 0 {
		for ci := range arrTabRows[0] {
			c := &arrTabRows[0][ci]
			if c.Options != nil && c.Options.Colspan != 0 {
				intColCnt += c.Options.Colspan
			} else {
				intColCnt++
			}
		}
	}

	var strXml strings.Builder

	strXml.WriteString(`<p:graphicFrame><p:nvGraphicFramePr><p:cNvPr id="` + itoa(intTableNum*slide.SlideNum+1) + `" name="` + opts.ObjectName + `"/>`)
	strXml.WriteString(`<p:cNvGraphicFramePr><a:graphicFrameLocks noGrp="1"/></p:cNvGraphicFramePr>` +
		`  <p:nvPr><p:extLst><p:ext uri="{D42A27DB-BD31-4B8C-83A1-F6EECF244321}"><p14:modId xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" val="1579011935"/></p:ext></p:extLst></p:nvPr>` +
		`</p:nvGraphicFramePr>`)
	strXml.WriteString(`<p:xfrm><a:off x="` + tblCoord(x) + `" y="` + tblCoord(y) + `"/><a:ext cx="` + tblCoord(cx) + `" cy="` + tblCoordCy(cy) + `"/></p:xfrm>`)
	strXml.WriteString(`<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/table"><a:tbl><a:tblPr/>`)

	// Column widths
	if len(objTabOpts.ColW) > 0 {
		strXml.WriteString(`<a:tblGrid>`)
		for col := 0; col < intColCnt; col++ {
			var w int
			if col < len(objTabOpts.ColW) {
				w = inch2Emu(objTabOpts.ColW[col])
			} else {
				wv := 1.0
				if opts.W != nil {
					wv = opts.W.Val
				}
				w = int(wv / float64(intColCnt))
			}
			strXml.WriteString(`<a:gridCol w="` + itoa(int(jsRound(float64(w)))) + `"/>`)
		}
		strXml.WriteString(`</a:tblGrid>`)
	} else {
		intColW := EMU
		if opts.W != nil && len(objTabOpts.ColW) == 0 {
			wv := opts.W.Val
			intColW = int(jsRound(wv / float64(intColCnt)))
		}
		strXml.WriteString(`<a:tblGrid>`)
		for colw := 0; colw < intColCnt; colw++ {
			strXml.WriteString(`<a:gridCol w="` + itoa(intColW) + `"/>`)
		}
		strXml.WriteString(`</a:tblGrid>`)
	}

	// STEP 3A: add _hmerge cells for colspan
	for ri := range arrTabRows {
		cells := arrTabRows[ri]
		for cIdx := 0; cIdx < len(cells); {
			cell := cells[cIdx]
			colspan := 0
			var rowspan int
			if cell.Options != nil {
				colspan = cell.Options.Colspan
				rowspan = cell.Options.Rowspan
			}
			if colspan > 1 {
				dummies := make([]TableCell, colspan-1)
				for k := range dummies {
					dummies[k] = TableCell{Type: SlideObjectTypeTablecell, Options: &TableCellProps{Rowspan: rowspan}, Hmerge: ptr(true)}
				}
				cells = append(cells[:cIdx+1], append(dummies, cells[cIdx+1:]...)...)
				cIdx += colspan
			} else {
				cIdx++
			}
		}
		arrTabRows[ri] = cells
	}

	// STEP 3B: add _vmerge cells for rowspan
	for rIdx := range arrTabRows {
		if rIdx+1 >= len(arrTabRows) {
			continue
		}
		cells := arrTabRows[rIdx]
		for cIdx := range cells {
			cell := cells[cIdx]
			rowspan := cell.RowContinue
			if rowspan == 0 && cell.Options != nil {
				rowspan = cell.Options.Rowspan
			}
			colspan := 0
			if cell.Options != nil {
				colspan = cell.Options.Colspan
			}
			if rowspan > 1 {
				hmerge := cell.Hmerge
				merge := TableCell{Type: SlideObjectTypeTablecell, Options: &TableCellProps{Colspan: colspan}, RowContinue: rowspan - 1, Vmerge: ptr(true), Hmerge: hmerge}
				nextRow := arrTabRows[rIdx+1]
				if cIdx <= len(nextRow) {
					nextRow = append(nextRow[:cIdx], append([]TableCell{merge}, nextRow[cIdx:]...)...)
					arrTabRows[rIdx+1] = nextRow
				}
			}
		}
	}

	// STEP 4: rows/cells
	for rIdx := range arrTabRows {
		cells := arrTabRows[rIdx]
		intRowH := 0
		if len(objTabOpts.RowH) > rIdx && objTabOpts.RowH[rIdx] != 0 {
			intRowH = inch2Emu(objTabOpts.RowH[rIdx])
		} else if len(objTabOpts.RowH) == 1 && objTabOpts.RowH[0] != 0 {
			intRowH = inch2Emu(objTabOpts.RowH[0])
		} else if opts.Cy != nil || opts.H != nil {
			var num float64
			if opts.H != nil {
				num = float64(inch2Emu(opts.H.Val))
			} else if opts.Cy != nil {
				num = opts.Cy.Val
			} else {
				num = 1
			}
			intRowH = int(jsRound(num / float64(len(arrTabRows))))
		}

		strXml.WriteString(`<a:tr h="` + itoa(intRowH) + `">`)

		for cellIdx := range cells {
			cell := &cells[cellIdx]

			cellSpanAttrStr := tableCellSpanAttrs(cell)

			if boolDeref(cell.Hmerge) || boolDeref(cell.Vmerge) {
				strXml.WriteString(`<a:tc` + cellSpanAttrStr + `><a:tcPr/></a:tc>`)
				continue
			}

			if cell.Options == nil {
				cell.Options = &TableCellProps{}
			}
			cellOpts := cell.Options

			// Inherit table opts into cell
			inheritCellFromTable(cellOpts, objTabOpts)

			cellValign := ""
			if cellOpts.Valign != "" {
				v := cellOpts.Valign
				v = replaceValign(v)
				cellValign = ` anchor="` + v + `"`
			}
			cellTextDir := ""
			if cellOpts.TextDirection != "" && cellOpts.TextDirection != "horz" {
				cellTextDir = ` vert="` + cellOpts.TextDirection + `"`
			}

			cellFill := ""
			if cellOpts.Fill != nil && cellOpts.Fill.Color != "" {
				cellFill = genXmlColorSelection(cellOpts.Fill)
			}

			// Margin
			var cellMargin []float64
			if cellOpts.Margin != nil {
				cellMargin = append([]float64(nil), cellOpts.Margin...)
			} else {
				cellMargin = DEF_CELL_MARGIN_IN[:]
			}
			if len(cellMargin) == 1 {
				m := cellMargin[0]
				cellMargin = []float64{m, m, m, m}
			}
			cellMarginXml := ""
			if len(cellMargin) >= 4 {
				if cellMargin[0] >= 1 {
					cellMarginXml = ` marL="` + itoa(valToPts(cellMargin[3])) + `" marR="` + itoa(valToPts(cellMargin[1])) + `" marT="` + itoa(valToPts(cellMargin[0])) + `" marB="` + itoa(valToPts(cellMargin[2])) + `"`
				} else {
					cellMarginXml = ` marL="` + itoa(inch2Emu(cellMargin[3])) + `" marR="` + itoa(inch2Emu(cellMargin[1])) + `" marT="` + itoa(inch2Emu(cellMargin[0])) + `" marB="` + itoa(inch2Emu(cellMargin[2])) + `"`
				}
			}

			strXml.WriteString(`<a:tc` + cellSpanAttrStr + `>` + genXmlTableCellTextBody(cell) + `<a:tcPr` + cellMarginXml + cellValign + cellTextDir + `>`)

			// Borders (LRTB)
			if len(cellOpts.Border) >= 4 {
				for _, bd := range []struct {
					idx  int
					name string
				}{{3, "lnL"}, {1, "lnR"}, {0, "lnT"}, {2, "lnB"}} {
					b := cellOpts.Border[bd.idx]
					if b.Type != "none" {
						strXml.WriteString(`<a:` + bd.name + ` w="` + itoa(valToPts(b.Pt)) + `" cap="flat" cmpd="sng" algn="ctr">`)
						strXml.WriteString(`<a:solidFill>` + createColorElement(b.Color, "") + `</a:solidFill>`)
						dash := "solid"
						if b.Type == "dash" {
							dash = "sysDash"
						}
						strXml.WriteString(`<a:prstDash val="` + dash + `"/><a:round/><a:headEnd type="none" w="med" len="med"/><a:tailEnd type="none" w="med" len="med"/>`)
						strXml.WriteString(`</a:` + bd.name + `>`)
					} else {
						strXml.WriteString(`<a:` + bd.name + ` w="0" cap="flat" cmpd="sng" algn="ctr"><a:noFill/></a:` + bd.name + `>`)
					}
				}
			}

			strXml.WriteString(cellFill)
			strXml.WriteString(`  </a:tcPr>`)
			strXml.WriteString(` </a:tc>`)
		}

		strXml.WriteString(`</a:tr>`)
	}

	strXml.WriteString(`      </a:tbl>`)
	strXml.WriteString(`    </a:graphicData>`)
	strXml.WriteString(`  </a:graphic>`)
	strXml.WriteString(`</p:graphicFrame>`)

	return strXml.String()
}

// tblCoord mirrors `x || (x === 0 ? 0 : EMU)` for the table xfrm offsets/exts.
func tblCoord(v int) string {
	if v != 0 {
		return itoa(v)
	}
	return "0"
}

// tblCoordCy mirrors `cy || EMU` (cy of 0 falls back to EMU).
func tblCoordCy(v int) string {
	if v != 0 {
		return itoa(v)
	}
	return itoa(EMU)
}

// tableCellSpanAttrs builds the rowSpan/gridSpan/vMerge/hMerge attribute string.
func tableCellSpanAttrs(cell *TableCell) string {
	parts := []string{}
	if cell.Options != nil && cell.Options.Rowspan > 1 {
		parts = append(parts, `rowSpan="`+itoa(cell.Options.Rowspan)+`"`)
	}
	if cell.Options != nil && cell.Options.Colspan > 1 {
		parts = append(parts, `gridSpan="`+itoa(cell.Options.Colspan)+`"`)
	}
	if boolDeref(cell.Vmerge) {
		parts = append(parts, `vMerge="1"`)
	}
	if boolDeref(cell.Hmerge) {
		parts = append(parts, `hMerge="1"`)
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// replaceValign applies the TS anchor normalization chain to a cell valign.
func replaceValign(v string) string {
	// ^c$ / ^m$ (case-insensitive) → ctr
	if v == "c" || v == "C" || v == "m" || v == "M" {
		return "ctr"
	}
	v = strings.ReplaceAll(v, "center", "ctr")
	v = strings.ReplaceAll(v, "middle", "ctr")
	v = strings.ReplaceAll(v, "top", "t")
	v = strings.ReplaceAll(v, "btm", "b")
	v = strings.ReplaceAll(v, "bottom", "b")
	return v
}

// inheritCellFromTable copies the inheritable table options into a cell when the
// cell has not set them (mirrors the TS name-list inheritance loop).
func inheritCellFromTable(cell *TableCellProps, tbl *ObjectOptions) {
	if tbl.Align != "" && cell.Align == "" {
		cell.Align = tbl.Align
	}
	if tbl.Bold != nil && cell.Bold == nil {
		cell.Bold = tbl.Bold
	}
	if len(tbl.Border) > 0 && cell.Border == nil {
		cell.Border = tbl.Border
	}
	if tbl.Color != "" && cell.Color == "" {
		cell.Color = tbl.Color
	}
	if tbl.Fill != nil && cell.Fill == nil {
		cell.Fill = tbl.Fill
	}
	if tbl.FontFace != "" && cell.FontFace == "" {
		cell.FontFace = tbl.FontFace
	}
	if tbl.FontSize != 0 && cell.FontSize == 0 {
		cell.FontSize = tbl.FontSize
	}
	if tbl.Margin != nil && cell.Margin == nil {
		cell.Margin = tbl.Margin
	}
	if tbl.TextDirection != "" && cell.TextDirection == "" {
		cell.TextDirection = tbl.TextDirection
	}
	if tbl.Underline != nil && cell.Underline == nil {
		cell.Underline = tbl.Underline
	}
	if tbl.Valign != "" && cell.Valign == "" {
		cell.Valign = tbl.Valign
	}
}

// genXmlTableCellTextBody renders a table cell's <a:txBody>.
func genXmlTableCellTextBody(cell *TableCell) string {
	opts := objOptsFromCellProps(cell.Options)
	var runs []textRun
	if len(cell.TextCells) > 0 {
		for i := range cell.TextCells {
			tc := &cell.TextCells[i]
			runs = append(runs, textRun{text: tc.Text, options: objOptsFromCellProps(tc.Options)})
		}
	} else {
		runs = append(runs, textRun{text: cell.Text, options: objOptsFromCellProps(cell.Options)})
	}
	return genXmlTextBodyCore(SlideObjectTypeTablecell, opts, runs, false)
}

// objOptsFromCellProps converts a *TableCellProps to *ObjectOptions for the
// shared text builders.
func objOptsFromCellProps(tcp *TableCellProps) *ObjectOptions {
	o := &ObjectOptions{}
	if tcp == nil {
		return o
	}
	o.TextBaseProps = tcp.TextBaseProps
	o.Border = tcp.Border
	o.Fill = tcp.Fill
	o.Hyperlink = tcp.Hyperlink
	o.Margin = tcp.Margin
	o.Colspan = tcp.Colspan
	o.Rowspan = tcp.Rowspan
	o.AutoPageCharWeight = tcp.AutoPageCharWeight
	o.AutoPageLineWeight = tcp.AutoPageLineWeight
	return o
}

// ---------------------------------------------------------------------------
// slideObjectRelationsToXml
// ---------------------------------------------------------------------------

type defaultRel struct {
	target string
	typ    string
}

func slideObjectRelationsToXml(slide *SlideBaseProps, defaultRels []defaultRel) string {
	lastRid := 0
	strXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`

	for i := range slide.Rels {
		rel := &slide.Rels[i]
		if rel.RID > lastRid {
			lastRid = rel.RID
		}
		lt := strings.ToLower(string(rel.Type))
		if strings.Contains(lt, "hyperlink") {
			if dataStr, ok := rel.Data.(string); ok && dataStr == "slide" {
				strXml += `<Relationship Id="rId` + itoa(rel.RID) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slide` + rel.Target + `.xml"/>`
			} else {
				strXml += `<Relationship Id="rId` + itoa(rel.RID) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="` + rel.Target + `" TargetMode="External"/>`
			}
		} else if strings.Contains(lt, "notesslide") {
			strXml += `<Relationship Id="rId` + itoa(rel.RID) + `" Target="` + rel.Target + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide"/>`
		}
	}
	for i := range slide.RelsChart {
		rel := &slide.RelsChart[i]
		if rel.RID > lastRid {
			lastRid = rel.RID
		}
		strXml += `<Relationship Id="rId` + itoa(rel.RID) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart" Target="` + rel.Target + `"/>`
	}
	for i := range slide.RelsMedia {
		rel := &slide.RelsMedia[i]
		relRid := itoa(rel.RID)
		if rel.RID > lastRid {
			lastRid = rel.RID
		}
		lt := strings.ToLower(rel.Type)
		if strings.Contains(lt, "image") {
			strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="` + rel.Target + `"/>`
		} else if strings.Contains(lt, "audio") {
			if strings.Contains(strXml, ` Target="`+rel.Target+`"`) {
				strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.microsoft.com/office/2007/relationships/media" Target="` + rel.Target + `"/>`
			} else {
				strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/audio" Target="` + rel.Target + `"/>`
			}
		} else if strings.Contains(lt, "video") {
			if strings.Contains(strXml, ` Target="`+rel.Target+`"`) {
				strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.microsoft.com/office/2007/relationships/media" Target="` + rel.Target + `"/>`
			} else {
				strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/video" Target="` + rel.Target + `"/>`
			}
		} else if strings.Contains(lt, "online") {
			if strings.Contains(strXml, ` Target="`+rel.Target+`"`) {
				strXml += `<Relationship Id="rId` + relRid + `" Type="http://schemas.microsoft.com/office/2007/relationships/image" Target="` + rel.Target + `"/>`
			} else {
				strXml += `<Relationship Id="rId` + relRid + `" Target="` + rel.Target + `" TargetMode="External" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/video"/>`
			}
		}
	}

	for idx, rel := range defaultRels {
		strXml += `<Relationship Id="rId` + itoa(lastRid+idx+1) + `" Type="` + rel.typ + `" Target="` + rel.target + `"/>`
	}

	strXml += "</Relationships>"
	return strXml
}

// ---------------------------------------------------------------------------
// genXmlParagraphProperties
// ---------------------------------------------------------------------------

func genXmlParagraphProperties(opts *ObjectOptions, isDefault bool) string {
	strXmlBullet := ""
	strXmlLnSpc := ""
	strXmlParaSpc := ""
	strXmlTabStops := ""
	tag := "a:pPr"
	if isDefault {
		tag = "a:lvl1pPr"
	}
	bulletMarL := valToPts(DEF_BULLET_MARGIN)

	rtl := ""
	if boolDeref(opts.RtlMode) {
		rtl = ` rtl="1" `
	}
	paragraphPropXml := "<" + tag + rtl

	// align
	switch opts.Align {
	case "left":
		paragraphPropXml += ` algn="l"`
	case "right":
		paragraphPropXml += ` algn="r"`
	case "center":
		paragraphPropXml += ` algn="ctr"`
	case "justify":
		paragraphPropXml += ` algn="just"`
	}

	if opts.LineSpacing != 0 {
		strXmlLnSpc = `<a:lnSpc><a:spcPts val="` + itoa(int(jsRound(opts.LineSpacing*100))) + `"/></a:lnSpc>`
	} else if opts.LineSpacingMultiple != 0 {
		strXmlLnSpc = `<a:lnSpc><a:spcPct val="` + itoa(int(jsRound(opts.LineSpacingMultiple*100000))) + `"/></a:lnSpc>`
	}

	if opts.IndentLevel > 0 {
		paragraphPropXml += ` lvl="` + itoa(opts.IndentLevel) + `"`
	}

	if opts.ParaSpaceBefore > 0 {
		strXmlParaSpc += `<a:spcBef><a:spcPts val="` + itoa(int(jsRound(opts.ParaSpaceBefore*100))) + `"/></a:spcBef>`
	}
	if opts.ParaSpaceAfter > 0 {
		strXmlParaSpc += `<a:spcAft><a:spcPts val="` + itoa(int(jsRound(opts.ParaSpaceAfter*100))) + `"/></a:spcAft>`
	}

	// bullet
	marLIndent := func() string {
		var marL int
		if opts.IndentLevel > 0 {
			marL = bulletMarL + bulletMarL*opts.IndentLevel
		} else {
			marL = bulletMarL
		}
		return ` marL="` + itoa(marL) + `" indent="-` + itoa(bulletMarL) + `"`
	}
	if opts.Bullet != nil {
		bl := opts.Bullet
		if bl.Indent != 0 {
			bulletMarL = valToPts(bl.Indent)
		}
		if bl.Type != "" {
			if strings.ToLower(bl.Type) == "number" {
				paragraphPropXml += marLIndent()
				startAt := bl.NumberStartAt
				if startAt == 0 {
					startAt = bl.StartAt
				}
				startStr := "1"
				if startAt != 0 {
					startStr = itoa(startAt)
				}
				strXmlBullet = `<a:buSzPct val="100000"/><a:buFont typeface="+mj-lt"/><a:buAutoNum type="` + strOr(bl.Style, "arabicPeriod") + `" startAt="` + startStr + `"/>`
			}
			// NOTE: bullet.type set but not "number" produces no bullet XML (TS quirk).
		} else if bl.CharacterCode != "" {
			bulletCode := "&#x" + bl.CharacterCode + ";"
			if !isHex4(bl.CharacterCode) {
				bulletCode = string(BulletTypeDefault)
			}
			paragraphPropXml += marLIndent()
			strXmlBullet = `<a:buSzPct val="100000"/><a:buChar char="` + bulletCode + `"/>`
		} else if bl.Code != "" {
			bulletCode := "&#x" + bl.Code + ";"
			if !isHex4(bl.Code) {
				bulletCode = string(BulletTypeDefault)
			}
			paragraphPropXml += marLIndent()
			strXmlBullet = `<a:buSzPct val="100000"/><a:buChar char="` + bulletCode + `"/>`
		} else {
			paragraphPropXml += marLIndent()
			strXmlBullet = `<a:buSzPct val="100000"/><a:buChar char="` + string(BulletTypeDefault) + `"/>`
		}
	} else {
		paragraphPropXml += ` indent="0" marL="0"`
		strXmlBullet = "<a:buNone/>"
	}

	// tabStops
	if len(opts.TabStops) > 0 {
		var ts strings.Builder
		for _, stop := range opts.TabStops {
			pos := stop.Position
			if pos == 0 {
				pos = 1
			}
			ts.WriteString(`<a:tab pos="` + itoa(inch2Emu(pos)) + `" algn="` + strOr(stop.Alignment, "l") + `"/>`)
		}
		strXmlTabStops = `<a:tabLst>` + ts.String() + `</a:tabLst>`
	}

	paragraphPropXml += ">" + strXmlLnSpc + strXmlParaSpc + strXmlBullet + strXmlTabStops
	if isDefault {
		paragraphPropXml += genXmlTextRunProperties(opts, true)
	}
	paragraphPropXml += "</" + tag + ">"

	return paragraphPropXml
}

// isHex4 reports whether s is exactly 4 hex digits.
func isHex4(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// genXmlTextRunProperties
// ---------------------------------------------------------------------------

func genXmlTextRunProperties(opts *ObjectOptions, isDefault bool) string {
	runProps := ""
	runPropsTag := "a:rPr"
	if isDefault {
		runPropsTag = "a:defRPr"
	}

	runProps += "<" + runPropsTag + ` lang="` + strOr(opts.Lang, "en-US") + `"`
	if opts.Lang != "" {
		runProps += ` altLang="en-US"`
	}
	if opts.FontSize != 0 {
		runProps += ` sz="` + itoa(int(jsRound(opts.FontSize*100))) + `"`
	}
	if boolDeref(opts.Bold) {
		runProps += ` b="1"`
	}
	if boolDeref(opts.Italic) {
		runProps += ` i="1"`
	}
	if opts.Strike != "" {
		runProps += ` strike="` + opts.Strike + `"`
	}
	if opts.Underline != nil && opts.Underline.Style != "" {
		runProps += ` u="` + opts.Underline.Style + `"`
	} else if opts.Hyperlink != nil {
		runProps += ` u="sng"`
	}
	if opts.Baseline != 0 {
		runProps += ` baseline="` + itoa(int(jsRound(opts.Baseline*50))) + `"`
	} else if boolDeref(opts.Subscript) {
		runProps += ` baseline="-40000"`
	} else if boolDeref(opts.Superscript) {
		runProps += ` baseline="30000"`
	}
	if opts.CharSpacing != 0 {
		runProps += ` spc="` + itoa(int(jsRound(opts.CharSpacing*100))) + `" kern="0"`
	}
	runProps += ` dirty="0">`

	// Children
	hasUnderlineColor := opts.Underline != nil && opts.Underline.Color != ""
	if opts.Color != "" || opts.FontFace != "" || opts.Outline != nil || hasUnderlineColor {
		if opts.Outline != nil {
			sz := opts.Outline.Size
			if sz == 0 {
				sz = 0.75
			}
			runProps += `<a:ln w="` + itoa(valToPts(sz)) + `">` + genXmlColorSelection(&ShapeFillProps{Color: strOr(opts.Outline.Color, "FFFFFF")}) + `</a:ln>`
		}
		if opts.Color != "" {
			runProps += genXmlColorSelection(&ShapeFillProps{Color: opts.Color, Transparency: opts.Transparency})
		}
		if opts.Highlight != "" {
			runProps += `<a:highlight>` + createColorElement(opts.Highlight, "") + `</a:highlight>`
		}
		if hasUnderlineColor {
			runProps += `<a:uFill>` + genXmlColorSelection(&ShapeFillProps{Color: opts.Underline.Color}) + `</a:uFill>`
		}
		if opts.Glow != nil {
			runProps += `<a:effectLst>` + createGlowElement(*opts.Glow, DEF_TEXT_GLOW) + `</a:effectLst>`
		}
		if opts.FontFace != "" {
			runProps += `<a:latin typeface="` + opts.FontFace + `" pitchFamily="34" charset="0"/><a:ea typeface="` + opts.FontFace + `" pitchFamily="34" charset="-122"/><a:cs typeface="` + opts.FontFace + `" pitchFamily="34" charset="-120"/>`
		}
	}

	// Hyperlink
	if opts.Hyperlink != nil {
		closeChar := "/>"
		if opts.Color != "" {
			closeChar = ">"
		}
		if opts.Hyperlink.URL != "" {
			runProps += `<a:hlinkClick r:id="rId` + itoa(opts.Hyperlink.RID) + `" invalidUrl="" action="" tgtFrame="" tooltip="` + tooltipVal(opts.Hyperlink) + `" history="1" highlightClick="0" endSnd="0"` + closeChar
		} else if opts.Hyperlink.Slide != 0 {
			runProps += `<a:hlinkClick r:id="rId` + itoa(opts.Hyperlink.RID) + `" action="ppaction://hlinksldjump" tooltip="` + tooltipVal(opts.Hyperlink) + `"` + closeChar
		}
		if opts.Color != "" {
			runProps += ` <a:extLst>`
			runProps += `  <a:ext uri="{A12FA001-AC4F-418D-AE19-62706E023703}">`
			runProps += `   <ahyp:hlinkClr xmlns:ahyp="http://schemas.microsoft.com/office/drawing/2018/hyperlinkcolor" val="tx"/>`
			runProps += `  </a:ext>`
			runProps += ` </a:extLst>`
			runProps += `</a:hlinkClick>`
		}
	}

	runProps += "</" + runPropsTag + ">"
	return runProps
}

// ---------------------------------------------------------------------------
// genXmlTextRun
// ---------------------------------------------------------------------------

func genXmlTextRun(text string, opts *ObjectOptions) string {
	if text == "" {
		return ""
	}
	return `<a:r>` + genXmlTextRunProperties(opts, false) + `<a:t>` + encodeXmlEntities(text) + `</a:t></a:r>`
}

// ---------------------------------------------------------------------------
// genXmlBodyProperties
// ---------------------------------------------------------------------------

func genXmlBodyProperties(typ SlideObjectType, opts *ObjectOptions) string {
	bodyProperties := "<a:bodyPr"

	if typ == SlideObjectTypeText && opts.BodyProp != nil {
		bp := opts.BodyProp
		if boolDeref(bp.Wrap) {
			bodyProperties += ` wrap="square"`
		} else {
			bodyProperties += ` wrap="none"`
		}
		// NOTE: TS emits these when the inset is any number incl. 0 (only skips
		// when undefined). BodyProps has no "unset" sentinel for float64, so we
		// emit on != 0; a caller-set inset of exactly 0 is therefore omitted.
		if bp.LIns != 0 {
			bodyProperties += ` lIns="` + ftoa(bp.LIns) + `"`
		}
		if bp.TIns != 0 {
			bodyProperties += ` tIns="` + ftoa(bp.TIns) + `"`
		}
		if bp.RIns != 0 {
			bodyProperties += ` rIns="` + ftoa(bp.RIns) + `"`
		}
		if bp.BIns != 0 {
			bodyProperties += ` bIns="` + ftoa(bp.BIns) + `"`
		}
		bodyProperties += ` rtlCol="0"`
		if bp.Anchor != "" {
			bodyProperties += ` anchor="` + string(bp.Anchor) + `"`
		}
		if bp.Vert != "" {
			bodyProperties += ` vert="` + bp.Vert + `"`
		}
		bodyProperties += ">"

		if opts.Fit != "" {
			switch opts.Fit {
			case "none":
				// nothing
			case "shrink":
				bodyProperties += "<a:normAutofit/>"
			case "resize":
				bodyProperties += "<a:spAutoFit/>"
			}
		}
		if boolDeref(opts.ShrinkText) {
			bodyProperties += "<a:normAutofit/>"
		}
		if boolDeref(bp.AutoFit) {
			bodyProperties += "<a:spAutoFit/>"
		}
		bodyProperties += "</a:bodyPr>"
	} else {
		bodyProperties += ` wrap="square" rtlCol="0">`
		bodyProperties += "</a:bodyPr>"
	}

	if typ == SlideObjectTypeTablecell {
		return "<a:bodyPr/>"
	}
	return bodyProperties
}

// ---------------------------------------------------------------------------
// genXmlTextBody
// ---------------------------------------------------------------------------

// genXmlTextBody renders the <p:txBody> for a slide object (text/placeholder),
// or returns "" for a shape with no text. Table cells go through the internal
// genXmlTextBodyCore path from the table renderer.
func genXmlTextBody(slideObj *SlideObject) string {
	var opts ObjectOptions
	if slideObj.Options != nil {
		opts = *slideObj.Options // shallow copy so we don't mutate the model
	}
	// Shapes with no text render nothing.
	if slideObj.Type != SlideObjectTypeTablecell && slideObj.Text == nil {
		return ""
	}
	var tmp []textRun
	for i := range slideObj.Text {
		tmp = append(tmp, textRun{text: slideObj.Text[i].Text, options: objOptsFromTextProps(slideObj.Text[i].Options)})
	}
	return genXmlTextBodyCore(slideObj.Type, &opts, tmp, slideObj.Type == SlideObjectTypePlaceholder)
}

// textRun is a normalized (text, options) pair used internally.
type textRun struct {
	text    string
	options *ObjectOptions
}

// genXmlTextBodyCore builds a txBody from a normalized run list.
func genXmlTextBodyCore(typ SlideObjectType, opts *ObjectOptions, tmpTextObjects []textRun, isPlaceholder bool) string {
	// STEP 1: start
	strSlideXml := "<p:txBody>"
	if typ == SlideObjectTypeTablecell {
		strSlideXml = "<a:txBody>"
	}

	// STEP 2: bodyPr + lstStyle
	strSlideXml += genXmlBodyProperties(typ, opts)
	if opts.H != nil && opts.H.Val == 0 && !opts.H.IsPct && opts.Line != nil && opts.Align != "" {
		strSlideXml += `<a:lstStyle><a:lvl1pPr algn="l"/></a:lstStyle>`
	} else if isPlaceholder {
		strSlideXml += `<a:lstStyle>` + genXmlParagraphProperties(opts, true) + `</a:lstStyle>`
	} else {
		strSlideXml += `<a:lstStyle/>`
	}

	// STEP 4: normalize text (line-break splitting)
	var arrTextObjects []textRun
	for idx := range tmpTextObjects {
		itext := &tmpTextObjects[idx]
		if itext.options == nil {
			itext.options = &ObjectOptions{}
		}
		if idx == 0 && itext.options.Bullet == nil && opts.Bullet != nil {
			itext.options.Bullet = opts.Bullet
		}
		// Convert \r\n / \n into CRLF
		itext.text = normalizeCRLF(itext.text)

		if strings.Contains(itext.text, CRLF) && !strings.HasSuffix(itext.text, "\n") {
			for _, line := range strings.Split(itext.text, CRLF) {
				bl := true
				o := cloneObjOpts(itext.options)
				o.BreakLine = &bl
				arrTextObjects = append(arrTextObjects, textRun{text: line, options: o})
			}
		} else {
			arrTextObjects = append(arrTextObjects, *itext)
		}
	}

	// STEP 5: group into lines
	var arrLines [][]textRun
	var arrTexts []textRun
	for idx := range arrTextObjects {
		textObj := &arrTextObjects[idx]
		if len(arrTexts) > 0 && (textObj.options.Align != "" || opts.Align != "") {
			if textObj.options.Align != arrTextObjects[idx-1].options.Align {
				arrLines = append(arrLines, arrTexts)
				arrTexts = nil
			}
		} else if len(arrTexts) > 0 && textObj.options.Bullet != nil {
			arrLines = append(arrLines, arrTexts)
			arrTexts = nil
			f := false
			textObj.options.BreakLine = &f
		}

		arrTexts = append(arrTexts, *textObj)

		if len(arrTexts) > 0 && boolDeref(textObj.options.BreakLine) {
			if idx+1 < len(arrTextObjects) {
				arrLines = append(arrLines, arrTexts)
				arrTexts = nil
			}
		}
		if idx+1 == len(arrTextObjects) {
			arrLines = append(arrLines, arrTexts)
		}
	}

	// STEP 6: render lines
	for _, line := range arrLines {
		reqsClosingFontSize := false
		strSlideXml += "<a:p>"

		for idx := range line {
			textObj := &line[idx]
			textObj.options.LineIdx = idx

			if idx > 0 && boolDeref(textObj.options.SoftBreakBefore) {
				strSlideXml += "<a:br/>"
			}

			// Inherit pPr-type options from parent shape
			if textObj.options.Align == "" {
				textObj.options.Align = opts.Align
			}
			if textObj.options.LineSpacing == 0 {
				textObj.options.LineSpacing = opts.LineSpacing
			}
			if textObj.options.LineSpacingMultiple == 0 {
				textObj.options.LineSpacingMultiple = opts.LineSpacingMultiple
			}
			if textObj.options.IndentLevel == 0 {
				textObj.options.IndentLevel = opts.IndentLevel
			}
			if textObj.options.ParaSpaceBefore == 0 {
				textObj.options.ParaSpaceBefore = opts.ParaSpaceBefore
			}
			if textObj.options.ParaSpaceAfter == 0 {
				textObj.options.ParaSpaceAfter = opts.ParaSpaceAfter
			}
			paragraphPropXml := genXmlParagraphProperties(textObj.options, false)
			strSlideXml += strings.ReplaceAll(paragraphPropXml, "<a:pPr></a:pPr>", "")

			// Inherit run-level options from shape opts (except bullet; and color
			// is not inherited when the run has a hyperlink).
			inheritRunOptions(textObj.options, opts)

			strSlideXml += genXmlTextRun(textObj.text, textObj.options)

			if (textObj.text == "" && opts.FontSize != 0) || textObj.options.FontSize != 0 {
				reqsClosingFontSize = true
				if opts.FontSize == 0 {
					opts.FontSize = textObj.options.FontSize
				}
			}
		}

		if typ == SlideObjectTypeTablecell && (opts.FontSize != 0 || opts.FontFace != "") {
			if opts.FontFace != "" {
				strSlideXml += `<a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"`
				if opts.FontSize != 0 {
					strSlideXml += ` sz="` + itoa(int(jsRound(opts.FontSize*100))) + `"`
				}
				strSlideXml += ` dirty="0">`
				strSlideXml += `<a:latin typeface="` + opts.FontFace + `" charset="0"/>`
				strSlideXml += `<a:ea typeface="` + opts.FontFace + `" charset="0"/>`
				strSlideXml += `<a:cs typeface="` + opts.FontFace + `" charset="0"/>`
				strSlideXml += `</a:endParaRPr>`
			} else {
				strSlideXml += `<a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"`
				if opts.FontSize != 0 {
					strSlideXml += ` sz="` + itoa(int(jsRound(opts.FontSize*100))) + `"`
				}
				strSlideXml += ` dirty="0"/>`
			}
		} else if reqsClosingFontSize {
			strSlideXml += `<a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `"`
			if opts.FontSize != 0 {
				strSlideXml += ` sz="` + itoa(int(jsRound(opts.FontSize*100))) + `"`
			}
			strSlideXml += ` dirty="0"/>`
		} else {
			strSlideXml += `<a:endParaRPr lang="` + strOr(opts.Lang, "en-US") + `" dirty="0"/>`
		}

		strSlideXml += "</a:p>"
	}

	if !strings.Contains(strSlideXml, "<a:p>") {
		strSlideXml += "<a:p><a:endParaRPr/></a:p>"
	}

	if typ == SlideObjectTypeTablecell {
		strSlideXml += "</a:txBody>"
	} else {
		strSlideXml += "</p:txBody>"
	}

	return strSlideXml
}

// normalizeCRLF replaces \r*\n sequences with CRLF (mirrors /\r*\n/g → CRLF).
func normalizeCRLF(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			b.WriteString(CRLF)
		} else if s[i] == '\r' {
			// peek: collapse \r+\n; a lone \r is left as-is (JS regex requires \n)
			j := i
			for j < len(s) && s[j] == '\r' {
				j++
			}
			if j < len(s) && s[j] == '\n' {
				b.WriteString(CRLF)
				i = j // skip to the \n (loop ++ moves past it)
			} else {
				b.WriteByte('\r')
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// cloneObjOpts makes a shallow copy of an ObjectOptions.
func cloneObjOpts(o *ObjectOptions) *ObjectOptions {
	if o == nil {
		return &ObjectOptions{}
	}
	c := *o
	return &c
}

// inheritRunOptions copies run-relevant fields from the shape options to the run
// options when the run's field is unset. Mirrors the TS Object.entries(opts)
// inheritance loop (bullet excluded; color skipped when run has a hyperlink).
func inheritRunOptions(run, opts *ObjectOptions) {
	hasHyperlink := run.Hyperlink != nil
	if run.Lang == "" {
		run.Lang = opts.Lang
	}
	if run.FontSize == 0 {
		run.FontSize = opts.FontSize
	}
	if run.FontFace == "" {
		run.FontFace = opts.FontFace
	}
	if run.Bold == nil {
		run.Bold = opts.Bold
	}
	if run.Italic == nil {
		run.Italic = opts.Italic
	}
	if run.Strike == "" {
		run.Strike = opts.Strike
	}
	if run.Underline == nil {
		run.Underline = opts.Underline
	}
	if run.Baseline == 0 {
		run.Baseline = opts.Baseline
	}
	if run.Subscript == nil {
		run.Subscript = opts.Subscript
	}
	if run.Superscript == nil {
		run.Superscript = opts.Superscript
	}
	if run.CharSpacing == 0 {
		run.CharSpacing = opts.CharSpacing
	}
	if !hasHyperlink && run.Color == "" {
		run.Color = opts.Color
	}
	if run.Transparency == 0 {
		run.Transparency = opts.Transparency
	}
	if run.Highlight == "" {
		run.Highlight = opts.Highlight
	}
	if run.Outline == nil {
		run.Outline = opts.Outline
	}
	if run.Glow == nil {
		run.Glow = opts.Glow
	}
	if run.Hyperlink == nil {
		run.Hyperlink = opts.Hyperlink
	}
	if run.RtlMode == nil {
		run.RtlMode = opts.RtlMode
	}
	if run.TabStops == nil {
		run.TabStops = opts.TabStops
	}
}

// objOptsFromTextProps converts a run's *TextPropsOptions to *ObjectOptions
// (ObjectOptions is a superset of the fields used by the XML builders).
func objOptsFromTextProps(tp *TextPropsOptions) *ObjectOptions {
	o := &ObjectOptions{}
	if tp == nil {
		return o
	}
	o.PositionProps = tp.PositionProps
	o.DataOrPathProps = tp.DataOrPathProps
	o.TextBaseProps = tp.TextBaseProps
	o.ObjectNameProps = tp.ObjectNameProps
	o.BodyProp = tp.BodyProp
	o.LineIdx = tp.LineIdx
	o.Baseline = tp.Baseline
	o.CharSpacing = tp.CharSpacing
	o.Fit = tp.Fit
	o.Fill = tp.Fill
	o.FlipH = tp.FlipH
	o.FlipV = tp.FlipV
	o.Glow = tp.Glow
	o.Hyperlink = tp.Hyperlink
	o.IndentLevel = tp.IndentLevel
	o.IsTextBox = tp.IsTextBox
	o.Line = tp.Line
	o.LineSpacing = tp.LineSpacing
	o.LineSpacingMultiple = tp.LineSpacingMultiple
	o.Margin = tp.Margin
	o.Outline = tp.Outline
	o.ParaSpaceAfter = tp.ParaSpaceAfter
	o.ParaSpaceBefore = tp.ParaSpaceBefore
	o.RectRadius = tp.RectRadius
	o.Rotate = tp.Rotate
	o.RtlMode = tp.RtlMode
	o.Shadow = tp.Shadow
	o.Shape = tp.Shape
	o.Strike = tp.Strike
	o.Subscript = tp.Subscript
	o.Superscript = tp.Superscript
	o.Vert = tp.Vert
	o.Wrap = tp.Wrap
	o.AutoFit = tp.AutoFit
	o.ShrinkText = tp.ShrinkText
	o.Inset = tp.Inset
	return o
}

// ---------------------------------------------------------------------------
// genXmlPlaceholder
// ---------------------------------------------------------------------------

// placeholderTypesMap mirrors the TS PLACEHOLDER_TYPES string enum
// (key → XML value). Only these keys are truthy in the double-lookup below.
var placeholderTypesMap = map[string]string{
	"title": "title",
	"body":  "body",
	"image": "pic",
	"chart": "chart",
	"table": "tbl",
	"media": "media",
}

func genXmlPlaceholder(placeholderObj *SlideObject) string {
	if placeholderObj == nil {
		return ""
	}

	placeholderIdx := ""
	placeholderTyp := ""
	if placeholderObj.Options != nil {
		if placeholderObj.Options.PlaceholderIdx != 0 {
			placeholderIdx = itoa(placeholderObj.Options.PlaceholderIdx)
		}
		placeholderTyp = string(placeholderObj.Options.PlaceholderType)
	}

	placeholderType := ""
	if placeholderTyp != "" {
		if v, ok := placeholderTypesMap[placeholderTyp]; ok {
			placeholderType = v
		}
	}

	idxAttr := ""
	if placeholderIdx != "" {
		idxAttr = ` idx="` + placeholderIdx + `"`
	}
	typeAttr := ""
	if placeholderType != "" {
		if _, ok := placeholderTypesMap[placeholderType]; ok {
			typeAttr = ` type="` + placeholderType + `"`
		}
	}
	customPrompt := ""
	if len(placeholderObj.Text) > 0 {
		customPrompt = ` hasCustomPrompt="1"`
	}

	return "<p:ph\n\t\t" + idxAttr + "\n\t\t" + typeAttr + "\n\t\t" + customPrompt + "\n\t\t/>"
}

// ---------------------------------------------------------------------------
// makeXmlContTypes — [Content_Types].xml
// ---------------------------------------------------------------------------

func makeXmlContTypes(slides []PresSlide, slideLayouts []SlideLayout, masterSlide *PresSlide, embeddedFonts []*EmbeddedFont) string {
	strXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF
	strXml += `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`
	strXml += `<Default Extension="xml" ContentType="application/xml"/>`
	strXml += `<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`
	strXml += `<Default Extension="jpeg" ContentType="image/jpeg"/>`
	strXml += `<Default Extension="jpg" ContentType="image/jpg"/>`
	strXml += `<Default Extension="svg" ContentType="image/svg+xml"/>`

	strXml += `<Default Extension="png" ContentType="image/png"/>`
	strXml += `<Default Extension="gif" ContentType="image/gif"/>`
	strXml += `<Default Extension="m4v" ContentType="video/mp4"/>`
	strXml += `<Default Extension="mp4" ContentType="video/mp4"/>`
	for si := range slides {
		for ri := range slides[si].RelsMedia {
			rel := &slides[si].RelsMedia[ri]
			if rel.Type != "image" && rel.Type != "online" && rel.Type != "chart" && rel.Extn != "m4v" && !strings.Contains(strXml, rel.Type) {
				strXml += `<Default Extension="` + rel.Extn + `" ContentType="` + rel.Type + `"/>`
			}
		}
	}
	strXml += `<Default Extension="vml" ContentType="application/vnd.openxmlformats-officedocument.vmlDrawing"/>`
	strXml += `<Default Extension="xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>`

	// Font embedding (net-new): declare fntdata default when fonts are embedded.
	if len(embeddedFonts) > 0 {
		strXml += `<Default Extension="fntdata" ContentType="application/x-fontdata"/>`
	}

	strXml += `<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>`
	strXml += `<Override PartName="/ppt/notesMasters/notesMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.notesMaster+xml"/>`
	for idx := range slides {
		strXml += `<Override PartName="/ppt/slideMasters/slideMaster` + itoa(idx+1) + `.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>`
		strXml += `<Override PartName="/ppt/slides/slide` + itoa(idx+1) + `.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`
		for ri := range slides[idx].RelsChart {
			strXml += `<Override PartName="` + slides[idx].RelsChart[ri].Target + `" ContentType="application/vnd.openxmlformats-officedocument.drawingml.chart+xml"/>`
		}
	}

	strXml += `<Override PartName="/ppt/presProps.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presProps+xml"/>`
	strXml += `<Override PartName="/ppt/viewProps.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.viewProps+xml"/>`
	strXml += `<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>`
	strXml += `<Override PartName="/ppt/tableStyles.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.tableStyles+xml"/>`

	for idx := range slideLayouts {
		strXml += `<Override PartName="/ppt/slideLayouts/slideLayout` + itoa(idx+1) + `.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>`
		for ri := range slideLayouts[idx].RelsChart {
			strXml += ` <Override PartName="` + slideLayouts[idx].RelsChart[ri].Target + `" ContentType="application/vnd.openxmlformats-officedocument.drawingml.chart+xml"/>`
		}
	}

	for idx := range slides {
		strXml += `<Override PartName="/ppt/notesSlides/notesSlide` + itoa(idx+1) + `.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.notesSlide+xml"/>`
	}

	if masterSlide != nil {
		for ri := range masterSlide.RelsChart {
			strXml += ` <Override PartName="` + masterSlide.RelsChart[ri].Target + `" ContentType="application/vnd.openxmlformats-officedocument.drawingml.chart+xml"/>`
		}
		for ri := range masterSlide.RelsMedia {
			rel := &masterSlide.RelsMedia[ri]
			if rel.Type != "image" && rel.Type != "online" && rel.Type != "chart" && rel.Extn != "m4v" && !strings.Contains(strXml, rel.Type) {
				strXml += ` <Default Extension="` + rel.Extn + `" ContentType="` + rel.Type + `"/>`
			}
		}
	}

	strXml += ` <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`
	strXml += ` <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`
	strXml += `</Types>`

	return strXml
}

// ---------------------------------------------------------------------------
// makeXmlRootRels — _rels/.rels
// ---------------------------------------------------------------------------

func makeXmlRootRels() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
		<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
		<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
		<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
		</Relationships>`
}

// ---------------------------------------------------------------------------
// makeXmlApp — docProps/app.xml
// ---------------------------------------------------------------------------

func makeXmlApp(slides []PresSlide, company string) string {
	n := len(slides)
	var slideTitles strings.Builder
	for idx := range slides {
		slideTitles.WriteString(`<vt:lpstr>Slide ` + itoa(idx+1) + `</vt:lpstr>`)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
	<TotalTime>0</TotalTime>
	<Words>0</Words>
	<Application>Microsoft Office PowerPoint</Application>
	<PresentationFormat>On-screen Show (16:9)</PresentationFormat>
	<Paragraphs>0</Paragraphs>
	<Slides>` + itoa(n) + `</Slides>
	<Notes>` + itoa(n) + `</Notes>
	<HiddenSlides>0</HiddenSlides>
	<MMClips>0</MMClips>
	<ScaleCrop>false</ScaleCrop>
	<HeadingPairs>
		<vt:vector size="6" baseType="variant">
			<vt:variant><vt:lpstr>Fonts Used</vt:lpstr></vt:variant>
			<vt:variant><vt:i4>2</vt:i4></vt:variant>
			<vt:variant><vt:lpstr>Theme</vt:lpstr></vt:variant>
			<vt:variant><vt:i4>1</vt:i4></vt:variant>
			<vt:variant><vt:lpstr>Slide Titles</vt:lpstr></vt:variant>
			<vt:variant><vt:i4>` + itoa(n) + `</vt:i4></vt:variant>
		</vt:vector>
	</HeadingPairs>
	<TitlesOfParts>
		<vt:vector size="` + itoa(n+1+2) + `" baseType="lpstr">
			<vt:lpstr>Arial</vt:lpstr>
			<vt:lpstr>Calibri</vt:lpstr>
			<vt:lpstr>Office Theme</vt:lpstr>
			` + slideTitles.String() + `
		</vt:vector>
	</TitlesOfParts>
	<Company>` + company + `</Company>
	<LinksUpToDate>false</LinksUpToDate>
	<SharedDoc>false</SharedDoc>
	<HyperlinksChanged>false</HyperlinksChanged>
	<AppVersion>16.0000</AppVersion>
	</Properties>`
}

// ---------------------------------------------------------------------------
// makeXmlCore — docProps/core.xml
//
// The created/modified timestamps use xmlNowFunc (overridable in tests), and the
// `.\d\d\dZ → Z` truncation of the JS ISO string is applied via w3cdtf().
// ---------------------------------------------------------------------------

func makeXmlCore(title, subject, author, revision string) string {
	ts := w3cdtf(xmlNowFunc())
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<dc:title>` + encodeXmlEntities(title) + `</dc:title>
		<dc:subject>` + encodeXmlEntities(subject) + `</dc:subject>
		<dc:creator>` + encodeXmlEntities(author) + `</dc:creator>
		<cp:lastModifiedBy>` + encodeXmlEntities(author) + `</cp:lastModifiedBy>
		<cp:revision>` + revision + `</cp:revision>
		<dcterms:created xsi:type="dcterms:W3CDTF">` + ts + `</dcterms:created>
		<dcterms:modified xsi:type="dcterms:W3CDTF">` + ts + `</dcterms:modified>
	</cp:coreProperties>`
}

// w3cdtf formats t as JS `new Date().toISOString().replace(/\.\d\d\dZ/, 'Z')` —
// UTC, second precision, e.g. "2026-07-20T04:28:22Z".
func w3cdtf(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

// ---------------------------------------------------------------------------
// makeXmlPresentationRels — ppt/_rels/presentation.xml.rels
// ---------------------------------------------------------------------------

func makeXmlPresentationRels(slides []PresSlide, embeddedFonts []*EmbeddedFont) string {
	intRelNum := 1
	strXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF
	strXml += `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`
	strXml += `<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>`
	for idx := 1; idx <= len(slides); idx++ {
		intRelNum++
		strXml += `<Relationship Id="rId` + itoa(intRelNum) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide` + itoa(idx) + `.xml"/>`
	}
	intRelNum++
	strXml += `<Relationship Id="rId` + itoa(intRelNum+0) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesMaster" Target="notesMasters/notesMaster1.xml"/>` +
		`<Relationship Id="rId` + itoa(intRelNum+1) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" Target="presProps.xml"/>` +
		`<Relationship Id="rId` + itoa(intRelNum+2) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/viewProps" Target="viewProps.xml"/>` +
		`<Relationship Id="rId` + itoa(intRelNum+3) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>` +
		`<Relationship Id="rId` + itoa(intRelNum+4) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/tableStyles" Target="tableStyles.xml"/>`

	// Font embedding (net-new): append one Relationship per font variant. rIds
	// continue after tableStyles; see embeddedFontVariants for the numbering.
	for _, v := range embeddedFontVariants(len(slides), embeddedFonts) {
		strXml += `<Relationship Id="rId` + itoa(v.rID) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/font" Target="fonts/font` + itoa(v.fileNum) + `.fntdata"/>`
	}

	strXml += `</Relationships>`
	return strXml
}

// fontVariantRel is one embedded-font relationship (typeface+style → rId/file#).
type fontVariantRel struct {
	typeface string
	style    FontStyle
	rID      int
	fileNum  int
}

// embeddedFontVariants enumerates all present font variants across the given
// typefaces (typeface outer, fontStyleOrder inner) and assigns each a
// sequential file number (1-based) and relationship id.
//
// Rel-numbering scheme: makeXmlPresentationRels emits 1 slideMaster rel (rId1),
// len(slides) slide rels, then 5 tail rels (notesMaster, presProps, viewProps,
// theme, tableStyles). The highest existing rId is therefore len(slides)+6.
// Font rels start at len(slides)+7; variant j (0-based) → rId len(slides)+7+j,
// file number j+1.
func embeddedFontVariants(numSlides int, fonts []*EmbeddedFont) []fontVariantRel {
	base := numSlides + 6
	var out []fontVariantRel
	counter := 0
	for _, f := range fonts {
		if f == nil {
			continue
		}
		for _, st := range fontStyleOrder {
			if _, ok := f.Variants[st]; ok {
				counter++
				out = append(out, fontVariantRel{typeface: f.Typeface, style: st, rID: base + counter, fileNum: counter})
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// makeXmlSlide — ppt/slides/slideN.xml
// ---------------------------------------------------------------------------

func makeXmlSlide(slide *PresSlide) string {
	show := ""
	if boolDeref(slide.Hidden) {
		show = ` show="0"`
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF +
		`<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" ` +
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"` +
		show + `>` +
		slideObjectToXml(&slide.SlideBaseProps, &slide.SlideLayout) +
		`<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>`
}

// ---------------------------------------------------------------------------
// getNotesFromSlide
// ---------------------------------------------------------------------------

func getNotesFromSlide(slide *PresSlide) string {
	notesText := ""
	for i := range slide.SlideObjects {
		data := &slide.SlideObjects[i]
		if data.Type == SlideObjectTypeNotes {
			if len(data.Text) > 0 {
				notesText += data.Text[0].Text
			}
		}
	}
	return normalizeCRLF(notesText)
}

// ---------------------------------------------------------------------------
// makeXmlNotesMaster — ppt/notesMasters/notesMaster1.xml
// ---------------------------------------------------------------------------

func makeXmlNotesMaster() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<p:notesMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:cSld><p:bg><p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef></p:bg><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr><p:sp><p:nvSpPr><p:cNvPr id="2" name="Header Placeholder 1"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="hdr" sz="quarter"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="2971800" cy="458788"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0"/><a:lstStyle><a:lvl1pPr algn="l"><a:defRPr sz="1200"/></a:lvl1pPr></a:lstStyle><a:p><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="3" name="Date Placeholder 2"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="dt" idx="1"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="3884613" y="0"/><a:ext cx="2971800" cy="458788"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0"/><a:lstStyle><a:lvl1pPr algn="r"><a:defRPr sz="1200"/></a:lvl1pPr></a:lstStyle><a:p><a:fld id="{5282F153-3F37-0F45-9E97-73ACFA13230C}" type="datetimeFigureOut"><a:rPr lang="en-US"/><a:t>7/23/19</a:t></a:fld><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="4" name="Slide Image Placeholder 3"/><p:cNvSpPr><a:spLocks noGrp="1" noRot="1" noChangeAspect="1"/></p:cNvSpPr><p:nvPr><p:ph type="sldImg" idx="2"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="685800" y="1143000"/><a:ext cx="5486400" cy="3086100"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln w="12700"><a:solidFill><a:prstClr val="black"/></a:solidFill></a:ln></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0" anchor="ctr"/><a:lstStyle/><a:p><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="5" name="Notes Placeholder 4"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="body" sz="quarter" idx="3"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="685800" y="4400550"/><a:ext cx="5486400" cy="3600450"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0"/><a:lstStyle/><a:p><a:pPr lvl="0"/><a:r><a:rPr lang="en-US"/><a:t>Click to edit Master text styles</a:t></a:r></a:p><a:p><a:pPr lvl="1"/><a:r><a:rPr lang="en-US"/><a:t>Second level</a:t></a:r></a:p><a:p><a:pPr lvl="2"/><a:r><a:rPr lang="en-US"/><a:t>Third level</a:t></a:r></a:p><a:p><a:pPr lvl="3"/><a:r><a:rPr lang="en-US"/><a:t>Fourth level</a:t></a:r></a:p><a:p><a:pPr lvl="4"/><a:r><a:rPr lang="en-US"/><a:t>Fifth level</a:t></a:r></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="6" name="Footer Placeholder 5"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="ftr" sz="quarter" idx="4"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="0" y="8685213"/><a:ext cx="2971800" cy="458787"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0" anchor="b"/><a:lstStyle><a:lvl1pPr algn="l"><a:defRPr sz="1200"/></a:lvl1pPr></a:lstStyle><a:p><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="7" name="Slide Number Placeholder 6"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="sldNum" sz="quarter" idx="5"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="3884613" y="8685213"/><a:ext cx="2971800" cy="458787"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr><p:txBody><a:bodyPr vert="horz" lIns="91440" tIns="45720" rIns="91440" bIns="45720" rtlCol="0" anchor="b"/><a:lstStyle><a:lvl1pPr algn="r"><a:defRPr sz="1200"/></a:lvl1pPr></a:lstStyle><a:p><a:fld id="{CE5E9CC1-C706-0F49-92D6-E571CC5EEA8F}" type="slidenum"><a:rPr lang="en-US"/><a:t>‹#›</a:t></a:fld><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp></p:spTree><p:extLst><p:ext uri="{BB962C8B-B14F-4D97-AF65-F5344CB8AC3E}"><p14:creationId xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" val="1024086991"/></p:ext></p:extLst></p:cSld><p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/><p:notesStyle><a:lvl1pPr marL="0" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl1pPr><a:lvl2pPr marL="457200" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl2pPr><a:lvl3pPr marL="914400" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl3pPr><a:lvl4pPr marL="1371600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl4pPr><a:lvl5pPr marL="1828800" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl5pPr><a:lvl6pPr marL="2286000" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl6pPr><a:lvl7pPr marL="2743200" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl7pPr><a:lvl8pPr marL="3200400" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl8pPr><a:lvl9pPr marL="3657600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl9pPr></p:notesStyle></p:notesMaster>`
}

// ---------------------------------------------------------------------------
// makeXmlNotesSlide — ppt/notesSlides/notesSlideN.xml
// ---------------------------------------------------------------------------

func makeXmlNotesSlide(slide *PresSlide) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<p:notes xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr><p:sp><p:nvSpPr><p:cNvPr id="2" name="Slide Image Placeholder 1"/><p:cNvSpPr><a:spLocks noGrp="1" noRot="1" noChangeAspect="1"/></p:cNvSpPr><p:nvPr><p:ph type="sldImg"/></p:nvPr></p:nvSpPr><p:spPr/></p:sp><p:sp><p:nvSpPr><p:cNvPr id="3" name="Notes Placeholder 2"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="body" idx="1"/></p:nvPr></p:nvSpPr><p:spPr/><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr lang="en-US" dirty="0"/><a:t>` + encodeXmlEntities(getNotesFromSlide(slide)) + `</a:t></a:r><a:endParaRPr lang="en-US" dirty="0"/></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="4" name="Slide Number Placeholder 3"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="sldNum" sz="quarter" idx="10"/></p:nvPr></p:nvSpPr><p:spPr/><p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:fld id="` + SLDNUMFLDID + `" type="slidenum"><a:rPr lang="en-US"/><a:t>` + itoa(slide.SlideNum) + `</a:t></a:fld><a:endParaRPr lang="en-US"/></a:p></p:txBody></p:sp></p:spTree><p:extLst><p:ext uri="{BB962C8B-B14F-4D97-AF65-F5344CB8AC3E}"><p14:creationId xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" val="1024086991"/></p:ext></p:extLst></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:notes>`
}

// ---------------------------------------------------------------------------
// makeXmlLayout — ppt/slideLayouts/slideLayoutN.xml
// ---------------------------------------------------------------------------

func makeXmlLayout(layout *SlideLayout) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" preserve="1">
		` + slideObjectToXml(&layout.SlideBaseProps, nil) + `
		<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sldLayout>`
}

// ---------------------------------------------------------------------------
// makeXmlMaster — ppt/slideMasters/slideMaster1.xml
// ---------------------------------------------------------------------------

func makeXmlMaster(slide *PresSlide, layouts []SlideLayout) string {
	var layoutDefs strings.Builder
	for idx := range layouts {
		layoutDefs.WriteString(`<p:sldLayoutId id="` + itoa(LAYOUT_IDX_SERIES_BASE+idx) + `" r:id="rId` + itoa(len(slide.Rels)+idx+1) + `"/>`)
	}

	strXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF
	strXml += `<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">`
	strXml += slideObjectToXml(&slide.SlideBaseProps, nil)
	strXml += `<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>`
	strXml += `<p:sldLayoutIdLst>` + layoutDefs.String() + `</p:sldLayoutIdLst>`
	strXml += `<p:hf sldNum="0" hdr="0" ftr="0" dt="0"/>`
	strXml += `<p:txStyles>` +
		` <p:titleStyle>` +
		`  <a:lvl1pPr algn="ctr" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="0"/></a:spcBef><a:buNone/><a:defRPr sz="4400" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mj-lt"/><a:ea typeface="+mj-ea"/><a:cs typeface="+mj-cs"/></a:defRPr></a:lvl1pPr>` +
		` </p:titleStyle>` +
		` <p:bodyStyle>` +
		`  <a:lvl1pPr marL="342900" indent="-342900" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="3200" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl1pPr>` +
		`  <a:lvl2pPr marL="742950" indent="-285750" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="–"/><a:defRPr sz="2800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl2pPr>` +
		`  <a:lvl3pPr marL="1143000" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="2400" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl3pPr>` +
		`  <a:lvl4pPr marL="1600200" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="–"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl4pPr>` +
		`  <a:lvl5pPr marL="2057400" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="»"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl5pPr>` +
		`  <a:lvl6pPr marL="2514600" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl6pPr>` +
		`  <a:lvl7pPr marL="2971800" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl7pPr>` +
		`  <a:lvl8pPr marL="3429000" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl8pPr>` +
		`  <a:lvl9pPr marL="3886200" indent="-228600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:spcBef><a:spcPct val="20000"/></a:spcBef><a:buFont typeface="Arial" pitchFamily="34" charset="0"/><a:buChar char="•"/><a:defRPr sz="2000" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl9pPr>` +
		` </p:bodyStyle>` +
		` <p:otherStyle>` +
		`  <a:defPPr><a:defRPr lang="en-US"/></a:defPPr>` +
		`  <a:lvl1pPr marL="0" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl1pPr>` +
		`  <a:lvl2pPr marL="457200" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl2pPr>` +
		`  <a:lvl3pPr marL="914400" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl3pPr>` +
		`  <a:lvl4pPr marL="1371600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl4pPr>` +
		`  <a:lvl5pPr marL="1828800" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl5pPr>` +
		`  <a:lvl6pPr marL="2286000" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl6pPr>` +
		`  <a:lvl7pPr marL="2743200" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl7pPr>` +
		`  <a:lvl8pPr marL="3200400" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl8pPr>` +
		`  <a:lvl9pPr marL="3657600" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1"><a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/></a:defRPr></a:lvl9pPr>` +
		` </p:otherStyle>` +
		`</p:txStyles>`
	strXml += `</p:sldMaster>`

	return strXml
}

// ---------------------------------------------------------------------------
// makeXmlSlideLayoutRel — ppt/slideLayouts/_rels/slideLayoutN.xml.rels
// ---------------------------------------------------------------------------

func makeXmlSlideLayoutRel(layoutNumber int, slideLayouts []SlideLayout) string {
	return slideObjectRelationsToXml(&slideLayouts[layoutNumber-1].SlideBaseProps, []defaultRel{
		{
			target: "../slideMasters/slideMaster1.xml",
			typ:    "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster",
		},
	})
}

// ---------------------------------------------------------------------------
// makeXmlSlideRel — ppt/slides/_rels/slideN.xml.rels
// ---------------------------------------------------------------------------

func makeXmlSlideRel(slides []PresSlide, slideLayouts []SlideLayout, slideNumber int) string {
	return slideObjectRelationsToXml(&slides[slideNumber-1].SlideBaseProps, []defaultRel{
		{
			target: `../slideLayouts/slideLayout` + itoa(getLayoutIdxForSlide(slides, slideLayouts, slideNumber)) + `.xml`,
			typ:    "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		},
		{
			target: `../notesSlides/notesSlide` + itoa(slideNumber) + `.xml`,
			typ:    "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide",
		},
	})
}

// ---------------------------------------------------------------------------
// makeXmlNotesSlideRel — ppt/notesSlides/_rels/notesSlideN.xml.rels
// ---------------------------------------------------------------------------

func makeXmlNotesSlideRel(slideNumber int) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
			<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesMaster" Target="../notesMasters/notesMaster1.xml"/>
			<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="../slides/slide` + itoa(slideNumber) + `.xml"/>
		</Relationships>`
}

// ---------------------------------------------------------------------------
// makeXmlMasterRel — ppt/slideMasters/_rels/slideMaster1.xml.rels
// ---------------------------------------------------------------------------

func makeXmlMasterRel(masterSlide *PresSlide, slideLayouts []SlideLayout) string {
	defaultRels := make([]defaultRel, 0, len(slideLayouts)+1)
	for idx := range slideLayouts {
		defaultRels = append(defaultRels, defaultRel{
			target: `../slideLayouts/slideLayout` + itoa(idx+1) + `.xml`,
			typ:    "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		})
	}
	defaultRels = append(defaultRels, defaultRel{target: "../theme/theme1.xml", typ: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"})

	return slideObjectRelationsToXml(&masterSlide.SlideBaseProps, defaultRels)
}

// ---------------------------------------------------------------------------
// makeXmlNotesMasterRel — ppt/notesMasters/_rels/notesMaster1.xml.rels
// ---------------------------------------------------------------------------

func makeXmlNotesMasterRel() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
		<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
		</Relationships>`
}

// ---------------------------------------------------------------------------
// getLayoutIdxForSlide
// ---------------------------------------------------------------------------

func getLayoutIdxForSlide(slides []PresSlide, slideLayouts []SlideLayout, slideNumber int) int {
	for i := range slideLayouts {
		if slideLayouts[i].Name == slides[slideNumber-1].SlideLayout.Name {
			return i + 1
		}
	}
	return 1
}

// ---------------------------------------------------------------------------
// makeXmlTheme — ppt/theme/theme1.xml
// ---------------------------------------------------------------------------

func makeXmlTheme(pres *IPresentationProps) string {
	majorFont := `<a:latin typeface="Calibri Light" panose="020F0302020204030204"/>`
	if pres.Theme.HeadFontFace != "" {
		majorFont = `<a:latin typeface="` + pres.Theme.HeadFontFace + `"/>`
	}
	minorFont := `<a:latin typeface="Calibri" panose="020F0502020204030204"/>`
	if pres.Theme.BodyFontFace != "" {
		minorFont = `<a:latin typeface="` + pres.Theme.BodyFontFace + `"/>`
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme"><a:themeElements><a:clrScheme name="Office"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="44546A"/></a:dk2><a:lt2><a:srgbClr val="E7E6E6"/></a:lt2><a:accent1><a:srgbClr val="4472C4"/></a:accent1><a:accent2><a:srgbClr val="ED7D31"/></a:accent2><a:accent3><a:srgbClr val="A5A5A5"/></a:accent3><a:accent4><a:srgbClr val="FFC000"/></a:accent4><a:accent5><a:srgbClr val="5B9BD5"/></a:accent5><a:accent6><a:srgbClr val="70AD47"/></a:accent6><a:hlink><a:srgbClr val="0563C1"/></a:hlink><a:folHlink><a:srgbClr val="954F72"/></a:folHlink></a:clrScheme><a:fontScheme name="Office"><a:majorFont>` + majorFont + `<a:ea typeface=""/><a:cs typeface=""/><a:font script="Jpan" typeface="游ゴシック Light"/><a:font script="Hang" typeface="맑은 고딕"/><a:font script="Hans" typeface="等线 Light"/><a:font script="Hant" typeface="新細明體"/><a:font script="Arab" typeface="Times New Roman"/><a:font script="Hebr" typeface="Times New Roman"/><a:font script="Thai" typeface="Angsana New"/><a:font script="Ethi" typeface="Nyala"/><a:font script="Beng" typeface="Vrinda"/><a:font script="Gujr" typeface="Shruti"/><a:font script="Khmr" typeface="MoolBoran"/><a:font script="Knda" typeface="Tunga"/><a:font script="Guru" typeface="Raavi"/><a:font script="Cans" typeface="Euphemia"/><a:font script="Cher" typeface="Plantagenet Cherokee"/><a:font script="Yiii" typeface="Microsoft Yi Baiti"/><a:font script="Tibt" typeface="Microsoft Himalaya"/><a:font script="Thaa" typeface="MV Boli"/><a:font script="Deva" typeface="Mangal"/><a:font script="Telu" typeface="Gautami"/><a:font script="Taml" typeface="Latha"/><a:font script="Syrc" typeface="Estrangelo Edessa"/><a:font script="Orya" typeface="Kalinga"/><a:font script="Mlym" typeface="Kartika"/><a:font script="Laoo" typeface="DokChampa"/><a:font script="Sinh" typeface="Iskoola Pota"/><a:font script="Mong" typeface="Mongolian Baiti"/><a:font script="Viet" typeface="Times New Roman"/><a:font script="Uigh" typeface="Microsoft Uighur"/><a:font script="Geor" typeface="Sylfaen"/><a:font script="Armn" typeface="Arial"/><a:font script="Bugi" typeface="Leelawadee UI"/><a:font script="Bopo" typeface="Microsoft JhengHei"/><a:font script="Java" typeface="Javanese Text"/><a:font script="Lisu" typeface="Segoe UI"/><a:font script="Mymr" typeface="Myanmar Text"/><a:font script="Nkoo" typeface="Ebrima"/><a:font script="Olck" typeface="Nirmala UI"/><a:font script="Osma" typeface="Ebrima"/><a:font script="Phag" typeface="Phagspa"/><a:font script="Syrn" typeface="Estrangelo Edessa"/><a:font script="Syrj" typeface="Estrangelo Edessa"/><a:font script="Syre" typeface="Estrangelo Edessa"/><a:font script="Sora" typeface="Nirmala UI"/><a:font script="Tale" typeface="Microsoft Tai Le"/><a:font script="Talu" typeface="Microsoft New Tai Lue"/><a:font script="Tfng" typeface="Ebrima"/></a:majorFont><a:minorFont>` + minorFont + `<a:ea typeface=""/><a:cs typeface=""/><a:font script="Jpan" typeface="游ゴシック"/><a:font script="Hang" typeface="맑은 고딕"/><a:font script="Hans" typeface="等线"/><a:font script="Hant" typeface="新細明體"/><a:font script="Arab" typeface="Arial"/><a:font script="Hebr" typeface="Arial"/><a:font script="Thai" typeface="Cordia New"/><a:font script="Ethi" typeface="Nyala"/><a:font script="Beng" typeface="Vrinda"/><a:font script="Gujr" typeface="Shruti"/><a:font script="Khmr" typeface="DaunPenh"/><a:font script="Knda" typeface="Tunga"/><a:font script="Guru" typeface="Raavi"/><a:font script="Cans" typeface="Euphemia"/><a:font script="Cher" typeface="Plantagenet Cherokee"/><a:font script="Yiii" typeface="Microsoft Yi Baiti"/><a:font script="Tibt" typeface="Microsoft Himalaya"/><a:font script="Thaa" typeface="MV Boli"/><a:font script="Deva" typeface="Mangal"/><a:font script="Telu" typeface="Gautami"/><a:font script="Taml" typeface="Latha"/><a:font script="Syrc" typeface="Estrangelo Edessa"/><a:font script="Orya" typeface="Kalinga"/><a:font script="Mlym" typeface="Kartika"/><a:font script="Laoo" typeface="DokChampa"/><a:font script="Sinh" typeface="Iskoola Pota"/><a:font script="Mong" typeface="Mongolian Baiti"/><a:font script="Viet" typeface="Arial"/><a:font script="Uigh" typeface="Microsoft Uighur"/><a:font script="Geor" typeface="Sylfaen"/><a:font script="Armn" typeface="Arial"/><a:font script="Bugi" typeface="Leelawadee UI"/><a:font script="Bopo" typeface="Microsoft JhengHei"/><a:font script="Java" typeface="Javanese Text"/><a:font script="Lisu" typeface="Segoe UI"/><a:font script="Mymr" typeface="Myanmar Text"/><a:font script="Nkoo" typeface="Ebrima"/><a:font script="Olck" typeface="Nirmala UI"/><a:font script="Osma" typeface="Ebrima"/><a:font script="Phag" typeface="Phagspa"/><a:font script="Syrn" typeface="Estrangelo Edessa"/><a:font script="Syrj" typeface="Estrangelo Edessa"/><a:font script="Syre" typeface="Estrangelo Edessa"/><a:font script="Sora" typeface="Nirmala UI"/><a:font script="Tale" typeface="Microsoft Tai Le"/><a:font script="Talu" typeface="Microsoft New Tai Lue"/><a:font script="Tfng" typeface="Ebrima"/></a:minorFont></a:fontScheme><a:fmtScheme name="Office"><a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"><a:lumMod val="110000"/><a:satMod val="105000"/><a:tint val="67000"/></a:schemeClr></a:gs><a:gs pos="50000"><a:schemeClr val="phClr"><a:lumMod val="105000"/><a:satMod val="103000"/><a:tint val="73000"/></a:schemeClr></a:gs><a:gs pos="100000"><a:schemeClr val="phClr"><a:lumMod val="105000"/><a:satMod val="109000"/><a:tint val="81000"/></a:schemeClr></a:gs></a:gsLst><a:lin ang="5400000" scaled="0"/></a:gradFill><a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"><a:satMod val="103000"/><a:lumMod val="102000"/><a:tint val="94000"/></a:schemeClr></a:gs><a:gs pos="50000"><a:schemeClr val="phClr"><a:satMod val="110000"/><a:lumMod val="100000"/><a:shade val="100000"/></a:schemeClr></a:gs><a:gs pos="100000"><a:schemeClr val="phClr"><a:lumMod val="99000"/><a:satMod val="120000"/><a:shade val="78000"/></a:schemeClr></a:gs></a:gsLst><a:lin ang="5400000" scaled="0"/></a:gradFill></a:fillStyleLst><a:lnStyleLst><a:ln w="6350" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln><a:ln w="12700" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln><a:ln w="19050" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/><a:miter lim="800000"/></a:ln></a:lnStyleLst><a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst><a:outerShdw blurRad="57150" dist="19050" dir="5400000" algn="ctr" rotWithShape="0"><a:srgbClr val="000000"><a:alpha val="63000"/></a:srgbClr></a:outerShdw></a:effectLst></a:effectStyle></a:effectStyleLst><a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"><a:tint val="95000"/><a:satMod val="170000"/></a:schemeClr></a:solidFill><a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"><a:tint val="93000"/><a:satMod val="150000"/><a:shade val="98000"/><a:lumMod val="102000"/></a:schemeClr></a:gs><a:gs pos="50000"><a:schemeClr val="phClr"><a:tint val="98000"/><a:satMod val="130000"/><a:shade val="90000"/><a:lumMod val="103000"/></a:schemeClr></a:gs><a:gs pos="100000"><a:schemeClr val="phClr"><a:shade val="63000"/><a:satMod val="120000"/></a:schemeClr></a:gs></a:gsLst><a:lin ang="5400000" scaled="0"/></a:gradFill></a:bgFillStyleLst></a:fmtScheme></a:themeElements><a:objectDefaults/><a:extraClrSchemeLst/><a:extLst><a:ext uri="{05A4C25C-085E-4340-85A3-A5531E510DB2}"><thm15:themeFamily xmlns:thm15="http://schemas.microsoft.com/office/thememl/2012/main" name="Office Theme" id="{62F939B6-93AF-4DB8-9C6B-D6C7DFDC589F}" vid="{4A3C46E8-61CC-4603-A589-7422A47A8E4A}"/></a:ext></a:extLst></a:theme>`
}

// ---------------------------------------------------------------------------
// makeXmlPresentation — ppt/presentation.xml
// ---------------------------------------------------------------------------

func makeXmlPresentation(pres *IPresentationProps) string {
	rtl := ""
	if pres.RtlMode {
		rtl = `rtl="1"`
	}
	embed := ""
	if len(pres.EmbeddedFonts) > 0 {
		embed = `embedTrueTypeFonts="1" `
	}
	strXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF +
		`<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" ` +
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" ` + rtl + ` ` + embed + `saveSubsetFonts="1" autoCompressPictures="0">`

	// STEP 1: slide master
	strXml += `<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>`

	// Font embedding (net-new): embeddedFontLst immediately after sldMasterIdLst,
	// before sldIdLst (per PORTING.md / ECMA-376 placement guidance).
	if len(pres.EmbeddedFonts) > 0 {
		strXml += makeEmbeddedFontLst(len(pres.Slides), pres.EmbeddedFonts)
	}

	// STEP 2: slides
	strXml += `<p:sldIdLst>`
	for i := range pres.Slides {
		strXml += `<p:sldId id="` + itoa(pres.Slides[i].SlideID) + `" r:id="rId` + itoa(pres.Slides[i].RID) + `"/>`
	}
	strXml += `</p:sldIdLst>`

	// STEP 3: notes master
	strXml += `<p:notesMasterIdLst><p:notesMasterId r:id="rId` + itoa(len(pres.Slides)+2) + `"/></p:notesMasterIdLst>`

	// STEP 4: sizes
	strXml += `<p:sldSz cx="` + itoa(pres.PresLayout.Width) + `" cy="` + itoa(pres.PresLayout.Height) + `"/>`
	strXml += `<p:notesSz cx="` + itoa(pres.PresLayout.Height) + `" cy="` + itoa(pres.PresLayout.Width) + `"/>`

	// STEP 5: text styles
	strXml += `<p:defaultTextStyle>`
	for idy := 1; idy < 10; idy++ {
		strXml += `<a:lvl` + itoa(idy) + `pPr marL="` + itoa((idy-1)*457200) + `" algn="l" defTabSz="914400" rtl="0" eaLnBrk="1" latinLnBrk="0" hangingPunct="1">` +
			`<a:defRPr sz="1800" kern="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/><a:ea typeface="+mn-ea"/><a:cs typeface="+mn-cs"/>` +
			`</a:defRPr></a:lvl` + itoa(idy) + `pPr>`
	}
	strXml += `</p:defaultTextStyle>`

	// STEP 6: sections
	if len(pres.Sections) > 0 {
		strXml += `<p:extLst><p:ext uri="{521415D9-36F7-43E2-AB2F-B90AF26B5E84}">`
		strXml += `<p14:sectionLst xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main">`
		for si := range pres.Sections {
			sect := &pres.Sections[si]
			strXml += `<p14:section name="` + encodeXmlEntities(sect.Title) + `" id="{` + getUuid("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx") + `}"><p14:sldIdLst>`
			for sli := range sect.Slides {
				strXml += `<p14:sldId id="` + itoa(sect.Slides[sli].SlideID) + `"/>`
			}
			strXml += `</p14:sldIdLst></p14:section>`
		}
		strXml += `</p14:sectionLst></p:ext>`
		strXml += `<p:ext uri="{EFAFB233-063F-42B5-8137-9DF3F51BA10A}"><p15:sldGuideLst xmlns:p15="http://schemas.microsoft.com/office/powerpoint/2012/main"/></p:ext>`
		strXml += `</p:extLst>`
	}

	strXml += `</p:presentation>`
	return strXml
}

// makeEmbeddedFontLst builds <p:embeddedFontLst> (one <p:embeddedFont> per
// typeface). r:ids match embeddedFontVariants (and makeXmlPresentationRels).
func makeEmbeddedFontLst(numSlides int, fonts []*EmbeddedFont) string {
	variants := embeddedFontVariants(numSlides, fonts)
	// index variants by typeface preserving order
	var b strings.Builder
	b.WriteString(`<p:embeddedFontLst>`)
	vi := 0
	for _, f := range fonts {
		if f == nil {
			continue
		}
		b.WriteString(`<p:embeddedFont><p:font typeface="` + f.Typeface + `"/>`)
		for _, st := range fontStyleOrder {
			if _, ok := f.Variants[st]; ok {
				b.WriteString(`<` + embeddedFontXMLTags[st] + ` r:id="rId` + itoa(variants[vi].rID) + `"/>`)
				vi++
			}
		}
		b.WriteString(`</p:embeddedFont>`)
	}
	b.WriteString(`</p:embeddedFontLst>`)
	return b.String()
}

// ---------------------------------------------------------------------------
// makeXmlPresProps / makeXmlTableStyles / makeXmlViewProps
// ---------------------------------------------------------------------------

func makeXmlPresProps() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<p:presentationPr xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`
}

func makeXmlTableStyles() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<a:tblStyleLst xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" def="{5C22544A-7EE6-4342-B048-85BDC9FD1C3A}"/>`
}

func makeXmlViewProps() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + CRLF + `<p:viewPr xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:normalViewPr horzBarState="maximized"><p:restoredLeft sz="15611"/><p:restoredTop sz="94610"/></p:normalViewPr><p:slideViewPr><p:cSldViewPr snapToGrid="0" snapToObjects="1"><p:cViewPr varScale="1"><p:scale><a:sx n="136" d="100"/><a:sy n="136" d="100"/></p:scale><p:origin x="216" y="312"/></p:cViewPr><p:guideLst/></p:cSldViewPr></p:slideViewPr><p:notesTextViewPr><p:cViewPr><p:scale><a:sx n="1" d="1"/><a:sy n="1" d="1"/></p:scale><p:origin x="0" y="0"/></p:cViewPr></p:notesTextViewPr><p:gridSpacing cx="76200" cy="76200"/></p:viewPr>`
}
