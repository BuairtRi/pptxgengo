#!/usr/bin/env swift
// Derived raster preview of pinned SVG artwork; retain the original SVG separately.
import AppKit
import Foundation

guard CommandLine.arguments.count == 3 else {
    fputs("Usage: render-svg-preview.swift source.svg new-preview.png\n", stderr)
    exit(2)
}
let source = URL(fileURLWithPath: CommandLine.arguments[1])
let output = URL(fileURLWithPath: CommandLine.arguments[2])
guard !FileManager.default.fileExists(atPath: output.path) else {
    fputs("Output must be a new file\n", stderr)
    exit(2)
}
guard let image = NSImage(contentsOf: source), image.size.width > 0, image.size.height > 0 else {
    fputs("AppKit cannot decode the source SVG\n", stderr)
    exit(1)
}
let scale = 8.0
let width = Int(ceil(image.size.width * scale))
let height = Int(ceil(image.size.height * scale))
guard width <= 8192, height <= 8192,
      let bitmap = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: width, pixelsHigh: height,
                                   bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true,
                                   isPlanar: false, colorSpaceName: .deviceRGB,
                                   bytesPerRow: 0, bitsPerPixel: 0),
      let context = NSGraphicsContext(bitmapImageRep: bitmap) else {
    fputs("Unsupported or oversized SVG preview dimensions\n", stderr)
    exit(1)
}
NSGraphicsContext.saveGraphicsState()
NSGraphicsContext.current = context
context.imageInterpolation = .high
image.draw(in: NSRect(x: 0, y: 0, width: width, height: height),
           from: .zero, operation: .copy, fraction: 1)
context.flushGraphics()
NSGraphicsContext.restoreGraphicsState()
guard let png = bitmap.representation(using: .png, properties: [:]) else {
    fputs("Cannot encode PNG preview\n", stderr)
    exit(1)
}
do {
    try png.write(to: output, options: .withoutOverwriting)
    print("\(output.path) \(width)x\(height); AppKit NSImage SVG preview at 8x")
} catch {
    fputs("\(error)\n", stderr)
    exit(1)
}
