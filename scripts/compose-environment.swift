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
for style in [("regular", false, false), ("bold", true, false), ("italic", false, true), ("bold_italic", true, true)] {
    let base = CTFontCreateWithName("Arial" as CFString, 12, nil)
    let font: CTFont
    var traits: CTFontSymbolicTraits = []
    if style.1 { traits.insert(.traitBold) }
    if style.2 { traits.insert(.traitItalic) }
    if !traits.isEmpty {
        guard let result = CTFontCreateCopyWithSymbolicTraits(base, 12, nil, traits, traits) else { fatalError("Arial \(style.0) unavailable") }
        font = result
    } else { font = base }
    guard CTFontCopyFamilyName(font) as String == "Arial",
          let url = CTFontCopyAttribute(font, kCTFontURLAttribute) as? URL else { fatalError("Arial resolved to a substitute or inaccessible font") }
    fonts.append(["family": CTFontCopyFamilyName(font) as String,
                  "postscript_name": CTFontCopyPostScriptName(font) as String,
                  "style": style.0, "path": url.path])
}
let result: [String: Any] = ["os": ProcessInfo.processInfo.operatingSystemVersionString,
 "powerpoint_version": version, "powerpoint_build": build, "fonts": fonts]
let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
print(String(data: data, encoding: .utf8)!)
