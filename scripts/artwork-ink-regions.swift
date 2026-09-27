#!/usr/bin/env swift
// Conservative occupied tiles for a pinned PNG raster of original SVG artwork.
// Read-only pixel analysis: no modification to the source image or its bytes.
import AppKit
import Foundation

guard CommandLine.arguments.count == 2,
      let source = try? Data(contentsOf: URL(fileURLWithPath: CommandLine.arguments[1])),
      let bitmap = NSBitmapImageRep(data: source) else {
    fputs("Usage: artwork-ink-regions.swift source-raster.png\n", stderr)
    exit(2)
}
let grid = 32
let w = bitmap.pixelsWide, h = bitmap.pixelsHigh
var occupied = Set<Int>()
for y in 0..<h {
    for x in 0..<w {
        if let c = bitmap.colorAt(x: x, y: y), c.alphaComponent > 0 {
            // Dilate by one raster pixel to include antialiasing uncertainty.
            for xx in max(0, x-1)...min(w-1, x+1) {
                for yy in max(0, y-1)...min(h-1, y+1) {
                    occupied.insert((yy*grid/h)*grid + xx*grid/w)
                }
            }
        }
    }
}
let regions = occupied.sorted().map { n -> [String: Double] in
    ["x": Double(n % grid)/Double(grid), "y": Double(n/grid)/Double(grid),
     "width": 1/Double(grid), "height": 1/Double(grid)]
}
let result: [String: Any] = ["regions": regions, "grid": grid,
    "raster_width": w, "raster_height": h, "dilation_pixels": 1,
    "method": "NSBitmapImageRep alpha > 0; conservative 32x32 occupied tiles with one raster pixel dilation"]
let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
FileHandle.standardOutput.write(data)
