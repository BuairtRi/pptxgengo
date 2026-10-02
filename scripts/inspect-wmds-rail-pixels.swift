#!/usr/bin/env swift
// Read-only pixel check of an existing 1920x1080 native PDF rendering.
import AppKit
import Foundation

do {
 guard CommandLine.arguments.count == 3 || CommandLine.arguments.count == 4 else {throw NSError(domain:"rail-pixels",code:64,userInfo:[NSLocalizedDescriptionKey:"Usage: inspect-wmds-rail-pixels.swift slide-012.png new-output.json [left|right]"])}
 let input=CommandLine.arguments[1], output=CommandLine.arguments[2]
 guard !FileManager.default.fileExists(atPath:output),let image=NSImage(contentsOfFile:input),let data=image.tiffRepresentation,let bitmap=NSBitmapImageRep(data:data),bitmap.pixelsWide==1920,bitmap.pixelsHigh==1080 else {throw NSError(domain:"rail-pixels",code:65,userInfo:[NSLocalizedDescriptionKey:"Expected new output and 1920x1080 PNG"])}
 func rgba(_ x:Int,_ y:Int)->[Double] {let c=bitmap.colorAt(x:x,y:y)!.usingColorSpace(.deviceRGB)!;return [c.redComponent,c.greenComponent,c.blueComponent,c.alphaComponent]}
 let side=CommandLine.arguments.count == 4 ? CommandLine.arguments[3] : "left"
 guard side == "left" || side == "right" else {throw NSError(domain:"rail-pixels",code:67,userInfo:[NSLocalizedDescriptionKey:"Expected left or right"])}
 let left=side == "left" ? 0 : 1374, right=side == "left" ? 545 : 1919
 let expected=rgba(side == "left" ? 10 : 1910,10)
 guard expected[0]<0.2 && expected[1]<0.2 && expected[2]<0.6 && expected[3]>0.99 else {throw NSError(domain:"rail-pixels",code:66,userInfo:[NSLocalizedDescriptionKey:"Reference pixel is not opaque navy"])}
 var total=0,bad=0
 func sample(_ x:Int,_ y:Int){total+=1;if zip(rgba(x,y),expected).contains(where:{abs($0.0-$0.1)>0.001}){bad+=1}}
 for y in 0..<1080 {sample(left,y);sample(right,y)}
 for x in left...right {sample(x,0);sample(x,1079)}
 let report:[String:Any]=["schema":"pptxgengo.wmds-rail-pixels.v1","input":input,"rail_width_pt":273,"side":side,"scale":2,"edge_pixels_checked":total,"edge_pixels_different_from_opaque_navy":bad,"passed":bad==0,"scope":"Outer raster rail perimeter only; PowerPoint editor boundary is not slide content"]
 try JSONSerialization.data(withJSONObject:report,options:[.prettyPrinted,.sortedKeys]).write(to:URL(fileURLWithPath:output),options:.withoutOverwriting)
 print("Rail perimeter: \(total) pixels checked, \(bad) different")
 if bad != 0 {exit(1)}
} catch {fputs("rail-pixels: \(error)\n",stderr);exit(1)}
