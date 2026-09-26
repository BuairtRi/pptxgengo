#!/usr/bin/env swift

import AppKit
import Foundation
import ImageIO
import PDFKit

struct UsageError: Error, CustomStringConvertible {
    let description: String
}

func usage() -> Never {
    fputs("Usage: render-pdf.swift [--overwrite] input.pdf output-dir page-number [page-number ...]\n", stderr)
    exit(2)
}

func run() throws {
    var args = Array(CommandLine.arguments.dropFirst())
    let overwrite: Bool
    if args.first == "--overwrite" {
        overwrite = true
        args.removeFirst()
    } else {
        overwrite = false
    }
    guard args.count >= 3 else { usage() }

    let inputURL = URL(fileURLWithPath: args[0]).standardizedFileURL
    let outputDir = URL(fileURLWithPath: args[1], isDirectory: true).standardizedFileURL
    let pageNumbers = try args.dropFirst(2).map { value -> Int in
        guard let number = Int(value), number > 0 else {
            throw UsageError(description: "Page numbers must be positive integers: \(value)")
        }
        return number
    }
    guard let document = PDFDocument(url: inputURL) else {
        throw UsageError(description: "Could not open PDF: \(inputURL.path)")
    }
    try FileManager.default.createDirectory(at: outputDir, withIntermediateDirectories: true)

    for pageNumber in pageNumbers {
        guard let page = document.page(at: pageNumber - 1) else {
            throw UsageError(description: "Page \(pageNumber) is outside the PDF (1–\(document.pageCount))")
        }
        let outputURL = outputDir.appendingPathComponent(String(format: "slide-%03d.png", pageNumber))
        if FileManager.default.fileExists(atPath: outputURL.path), !overwrite {
            throw UsageError(description: "Output already exists (pass --overwrite to replace it): \(outputURL.path)")
        }

        // Fit the page into a 1920×1080 pixel box while preserving the PDF's
        // crop-box aspect ratio. A 16:9 slide therefore renders at 1920×1080;
        // other page ratios use the largest proportional dimensions within it.
        let pageBounds = page.bounds(for: .cropBox).standardized
        guard pageBounds.width > 0, pageBounds.height > 0 else {
            throw UsageError(description: "Page \(pageNumber) has invalid crop-box dimensions")
        }
        let scale = min(1920.0 / pageBounds.width, 1080.0 / pageBounds.height)
        let pixelWidth = max(1, Int((pageBounds.width * scale).rounded()))
        let pixelHeight = max(1, Int((pageBounds.height * scale).rounded()))
        guard let image = page.thumbnail(of: NSSize(width: pixelWidth, height: pixelHeight), for: .cropBox).cgImage(forProposedRect: nil, context: nil, hints: nil) else {
            throw UsageError(description: "Could not rasterize PDF page \(pageNumber)")
        }

        guard let destination = CGImageDestinationCreateWithURL(outputURL as CFURL, "public.png" as CFString, 1, nil) else {
            throw UsageError(description: "Could not create PNG output: \(outputURL.path)")
        }
        CGImageDestinationAddImage(destination, image, nil)
        guard CGImageDestinationFinalize(destination) else {
            throw UsageError(description: "Could not write PNG output: \(outputURL.path)")
        }
        print("\(outputURL.path) \(image.width)x\(image.height)")
    }
}

do {
    try run()
} catch {
    fputs("render-pdf.swift: \(error)\n", stderr)
    exit(1)
}
