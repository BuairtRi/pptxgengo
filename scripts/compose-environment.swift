// File/CoreText inspection only; does not open or automate PowerPoint.
import Foundation
import CoreText
let app = "/Applications/Microsoft PowerPoint.app"
guard let info = Bundle(path: app)?.infoDictionary,
      let version = info["CFBundleShortVersionString"] as? String,
      let build = info["CFBundleVersion"] as? String else {
    fatalError("Microsoft PowerPoint bundle/version unavailable")
}
var fonts: [[String: String]] = []
for bold in [false, true] {
    let base = CTFontCreateWithName("Arial" as CFString, 12, nil)
    let font: CTFont
    if bold {
        guard let result = CTFontCreateCopyWithSymbolicTraits(base, 12, nil, .traitBold, .traitBold) else { fatalError("Arial Bold unavailable") }
        font = result
    } else { font = base }
    guard CTFontCopyFamilyName(font) as String == "Arial",
          let url = CTFontCopyAttribute(font, kCTFontURLAttribute) as? URL else { fatalError("Arial resolved to a substitute or inaccessible font") }
    fonts.append(["family": CTFontCopyFamilyName(font) as String,
                  "postscript_name": CTFontCopyPostScriptName(font) as String,
                  "style": bold ? "bold" : "regular", "path": url.path])
}
let result: [String: Any] = ["os": ProcessInfo.processInfo.operatingSystemVersionString,
 "powerpoint_version": version, "powerpoint_build": build, "fonts": fonts]
let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
print(String(data: data, encoding: .utf8)!)
