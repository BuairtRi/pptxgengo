#!/usr/bin/env swift
import AppKit
import Foundation

struct BoundsReport: Codable {
    let file: String
    let viewBox: [Double]
    let rasterResolution: Int
    let rasterDimensions: [Int]
    let alphaCriterion: String
    let visibleBounds: Rect
    let paddingFractions: Padding
    let transformedBoundsEMU: Rect?
}

struct Rect: Codable { let x: Double; let y: Double; let width: Double; let height: Double }
struct Padding: Codable { let left: Double; let top: Double; let right: Double; let bottom: Double }

final class RootViewBoxParser: NSObject, XMLParserDelegate {
    var viewBox: [Double]?
    func parser(_ parser: XMLParser, didStartElement elementName: String, namespaceURI: String?, qualifiedName qName: String?, attributes attributeDict: [String : String] = [:]) {
        guard viewBox == nil, elementName == "svg", let raw = attributeDict["viewBox"] else { return }
        let values = raw.split(whereSeparator: { $0 == " " || $0 == "," || $0 == "\t" || $0 == "\n" }).compactMap { Double($0) }
        if values.count == 4, values[2] > 0, values[3] > 0 { viewBox = values }
    }
}

func fail(_ message: String) -> Never {
    fputs("svg-visible-bounds: \(message)\n", stderr)
    exit(2)
}

var resolution = 4096
var pngPath: String?
var inputPath: String?
var placementEMU: [Double]?
var rotationDegrees = 0.0
var index = 1
while index < CommandLine.arguments.count {
    let arg = CommandLine.arguments[index]
    if arg == "--resolution" {
        index += 1
        guard index < CommandLine.arguments.count, let n = Int(CommandLine.arguments[index]), (256...16384).contains(n) else { fail("--resolution must be an integer from 256 through 16384") }
        resolution = n
    } else if arg == "--png" {
        index += 1
        guard index < CommandLine.arguments.count else { fail("--png requires an output path") }
        pngPath = CommandLine.arguments[index]
    } else if arg == "--placement-emu" {
        index += 1
        guard index < CommandLine.arguments.count else { fail("--placement-emu requires x,y,width,height") }
        let numbers = CommandLine.arguments[index].split(separator: ",").compactMap { Double($0) }
        guard numbers.count == 4, numbers[2] > 0, numbers[3] > 0 else { fail("--placement-emu requires x,y,width,height with positive width and height") }
        placementEMU = numbers
    } else if arg == "--rotation-degrees" {
        index += 1
        guard index < CommandLine.arguments.count, let degrees = Double(CommandLine.arguments[index]) else { fail("--rotation-degrees requires a numeric angle") }
        rotationDegrees = degrees
    } else if arg.hasPrefix("-") {
        fail("unknown option \(arg); usage: swift scripts/svg-visible-bounds.swift [--resolution N] [--placement-emu x,y,width,height] [--rotation-degrees degrees] [--png output.png] image.svg")
    } else if inputPath == nil {
        inputPath = arg
    } else {
        fail("provide exactly one SVG input; usage: swift scripts/svg-visible-bounds.swift [--resolution N] [--placement-emu x,y,width,height] [--rotation-degrees degrees] [--png output.png] image.svg")
    }
    index += 1
}
guard let inputPath else { fail("missing SVG input; usage: swift scripts/svg-visible-bounds.swift [--resolution N] [--placement-emu x,y,width,height] [--rotation-degrees degrees] [--png output.png] image.svg") }

let inputURL = URL(fileURLWithPath: inputPath)
guard let xml = try? Data(contentsOf: inputURL) else { fail("cannot read \(inputPath)") }
let viewBoxParser = RootViewBoxParser()
let isSVG = inputURL.pathExtension.lowercased() == "svg"
let vb: [Double]
let pngSourceBitmap: NSBitmapImageRep?
if isSVG {
    let xmlParser = XMLParser(data: xml)
    xmlParser.delegate = viewBoxParser
    guard xmlParser.parse(), let parsed = viewBoxParser.viewBox else { fail("could not parse a positive four-number SVG viewBox from \(inputPath)") }
    vb = parsed
    pngSourceBitmap = nil
} else {
    guard inputURL.pathExtension.lowercased() == "png",
          let sourceBitmap = NSBitmapImageRep(data: xml),
          sourceBitmap.pixelsWide > 0, sourceBitmap.pixelsHigh > 0 else { fail("input must be a decodable SVG or PNG: \(inputPath)") }
    vb = [0, 0, Double(sourceBitmap.pixelsWide), Double(sourceBitmap.pixelsHigh)]
    pngSourceBitmap = sourceBitmap
}
guard let svgImage = NSImage(contentsOf: inputURL) else { fail("AppKit NSImage could not decode \(inputPath)") }

