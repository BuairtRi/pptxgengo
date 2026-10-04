import AppKit
import Foundation
import ImageIO
import PDFKit

func run() throws {
 let args = Array(CommandLine.arguments.dropFirst())
 guard args.count == 3, let pdf = PDFDocument(url: URL(fileURLWithPath: args[0])), pdf.pageCount > 0 else {
  throw NSError(domain: "nativeexport", code: 1, userInfo: [NSLocalizedDescriptionKey: "Cannot read the exported PDF"])
 }
 if args[2] == "png" {
  let output = URL(fileURLWithPath: args[1], isDirectory: true)
  try FileManager.default.createDirectory(at: output, withIntermediateDirectories: true)
  for index in 0..<pdf.pageCount {
   guard let page = pdf.page(at: index) else { throw NSError(domain: "nativeexport", code: 2) }
   let bounds = page.bounds(for: .cropBox).standardized
   guard bounds.width > 0, bounds.height > 0 else { throw NSError(domain: "nativeexport", code: 3) }
   let scale = min(1920.0 / bounds.width, 1080.0 / bounds.height)
   let size = NSSize(width: max(1, (bounds.width * scale).rounded()), height: max(1, (bounds.height * scale).rounded()))
   let url = output.appendingPathComponent(String(format: "slide-%03d.png", index + 1))
   guard !FileManager.default.fileExists(atPath: url.path),
    let image = page.thumbnail(of: size, for: .cropBox).cgImage(forProposedRect: nil, context: nil, hints: nil),
    let destination = CGImageDestinationCreateWithURL(url as CFURL, "public.png" as CFString, 1, nil) else {
    throw NSError(domain: "nativeexport", code: 4, userInfo: [NSLocalizedDescriptionKey: "Cannot create PNG page \(index + 1)"])
   }
   CGImageDestinationAddImage(destination, image, nil)
   guard CGImageDestinationFinalize(destination) else { throw NSError(domain: "nativeexport", code: 5) }
  }
 }
 let json = try JSONSerialization.data(withJSONObject: ["pages": pdf.pageCount])
 print(String(decoding: json, as: UTF8.self))
}
do { try run() } catch { fputs("PDFKit: \(error)\n", stderr); exit(1) }
