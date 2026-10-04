import AppKit
import ApplicationServices
import CoreGraphics
import Foundation

var result: [String: String] = ["status": "unknown", "detail": "PowerPoint dialog visibility is unavailable; no permission prompt is assumed"]
let applications = NSWorkspace.shared.runningApplications.filter { $0.bundleIdentifier?.lowercased() == "com.microsoft.powerpoint" }
func attribute(_ element: AXUIElement, _ name: String) -> CFTypeRef? {
    var value: CFTypeRef?
    guard AXUIElementCopyAttributeValue(element, name as CFString, &value) == .success else { return nil }
    return value
}
func inspect(_ element: AXUIElement, depth: Int, remaining: inout Int) {
    guard depth < 6, remaining > 0 else { return }; remaining -= 1
    for name in [kAXTitleAttribute, kAXDescriptionAttribute, kAXValueAttribute] {
        if let text = attribute(element, name) as? String {
            let lower = text.lowercased()
            if lower.contains("grant file access") || lower.contains("additional permissions are required") {
                result = ["status": "blocked", "dialog": "Grant File Access", "detail": text]
            }
        }
    }
    if let children = attribute(element, kAXChildrenAttribute) as? [AXUIElement] {
        for child in children { inspect(child, depth: depth + 1, remaining: &remaining) }
    }
}
if let app = applications.first {
    if AXIsProcessTrusted() {
        let element = AXUIElementCreateApplication(app.processIdentifier)
        AXUIElementSetMessagingTimeout(element, 0.2)
        if let windows = attribute(element, kAXWindowsAttribute) as? [AXUIElement] {
            result = ["status": "clear", "detail": "No file-access dialog observed in accessible PowerPoint windows"]
            var budget = 120
            for window in windows { inspect(window, depth: 0, remaining: &budget) }
        }
    }
    if result["status"] == "unknown", let windows = CGWindowListCopyWindowInfo([.optionOnScreenOnly, .excludeDesktopElements], kCGNullWindowID) as? [[String: Any]] {
        for window in windows where window[kCGWindowOwnerPID as String] as? Int32 == app.processIdentifier {
            if let title = window[kCGWindowName as String] as? String, title.lowercased().contains("grant file access") {
                result = ["status": "blocked", "dialog": "Grant File Access", "detail": title]
            }
        }
    }
}
let data = try JSONSerialization.data(withJSONObject: result)
print(String(decoding: data, as: UTF8.self))
