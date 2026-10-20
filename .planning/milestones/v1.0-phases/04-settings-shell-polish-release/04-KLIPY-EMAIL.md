# Early user actions for Phase 4 (send / do these now)

Two things only you can do. Both take a while to come back, so start them before the release plan
(04-05) reaches its notarization checkpoint. The executor never sends this email.

---

## 1. Email to Klipy (written clearance for file actions, and ad labelling)

**To:** developers@klipy.com
**From:** Ken Scott
**Subject:** Gifsmith (macOS GIF picker): clearance for temporary local copies, and how to label ads

Hello Klipy team,

I'm building Gifsmith, a free, open-source macOS menu bar app for finding a GIF and copying or
pasting it into the app you're using (a port of the MIT-licensed Raycast "GIF Search"
extension). Klipy is one of its providers. Each user brings their own Klipy API key from the
Partner Panel, and every request goes from the user's Mac to your API. Klipy results sit in their
own tab, never mixed with other providers, in the order you return them, unfiltered, and the
search field reads "Search KLIPY".

Before the 1.0.0 release I'd like written clearance on two points from your integration
requirements.

**(a) Temporary local copies for Copy GIF / Paste GIF / Download**

To put an animated GIF on the macOS clipboard so that Slack, Messages and similar apps receive
it as a GIF, the app has to download the file and put a file reference on the pasteboard. For a
Klipy item, Gifsmith would:

- download the GIF from the exact URL your API returned (no URL changes) into the app's private
  cache folder (`~/Library/Caches/com.kenscott.gifsmith/staging`);
- keep only the most recent file. It is deleted on the next copy, 30 seconds after a paste, and
  when the app quits. Nothing from an earlier run is kept;
- never re-host, upload, mirror, index or share the file. It exists only so the user's own paste
  works;
- for "Download", save the one GIF the user explicitly asked for into a folder they chose (the
  same as saving it from a browser).

Is this acceptable under your "do not store, mirror, re-host or retain copies" rule? Until I have
a written answer, **Gifsmith 1.0.0 ships with all Klipy file actions disabled**: on Klipy items,
Enter copies the GIF's link and Cmd+Enter pastes the link, and Copy GIF, Paste GIF, the square
variants and Download are hidden.

**(b) Ad items in the grid**

Some search and trending pages include ad items. Today they appear as ordinary tiles with no
label. Some of them are hosted outside klipy.com, and copying those fails. How should ads be
presented? For example, a "Sponsored" or "Ad" label on the tile, a required click-through
behaviour, or anything else you need. I'll follow your guidance exactly.

Thank you. I'm happy to share a build or screenshots.

Ken Scott
Gifsmith: https://github.com/scottkw/gifsmith

---

## 2. Apple Developer agreement (needed before notarization)

Notarization is blocked by an HTTP 403 from Apple until the current agreement is accepted. Plan
04-05 stops at a blocking checkpoint for this. Doing it now avoids the wait.

1. As the Account Holder, sign in at https://developer.apple.com/account and accept the latest
   Apple Developer Program License Agreement if it asks.
2. In App Store Connect, open **Agreements, Tax, and Banking** and accept any pending agreement.
3. Store the notarization credentials once (you'll be asked for an app-specific password):

   ```bash
   xcrun notarytool store-credentials gifsmith-notary --apple-id {your Apple ID} --team-id S2K7P43927
   ```

4. Optional check: `xcrun notarytool history --keychain-profile gifsmith-notary` should list your
   submissions (or none) rather than return a 403.

If the 403 remains after accepting, a lapsed membership may need Apple Developer Support.
