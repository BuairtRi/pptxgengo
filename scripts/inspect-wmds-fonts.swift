// Independent CoreText resolution fingerprint. This does not prove which file
// PowerPoint used; native character names and PDF resources are separate evidence.
import Foundation
import CoreText
import CryptoKit

func fingerprint(_ font: CTFont) throws -> [String:Any] {
    guard let url = CTFontCopyAttribute(font,kCTFontURLAttribute) as? URL else {
        throw NSError(domain:"wmds-font",code:2,userInfo:[NSLocalizedDescriptionKey:"No font file URL"])
    }
    let data = try Data(contentsOf:url)
    var variations = [String:Double]()
    if let values = CTFontCopyVariation(font) as? [NSNumber:NSNumber] {
        for (key,value) in values { variations[key.stringValue] = value.doubleValue }
    }
    let axes = (CTFontCopyVariationAxes(font) as? [[String:Any]]) ?? []
    return ["postscript_name":CTFontCopyPostScriptName(font) as String,
        "family":CTFontCopyFamilyName(font) as String,"full_name":CTFontCopyFullName(font) as String,
        "style":CTFontCopyName(font,kCTFontStyleNameKey) as String? ?? "",
        "path":url.path,"sha256":SHA256.hash(data:data).map{String(format:"%02x",$0)}.joined(),
        "variations":variations,"variation_axes":axes,"variable":!axes.isEmpty]
}

do {
    let object = try JSONSerialization.jsonObject(with: FileHandle.standardInput.readDataToEndOfFile()) as! [String:Any]
    let requests = object["fonts"] as! [[String:Any]]
    let directory = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Fonts")
    var candidates = [[String:Any]]()
    for url in (try? FileManager.default.contentsOfDirectory(at:directory,includingPropertiesForKeys:nil)) ?? [] {
        guard ["ttf","otf","ttc"].contains(url.pathExtension.lowercased()),
              let descriptors = CTFontManagerCreateFontDescriptorsFromURL(url as CFURL) as? [CTFontDescriptor] else { continue }
        for descriptor in descriptors {
            let font = CTFontCreateWithFontDescriptor(descriptor,12,nil)
            // Metadata comes directly from the file, without globally registering it.
            candidates.append(try fingerprint(font))
        }
    }
    candidates.sort { (($0["path"] as! String)+($0["postscript_name"] as! String)) < (($1["path"] as! String)+($1["postscript_name"] as! String)) }
    var rows = [[String:Any]]()
    for request in requests {
        let name = request["pptx_typeface"] as! String
        let bold = request["pptx_bold"] as! Bool
        let italic = request["pptx_italic"] as! Bool
        var traits: CTFontSymbolicTraits = []
        if bold { traits.insert(.traitBold) }
        if italic { traits.insert(.traitItalic) }
        let base = CTFontCreateWithName(name as CFString,12,nil)
        guard let resolved = CTFontCreateCopyWithSymbolicTraits(base,12,nil,traits,[.traitBold,.traitItalic]),
              let url = CTFontCopyAttribute(resolved,kCTFontURLAttribute) as? URL else {
            throw NSError(domain:"wmds-font",code:1,userInfo:[NSLocalizedDescriptionKey:"Cannot resolve \(name)"])
        }
        let data = try Data(contentsOf:url)
        var variations = [String:Double]()
        if let axes = CTFontCopyVariation(resolved) as? [NSNumber:NSNumber] {
            for (key,value) in axes { variations[key.stringValue] = value.doubleValue }
        }
        let ps = request["postscript_name"] as! String
        let direct = try fingerprint(CTFontCreateWithName(ps as CFString,12,nil))
        let installed = candidates.filter { $0["postscript_name"] as! String == ps }
        let hashes = Set(installed.map { $0["sha256"] as! String })
        let state = hashes.count > 1 ? "competing_files" : (installed.isEmpty ? "no_matching_installed_descriptor" :
            (hashes.contains(request["sha256"] as! String) ? "bundled_static_bytes_present" : "different_file_or_variable_instance"))
        rows.append(["requested":request,"resolved_family":CTFontCopyFamilyName(resolved) as String,
                     "resolved_postscript":CTFontCopyPostScriptName(resolved) as String,
                     "resolved_path":url.path,"resolved_sha256":SHA256.hash(data:data).map{String(format:"%02x",$0)}.joined(),
                     "resolved_variations":variations,
                     "installed_file_candidates":installed, "installed_identity_state":state,
                     "postscript_resolution":direct, "postscript_resolution_matches":direct["postscript_name"] as! String == ps,
                     "native_file_identity_verified":false,
                     "postscript_matches":CTFontCopyPostScriptName(resolved) as String == request["postscript_name"] as! String,
                     "file_matches":SHA256.hash(data:data).map{String(format:"%02x",$0)}.joined() == request["sha256"] as! String])
    }
    let app = Bundle(path:"/Applications/Microsoft PowerPoint.app")!.infoDictionary!
    let result:[String:Any] = ["schema":"pptxgengo.wmds-font-environment.v1","os":ProcessInfo.processInfo.operatingSystemVersionString,
        "powerpoint_version":app["CFBundleShortVersionString"]!,"powerpoint_build":app["CFBundleVersion"]!,
        "identity_scope":"Independent CoreText selection by serialized family/bold/italic; not an observation of PowerPoint font file access",
        "installed_candidates_scope":"Read-only font file descriptors from ~/Library/Fonts; file presence is not native resolution proof",
        "fonts":rows]
    let data = try JSONSerialization.data(withJSONObject:result,options:[.prettyPrinted,.sortedKeys])
    FileHandle.standardOutput.write(data); print("")
} catch { fputs("inspect-wmds-fonts: \(error)\n",stderr); exit(1) }
