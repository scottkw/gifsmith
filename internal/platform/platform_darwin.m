//go:build darwin && !server

// Manual reference counting (cgo does not enable ARC).
#import <Cocoa/Cocoa.h>
#include <string.h>
#include "platform_darwin.h"

// Main-thread hop for AppKit window work. Same rationale as Wails mainthread_darwin.go:
// schedule on the main run loop in common modes, never dispatch_sync to the main queue
// (it stalls while a menu is tracking).
static void onMain(void (^block)(void)) {
    if ([NSThread isMainThread]) { block(); return; }
    dispatch_semaphore_t done = dispatch_semaphore_create(0);
    CFRunLoopRef loop = CFRunLoopGetMain();
    CFRunLoopPerformBlock(loop, kCFRunLoopCommonModes, ^{ block(); dispatch_semaphore_signal(done); });
    CFRunLoopWakeUp(loop);
    dispatch_semaphore_wait(done, DISPATCH_TIME_FOREVER);
    dispatch_release(done);
}

// Empty name = the user's general pasteboard; otherwise a private named one (tests).
// Pasteboard calls run on the caller's thread: NSPasteboard is an IPC client of the
// pasteboard server, and `go test` has no main run loop for onMain to hop to.
static NSPasteboard *pasteboard(const char *pbName) {
    if (pbName[0] == '\0') return [NSPasteboard generalPasteboard];
    return [NSPasteboard pasteboardWithName:[NSString stringWithUTF8String:pbName]];
}

int gsCopyFile(const char *pbName, const char *path) {
    BOOL ok = NO;
    @autoreleasepool {
        NSPasteboard *pb = pasteboard(pbName);
        [pb clearContents];
        ok = [pb writeObjects:@[[NSURL fileURLWithPath:[NSString stringWithUTF8String:path]]]];
    }
    return ok ? 1 : 0;
}

// First line "types:<space-separated pasteboard types>", then one file path per line.
// Caller frees.
char *gsPasteboardFileURLs(const char *pbName) {
    char *out = NULL;
    @autoreleasepool {
        NSPasteboard *pb = pasteboard(pbName);
        NSMutableArray *lines = [NSMutableArray array];
        [lines addObject:[@"types:" stringByAppendingString:[[pb types] componentsJoinedByString:@" "]]];
        NSArray *urls = [pb readObjectsForClasses:@[[NSURL class]]
                                          options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
        for (NSURL *u in urls) [lines addObject:u.path];
        out = strdup([[lines componentsJoinedByString:@"\n"] UTF8String]);
    }
    return out;
}

void gsReleasePasteboard(const char *pbName) {
    @autoreleasepool { [pasteboard(pbName) releaseGlobally]; }
}

// A non-activating Wails panel never makes its WKWebView first responder, and the panel's
// keyDown: swallows keys when the window itself is first responder (research Pitfall 1).
// Wails' notch window does the same makeFirstResponder for immediate keyboard input.
static NSView *findWebView(NSView *v) {
    if ([v isKindOfClass:NSClassFromString(@"WKWebView")]) return v;
    for (NSView *sub in v.subviews) {
        NSView *found = findWebView(sub);
        if (found) return found;
    }
    return nil;
}

void gsFocusWebView(void *nsWindow) {
    onMain(^{
        NSWindow *w = (NSWindow *)nsWindow;
        NSView *wv = findWebView(w.contentView);
        if (wv) [w makeFirstResponder:wv];
    });
}

// Visible frame (minus menu bar and Dock) of the screen under the mouse, converted to Wails
// window coordinates: Y down from the top of the primary screen, as windowSetPosition expects.
void gsCursorScreenVisibleFrame(int *x, int *y, int *w, int *h) {
    onMain(^{
        NSPoint p = [NSEvent mouseLocation];
        NSScreen *screen = [NSScreen mainScreen];
        for (NSScreen *s in [NSScreen screens]) {
            if (NSMouseInRect(p, s.frame, NO)) { screen = s; break; }
        }
        NSRect vf = screen.visibleFrame;
        CGFloat primaryHeight = [NSScreen screens][0].frame.size.height;
        *x = (int)vf.origin.x;
        *y = (int)(primaryHeight - (vf.origin.y + vf.size.height));
        *w = (int)vf.size.width;
        *h = (int)vf.size.height;
    });
}
