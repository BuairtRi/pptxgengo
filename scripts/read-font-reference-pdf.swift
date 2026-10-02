#!/usr/bin/env swift
// Read PowerPoint-exported PDF selections. This never reports TextRange bounds.
import AppKit
import CryptoKit
import CoreGraphics
import Foundation
import PDFKit

struct Rect: Codable { let x, y, width, height: Double }
struct Line: Codable { let text: String; let bounds: Rect; let font_names: [String]; let font_sizes_pt: [Double] }
struct Page: Encodable { let page_number: Int; let width_pt, height_pt: Double; let lines: [Line]; let bounds: Rect; let font_names: [String]; let selection_font_names: [String] }
struct Report: Encodable {
    let schema = "pptxgengo.powerpoint-pdf-reference.v1"
    let contract = "powerpoint-pdf-selection-bounds.v1"
    let font_metadata_source = "PDF page font resource BaseFont names; selection attributes may substitute fonts in the extractor process"
    let pdf_file, pdf_sha256: String
    let page_count: Int
    let pages: [Page]
}
func resourceFonts(_ page: PDFPage) -> [String] {
    guard let dictionary=page.pageRef?.dictionary else { return [] }
    var resources: CGPDFDictionaryRef?
    guard CGPDFDictionaryGetDictionary(dictionary,"Resources",&resources), let resources=resources else {return []}
    var dictionaryFonts: CGPDFDictionaryRef?
    guard CGPDFDictionaryGetDictionary(resources,"Font",&dictionaryFonts), let dictionaryFonts=dictionaryFonts else {return []}
    let result=NSMutableArray()
    CGPDFDictionaryApplyFunction(dictionaryFonts, { _,object,context in
        guard let context=context else {return}
        let result=Unmanaged<NSMutableArray>.fromOpaque(context).takeUnretainedValue()
        var font: CGPDFDictionaryRef?
        guard CGPDFObjectGetValue(object,.dictionary,&font),let font=font else {return}
        var name: UnsafePointer<CChar>?
        if CGPDFDictionaryGetName(font,"BaseFont",&name),let name=name {result.add(String(cString:name))}
    },Unmanaged.passUnretained(result).toOpaque())
    return Array(Set(result.compactMap{$0 as? String})).sorted()
}
func box(_ r: NSRect, page: NSRect) -> Rect {
    // PDF uses a bottom-left origin. Composer geometry uses a top-left origin.
    return Rect(x: Double(r.minX-page.minX), y: Double(page.maxY-r.maxY), width: Double(r.width), height: Double(r.height))
}
func fonts(_ a: NSAttributedString?) -> ([String],[Double]) {
    guard let a=a else { return ([],[]) }
    var names=Set<String>(); var sizes=Set<Double>()
    a.enumerateAttribute(.font, in: NSRange(location:0,length:a.length)) { value,_,_ in
        if let font=value as? NSFont { names.insert(font.fontName);sizes.insert(Double(font.pointSize)) }
    }
    return (names.sorted(),sizes.sorted())
}
do {
    guard CommandLine.arguments.count==3 else {throw NSError(domain:"reference-pdf",code:64,userInfo:[NSLocalizedDescriptionKey:"Usage: read-font-reference-pdf.swift input.pdf new-output.json"])}
    let url=URL(fileURLWithPath:CommandLine.arguments[1]).standardizedFileURL
    let output=URL(fileURLWithPath:CommandLine.arguments[2]).standardizedFileURL
    guard !FileManager.default.fileExists(atPath:output.path) else {throw NSError(domain:"reference-pdf",code:65,userInfo:[NSLocalizedDescriptionKey:"Output already exists"])}
    let data=try Data(contentsOf:url)
    guard let doc=PDFDocument(data:data) else {throw NSError(domain:"reference-pdf",code:66,userInfo:[NSLocalizedDescriptionKey:"Cannot read reference PDF"])}
    var pages=[Page]()
    for index in 0..<doc.pageCount {
        guard let page=doc.page(at:index) else {continue}
        let pageBox=page.bounds(for:.mediaBox)
        guard let selection=page.selection(for:pageBox) else {throw NSError(domain:"reference-pdf",code:67,userInfo:[NSLocalizedDescriptionKey:"Page \(index+1) has no text selection"])}
        var lines=[Line]();var union=NSRect.null;var fontNames=Set<String>()
        for line in selection.selectionsByLine() {
            let text=line.string ?? ""
            if text.trimmingCharacters(in:.whitespacesAndNewlines).isEmpty {continue}
            let bounds=line.bounds(for:page)
            let (names,sizes)=fonts(line.attributedString)
            names.forEach{fontNames.insert($0)}
            union=union.union(bounds)
            lines.append(Line(text:text,bounds:box(bounds,page:pageBox),font_names:names,font_sizes_pt:sizes))
        }
        guard !lines.isEmpty else {throw NSError(domain:"reference-pdf",code:68,userInfo:[NSLocalizedDescriptionKey:"Page \(index+1) has no visible text lines"])}
        pages.append(Page(page_number:index+1,width_pt:Double(pageBox.width),height_pt:Double(pageBox.height),lines:lines,bounds:box(union,page:pageBox),font_names:resourceFonts(page),selection_font_names:fontNames.sorted()))
    }
    let report=Report(pdf_file:url.path,pdf_sha256:SHA256.hash(data:data).map{String(format:"%02x",$0)}.joined(),page_count:doc.pageCount,pages:pages)
    let encoder=JSONEncoder();encoder.outputFormatting=[.prettyPrinted,.sortedKeys,.withoutEscapingSlashes]
    var result=try encoder.encode(report);result.append(10)
    try result.write(to:output,options:.withoutOverwriting)
    print("\(output.path): \(pages.count) pages")
} catch { fputs("read-font-reference-pdf: \(error)\n",stderr);exit(1) }