let scanWidth = resolution
let scanHeight = resolution
let bitmap = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: resolution, pixelsHigh: resolution,
                              bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
                              colorSpaceName: .deviceRGB, bytesPerRow: resolution * 4, bitsPerPixel: 32)!
bitmap.size = NSSize(width: vb[2], height: vb[3])
NSGraphicsContext.saveGraphicsState()
guard let graphics = NSGraphicsContext(bitmapImageRep: bitmap) else { fail("could not create bitmap drawing context") }
NSGraphicsContext.current = graphics
graphics.imageInterpolation = .high
svgImage.draw(in: NSRect(x: 0, y: 0, width: vb[2], height: vb[3]),
              from: NSRect(origin: .zero, size: svgImage.size), operation: .copy, fraction: 1)
graphics.flushGraphics()
NSGraphicsContext.restoreGraphicsState()

guard let pixels = bitmap.bitmapData else { fail("could not read rendered SVG pixels") }
var minX = scanWidth, minY = scanHeight, maxX = -1, maxY = -1
var transformedMinX = Double.infinity, transformedMinY = Double.infinity
var transformedMaxX = -Double.infinity, transformedMaxY = -Double.infinity
let angle = rotationDegrees * Double.pi / 180
let cosine = cos(angle), sine = sin(angle)
let sx = vb[2] / Double(scanWidth), sy = vb[3] / Double(scanHeight)
func includeTransformedCorner(pixelX: Int, pixelY: Int) {
    guard let place = placementEMU else { return }
    let svgX = vb[0] + Double(pixelX) * sx, svgY = vb[1] + Double(pixelY) * sy
    let dx = (svgX - (vb[0] + vb[2] / 2)) / vb[2] * place[2]
    let dy = (svgY - (vb[1] + vb[3] / 2)) / vb[3] * place[3]
    // OOXML positive rotation is clockwise; slide coordinates grow downwards.
    let tx = place[0] + place[2] / 2 + dx * cosine - dy * sine
    let ty = place[1] + place[3] / 2 + dx * sine + dy * cosine
    transformedMinX = min(transformedMinX, tx); transformedMaxX = max(transformedMaxX, tx)
    transformedMinY = min(transformedMinY, ty); transformedMaxY = max(transformedMaxY, ty)
}
for y in 0..<resolution {
    for x in 0..<resolution {
        // NSBitmapImageRep's default device RGB layout stores alpha as the fourth sample.
        if pixels[y * bitmap.bytesPerRow + x * 4 + 3] != 0 {
            minX = min(minX, x); minY = min(minY, y)
            maxX = max(maxX, x); maxY = max(maxY, y)
            includeTransformedCorner(pixelX: x, pixelY: y)
            includeTransformedCorner(pixelX: x + 1, pixelY: y)
            includeTransformedCorner(pixelX: x, pixelY: y + 1)
            includeTransformedCorner(pixelX: x + 1, pixelY: y + 1)
        }
    }
}
guard maxX >= minX, maxY >= minY else { fail("rendered SVG has no nontransparent pixels") }

// Pixel-edge bounds include antialiased edge pixels, so this is a conservative
// visible bound whose resolution error is at most one raster pixel per edge.
let visible = Rect(x: vb[0] + Double(minX) * sx, y: vb[1] + Double(minY) * sy,
                   width: Double(maxX - minX + 1) * sx, height: Double(maxY - minY + 1) * sy)
let pad = Padding(left: (visible.x - vb[0]) / vb[2],
                  top: (visible.y - vb[1]) / vb[3],
                  right: (vb[0] + vb[2] - (visible.x + visible.width)) / vb[2],
                  bottom: (vb[1] + vb[3] - (visible.y + visible.height)) / vb[3])
let transformed: Rect? = placementEMU == nil ? nil : Rect(x: transformedMinX, y: transformedMinY,
                                                           width: transformedMaxX - transformedMinX,
                                                           height: transformedMaxY - transformedMinY)
let report = BoundsReport(file: inputURL.lastPathComponent, viewBox: vb, rasterResolution: max(scanWidth, scanHeight), rasterDimensions: [scanWidth, scanHeight],
                          alphaCriterion: "alpha > 0; includes antialiased edge coverage for SVG and nontransparent source pixels for PNG",
                          visibleBounds: visible, paddingFractions: pad, transformedBoundsEMU: transformed)
let encoder = JSONEncoder()
encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
guard let json = try? encoder.encode(report) else { fail("could not encode JSON report") }
FileHandle.standardOutput.write(json)
FileHandle.standardOutput.write(Data([0x0A]))
if let pngPath {
    guard let png = bitmap.representation(using: .png, properties: [:]) else { fail("could not encode PNG") }
    do { try png.write(to: URL(fileURLWithPath: pngPath), options: .atomic) }
    catch { fail("could not write PNG \(pngPath): \(error.localizedDescription)") }
}
