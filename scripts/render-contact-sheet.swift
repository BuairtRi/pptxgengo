// Native PDF PNG contact sheets; no PowerPoint content is rasterized in the authored deck.
import AppKit
import Foundation
guard CommandLine.arguments.count == 3, let count = Int(CommandLine.arguments[2]), count > 0 else {
    fputs("Usage: swift scripts/render-contact-sheet.swift PNG-DIR PAGE-COUNT\n", stderr)
    exit(1)
}
let dir=URL(fileURLWithPath:CommandLine.arguments[1]); let out=dir.deletingLastPathComponent().appendingPathComponent("contact-sheets")
try FileManager.default.createDirectory(at:out,withIntermediateDirectories:true)
for start in stride(from:1,through:count,by:4) {
 let img=NSImage(size:NSSize(width:1920,height:1120));img.lockFocus()
 NSColor.lightGray.setFill();NSRect(x:0,y:0,width:1920,height:1120).fill()
 for (j,n) in Array(start...min(start+3,count)).enumerated() {
  let x=CGFloat(j%2)*960,y=CGFloat(1-j/2)*560
  let p=dir.appendingPathComponent(String(format:"slide-%03d.png",n));let v=NSImage(contentsOf:p)!
  v.draw(in:NSRect(x:x,y:y,width:960,height:540))
  ("Page \(n)" as NSString).draw(at:NSPoint(x:x+8,y:y+543),withAttributes:[.font:NSFont.systemFont(ofSize:14),.foregroundColor:NSColor.black])
 }
 img.unlockFocus();let b=NSBitmapImageRep(data:img.tiffRepresentation!)!;let data=b.representation(using:.jpeg,properties:[.compressionFactor:0.88])!
 try data.write(to:out.appendingPathComponent(String(format:"pages-%03d-%03d.jpg",start,min(start+3,count))))
}
print(out.path)
