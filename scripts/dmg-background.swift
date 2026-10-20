// Renders the DMG window background to the icon layout `wails3 tool package
// --format dmg` uses at 540×380: app icon centred at (151,164), Applications at
// (388,164), 96 px icons (Finder coordinates, origin top-left).
// Run from the repo root: swift scripts/dmg-background.swift
import AppKit

let width: CGFloat = 540, height: CGFloat = 380

func color(_ hex: Int) -> NSColor {
    NSColor(srgbRed: CGFloat((hex >> 16) & 0xFF) / 255, green: CGFloat((hex >> 8) & 0xFF) / 255,
            blue: CGFloat(hex & 0xFF) / 255, alpha: 1)
}

func render(scale: Int, to path: String) {
    let rep = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: Int(width) * scale, pixelsHigh: Int(height) * scale,
                               bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
                               colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0)!
    rep.size = NSSize(width: width, height: height) // draw in points; @2x gets twice the pixels
    NSGraphicsContext.saveGraphicsState()
    NSGraphicsContext.current = NSGraphicsContext(bitmapImageRep: rep)

    color(0xF2F2F7).setFill()
    NSRect(x: 0, y: 0, width: width, height: height).fill()

    // AppKit's origin is bottom-left, so Finder's y = 164 is height - 164.
    let y = height - 164
    let arrow = NSBezierPath()
    arrow.move(to: NSPoint(x: 215, y: y))
    arrow.line(to: NSPoint(x: 324, y: y))
    arrow.move(to: NSPoint(x: 316, y: y + 7))
    arrow.line(to: NSPoint(x: 324, y: y))
    arrow.line(to: NSPoint(x: 316, y: y - 7))
    arrow.lineWidth = 2
    arrow.lineCapStyle = .round
    arrow.lineJoinStyle = .round
    color(0x86868B).setStroke()
    arrow.stroke()

    let text = NSAttributedString(string: "Drag Gifsmith to Applications", attributes: [
        .font: NSFont.systemFont(ofSize: 14),
        .foregroundColor: color(0x6E6E73),
    ])
    let size = text.size()
    text.draw(at: NSPoint(x: (width - size.width) / 2, y: 48 - size.height / 2))

    NSGraphicsContext.restoreGraphicsState()
    try! rep.representation(using: .png, properties: [:])!.write(to: URL(fileURLWithPath: path))
}

render(scale: 1, to: "build/darwin/dmg-background.png")
render(scale: 2, to: "build/darwin/dmg-background@2x.png")
