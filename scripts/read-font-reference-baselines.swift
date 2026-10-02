#!/usr/bin/env swift
// Read text-show origins from simple horizontal PowerPoint probe PDFs.
// These are PDF baselines, never PowerPoint TextRange or selection bounds.
import CoreGraphics
import CryptoKit
import Foundation

struct Origin: Codable {
    let baseline_pt, font_size_pt: Double
    let font_resource: String
}
struct PageOrigins: Codable { let page_number: Int; let origins: [Origin] }
struct Report: Encodable {
    let schema = "pptxgengo.powerpoint-pdf-baselines.v1"
    let contract = "powerpoint-pdf-text-show-origins.v1"
    let pdf_sha256: String
    let pages: [PageOrigins]
}
final class State {
    var ctm = CGAffineTransform.identity
    var stack = [(CGAffineTransform, CGFloat, CGFloat, CGFloat, String)]()
    var matrix = CGAffineTransform.identity
    var lineMatrix = CGAffineTransform.identity
    var leading: CGFloat = 0
    var rise: CGFloat = 0
    var fontSize: CGFloat = 0
    var fontName = ""
    var origins = [Origin]()
    var errors = [String]()
    let pageHeight: CGFloat
    init(_ height: CGFloat) { pageHeight = height }
}
func state(_ info: UnsafeMutableRawPointer?) -> State {
    Unmanaged<State>.fromOpaque(info!).takeUnretainedValue()
}
func numbers(_ scanner: CGPDFScannerRef, _ count: Int, _ s: State) -> [CGFloat] {
    var values = [CGFloat]()
    for _ in 0..<count {
        var value: CGPDFReal = 0
        if !CGPDFScannerPopNumber(scanner, &value) { s.errors.append("Missing numeric operand") }
        values.insert(CGFloat(value), at: 0)
    }
    return values
}
func translate(_ s: State, _ x: CGFloat, _ y: CGFloat) {
    s.lineMatrix = CGAffineTransform(translationX: x, y: y).concatenating(s.lineMatrix)
    s.matrix = s.lineMatrix
}
func show(_ s: State) {
    guard abs(s.ctm.b) < 0.000001, abs(s.ctm.c) < 0.000001,
          abs(s.matrix.b) < 0.000001, abs(s.matrix.c) < 0.000001 else {
        s.errors.append("Rotated or skewed text is outside this probe reader's scope"); return
    }
    let point = CGPoint(x: 0, y: s.rise).applying(s.matrix).applying(s.ctm)
    s.origins.append(Origin(baseline_pt: Double(s.pageHeight-point.y),
        font_size_pt: Double(abs(s.fontSize*s.matrix.d*s.ctm.d)), font_resource: s.fontName))
}
do {
    guard CommandLine.arguments.count == 3 else { throw NSError(domain: "baselines", code: 64, userInfo: [NSLocalizedDescriptionKey: "Usage: input.pdf new-output.json"]) }
    let input = URL(fileURLWithPath: CommandLine.arguments[1])
    let output = URL(fileURLWithPath: CommandLine.arguments[2])
    guard !FileManager.default.fileExists(atPath: output.path) else { throw NSError(domain: "baselines", code: 65, userInfo: [NSLocalizedDescriptionKey: "Output already exists"]) }
    let data = try Data(contentsOf: input)
    guard let document = CGPDFDocument(input as CFURL), let table = CGPDFOperatorTableCreate() else { throw NSError(domain: "baselines", code: 66) }
    CGPDFOperatorTableSetCallback(table, "q") { _, info in let s=state(info); s.stack.append((s.ctm,s.leading,s.rise,s.fontSize,s.fontName)) }
    CGPDFOperatorTableSetCallback(table, "Q") { _, info in let s=state(info); if let value=s.stack.popLast() { (s.ctm,s.leading,s.rise,s.fontSize,s.fontName)=value } else { s.errors.append("Unbalanced graphics state") } }
    CGPDFOperatorTableSetCallback(table, "cm") { scanner, info in let s=state(info); let n=numbers(scanner,6,s); s.ctm=CGAffineTransform(a:n[0],b:n[1],c:n[2],d:n[3],tx:n[4],ty:n[5]).concatenating(s.ctm) }
    CGPDFOperatorTableSetCallback(table, "BT") { _, info in let s=state(info); s.matrix = .identity; s.lineMatrix = .identity }
    CGPDFOperatorTableSetCallback(table, "Tm") { scanner, info in let s=state(info); let n=numbers(scanner,6,s); s.matrix=CGAffineTransform(a:n[0],b:n[1],c:n[2],d:n[3],tx:n[4],ty:n[5]); s.lineMatrix=s.matrix }
    CGPDFOperatorTableSetCallback(table, "Td") { scanner, info in let s=state(info); let n=numbers(scanner,2,s); translate(s,n[0],n[1]) }
    CGPDFOperatorTableSetCallback(table, "TD") { scanner, info in let s=state(info); let n=numbers(scanner,2,s); s.leading = -n[1]; translate(s,n[0],n[1]) }
    CGPDFOperatorTableSetCallback(table, "TL") { scanner, info in let s=state(info); s.leading=numbers(scanner,1,s)[0] }
    CGPDFOperatorTableSetCallback(table, "Ts") { scanner, info in let s=state(info); s.rise=numbers(scanner,1,s)[0] }
    CGPDFOperatorTableSetCallback(table, "T*") { _, info in let s=state(info); translate(s,0,-s.leading) }
    CGPDFOperatorTableSetCallback(table, "Tf") { scanner, info in
        let s=state(info); s.fontSize=numbers(scanner,1,s)[0]
        var name: UnsafePointer<CChar>?
        if CGPDFScannerPopName(scanner,&name), let name=name { s.fontName=String(cString:name) } else { s.errors.append("Missing font name") }
    }
    CGPDFOperatorTableSetCallback(table, "Tj") { scanner, info in
        let s=state(info); var value: CGPDFStringRef?
        if CGPDFScannerPopString(scanner,&value), let value=value, CGPDFStringGetLength(value)>0 { show(s) }
    }
    CGPDFOperatorTableSetCallback(table, "TJ") { scanner, info in
        let s=state(info); var value: CGPDFArrayRef?
        if CGPDFScannerPopArray(scanner,&value), let value=value, CGPDFArrayGetCount(value)>0 { show(s) }
    }
    // Refuse uncommon forms rather than silently missing text or coordinates.
    for op in ["Do", "'", "\""] {
        CGPDFOperatorTableSetCallback(table, op) { _, info in state(info).errors.append("XObjects or shorthand text operators are outside this probe reader's scope") }
    }
    var pages = [PageOrigins]()
    for index in 1...document.numberOfPages {
        guard let page=document.page(at:index) else { throw NSError(domain:"baselines",code:67) }
        let s=State(page.getBoxRect(.mediaBox).maxY)
        let live=CGPDFScannerCreate(CGPDFContentStreamCreateWithPage(page),table,Unmanaged.passUnretained(s).toOpaque())
        guard CGPDFScannerScan(live), s.errors.isEmpty, s.stack.isEmpty, !s.origins.isEmpty else {
            throw NSError(domain:"baselines",code:68,userInfo:[NSLocalizedDescriptionKey:"Page \(index): \(s.errors)"])
        }
        pages.append(PageOrigins(page_number:index,origins:s.origins))
    }
    let report=Report(pdf_sha256:SHA256.hash(data:data).map{String(format:"%02x",$0)}.joined(),pages:pages)
    let encoder=JSONEncoder(); encoder.outputFormatting=[.prettyPrinted,.sortedKeys]
    var result=try encoder.encode(report); result.append(10)
    try result.write(to:output,options:.withoutOverwriting)
    print("\(output.path): \(pages.count) pages")
} catch { fputs("read-font-reference-baselines: \(error)\n",stderr); exit(1) }
