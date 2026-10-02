// File/CoreText inspection only; does not open or automate PowerPoint.
// stdin: [{"family":"IBM Plex Sans","style":"regular"}, ...]
import Foundation
import CoreText

func fail(_ message: String) -> Never {
    FileHandle.standardError.write(Data((message + "\n").utf8))
    exit(1)
}
struct FontRequest: Decodable {
    let family: String
    let style: String
}
let requests: [FontRequest]
do {
    requests = try JSONDecoder().decode([FontRequest].self, from: FileHandle.standardInput.readDataToEndOfFile())
} catch {
    fail("Expected font requirements JSON on stdin: \(error)")
}
let app = "/Applications/Microsoft PowerPoint.app"
guard let info = Bundle(path: app)?.infoDictionary,
      let version = info["CFBundleShortVersionString"] as? String,
      let build = info["CFBundleVersion"] as? String else {
    fail("Microsoft PowerPoint bundle/version unavailable")
}
let families = Set(CTFontManagerCopyAvailableFontFamilyNames() as! [String])
let styleMask: CTFontSymbolicTraits = [.traitBold, .traitItalic]
var fonts: [[String: Any]] = []
for request in requests {
    guard families.contains(request.family) else {
        fail("Font family '\(request.family)' is not installed. Install it before native measurement.")
    }
    var traits: CTFontSymbolicTraits = []
    switch request.style {
    case "regular": break
    case "bold": traits.insert(.traitBold)
    case "italic": traits.insert(.traitItalic)
    case "bold_italic": traits = styleMask
    default: fail("Unknown font style '\(request.style)' for '\(request.family)'")
    }
    let base = CTFontCreateWithName(request.family as CFString, 12, nil)
    guard let font = CTFontCreateCopyWithSymbolicTraits(base, 12, nil, traits, styleMask),
          CTFontCopyFamilyName(font) as String == request.family,
          CTFontGetSymbolicTraits(font).intersection(styleMask) == traits,
          let url = CTFontCopyAttribute(font, kCTFontURLAttribute) as? URL else {
        fail("Font '\(request.family)' style '\(request.style)' is unavailable or resolves to a substitute.")
    }
    var entry: [String: Any] = ["family": CTFontCopyFamilyName(font) as String,
                               "postscript_name": CTFontCopyPostScriptName(font) as String,
                               "style": request.style, "path": url.path]
    if let variations = CTFontCopyVariation(font) as? [NSNumber: NSNumber], !variations.isEmpty {
        entry["variations"] = Dictionary(uniqueKeysWithValues: variations.map { ($0.key.stringValue, $0.value.doubleValue) })
    }
    fonts.append(entry)
}
let result: [String: Any] = ["os": ProcessInfo.processInfo.operatingSystemVersionString,
                           "powerpoint_version": version, "powerpoint_build": build, "fonts": fonts]
let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
print(String(data: data, encoding: .utf8)!)
