package main

import (
	"encoding/base64"
	"github.com/buairtri/pptxgengo/internal/compose"
)

func canvasElements(items []compose.PlannedCanvas) []element {
	var out []element
	for _, c := range items {
		out = append(out, element{Name: "canvas:" + base64.RawURLEncoding.EncodeToString([]byte(c.ID)), Kind: c.Kind, Frame: rect(c.Bounds), Text: c.Text, Paragraphs: c.Paragraphs, FontFace: c.FontFace, FontSize: c.FontSizePt, Bold: c.Bold, Foreground: color(c.Foreground), Background: color(c.Background), InsetX: c.InsetX, InsetY: c.InsetY, Align: c.Align, Valign: c.Valign, MeasurementID: c.MeasurementID, LineWidth: c.LineWidthPt, AssetPath: c.AssetPath, AssetSHA256: c.AssetSHA256, AltText: c.AltText, ImageFit: c.ImageFit, ImageCrop: c.ImageCrop, FocalX: c.FocalX, FocalY: c.FocalY, Preset: c.Preset, Adjustments: c.Adjustments, Pattern: c.Pattern})
	}
	return out
}
func cardElements(cards []compose.PlannedCard) []element {
	var out []element
	for _, c := range cards {
		name := "card:" + base64.RawURLEncoding.EncodeToString([]byte(c.ID))
		out = append(out, element{Name: name + "-surface", Kind: "surface", Frame: rect(c.Bounds), Background: color(c.Surface)})
		if c.Kind == "numbered" {
			out = append(out, element{Name: name + "-accent", Kind: "surface", Frame: rect(c.AccentBounds), Background: color(c.Accent)})
		}
		for _, b := range c.Blocks {
			out = append(out, element{Name: name + "-" + b.Role, Kind: "text", Frame: rect(b.Bounds), Text: b.Text, FontFace: b.FontFace, FontSize: b.FontSizePt, Bold: b.Bold, Foreground: color(b.Foreground), InsetX: b.HorizontalInsetPt, InsetY: b.VerticalInsetPt, Align: b.Align, MeasurementID: b.MeasurementID})
		}
	}
	return out
}

func accentElements(items []compose.PlannedAccent) []element {
	var out []element
	for _, a := range items {
		out = append(out, element{Name: "accent:" + base64.RawURLEncoding.EncodeToString([]byte(a.ID)), Kind: "image", Frame: rect(a.Bounds), AssetPath: a.AssetPath, AssetSHA256: a.AssetSHA256, AltText: a.Mode + " for " + a.Target, AssetMode: "stretch"})
	}
	return out
}
