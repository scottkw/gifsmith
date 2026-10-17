# API Coverage — GIPHY (GIFs), GIPHY Clips, Klipy (native v1), The Finer Gifs Club

> Full coverage by default. Opt-outs are explicit, reasoned decisions. Phase 2 plans 02-01..02-06.
> Sources: 02-RESEARCH.md "Provider API Reference", REQUIREMENTS.md (v1, v2, Out of Scope), 02-CONTEXT.md decisions.

| capability | decision | reason |
|---|---|---|
| GIPHY GIFs: search endpoint (/v1/gifs/search) | INTEGRATE | |
| GIPHY GIFs: trending endpoint (/v1/gifs/trending) | INTEGRATE | |
| GIPHY GIFs: offset pagination with offset caps (4999 search, 499 trending) | INTEGRATE | |
| GIPHY GIFs: rating parameter (fixed pg-13) | INTEGRATE | |
| GIPHY GIFs: rating chosen by the user | OPT-OUT | not needed yet: user-facing content rating control is SRCH-09 in Phase 4; pg-13 is sent by default now |
| GIPHY GIFs: bundle=messaging_non_clips renditions (fixed_width, original) | INTEGRATE | |
| GIPHY GIFs: other renditions (downsized, still, webp, mp4, preview_gif) | OPT-OUT | not needed: grid uses fixed_width and copy uses original; still renditions for reduced motion are a Phase 4 grid setting |
| GIPHY GIFs: lang parameter (locale) | OPT-OUT | not needed yet: GIPHY locale is SETT-04 in Phase 4 |
| GIPHY GIFs: get GIF by ID (/v1/gifs/{id}) | OPT-OUT | not needed: favorites and recents render from saved metadata without API calls (LIBR-04, Phase 3) |
| GIPHY GIFs: get GIFs by IDs (/v1/gifs) | OPT-OUT | not needed: same reason as get by ID |
| GIPHY GIFs: random endpoint | OPT-OUT | explicitly out of scope: not part of the extension's feature set or any requirement |
| GIPHY GIFs: translate endpoint | OPT-OUT | explicitly out of scope: not part of the extension's feature set or any requirement |
| GIPHY GIFs: search suggestions, autocomplete, categories, trending search terms | OPT-OUT | not needed yet: category/suggestion chips and autocomplete are v2 (DIFF-02) |
| GIPHY GIFs: channel search | OPT-OUT | explicitly out of scope: no requirement |
| GIPHY Stickers: search, trending, random, translate, packs | OPT-OUT | explicitly out of scope: stickers are listed under Out of Scope (GIF recording/editing, stickers, AI) |
| GIPHY Emoji endpoints | OPT-OUT | explicitly out of scope: not the core workflow (stickers/emoji out of scope) |
| GIPHY Upload API | OPT-OUT | explicitly out of scope: no upload or GIF creation in v1 |
| GIPHY analytics pingbacks and random_id | OPT-OUT | explicitly out of scope: no telemetry or analytics pingbacks (REQUIREMENTS Out of Scope, privacy) |
| GIPHY Clips: search endpoint (/v1/clips/search) | INTEGRATE | |
| GIPHY Clips: trending endpoint (/v1/clips/trending) | INTEGRATE | |
| GIPHY Clips: access probe (one trending call with limit=1) | INTEGRATE | |
| GIPHY Clips: bundle=clips_grid_picker preview rendition (fixed_width) | INTEGRATE | |
| GIPHY Clips: MP4 video renditions for copy/download (360p–1080p, source) | OPT-OUT | not needed yet: Clips file handling as MP4 (720p default) is ACTN-09 in Phase 3 |
| GIPHY Clips: get clip by ID | OPT-OUT | not needed: favorites/recents render from saved metadata (Phase 3) |
| Klipy GIFs: search endpoint (/api/v1/{key}/gifs/search) | INTEGRATE | |
| Klipy GIFs: trending endpoint (/api/v1/{key}/gifs/trending) | INTEGRATE | |
| Klipy GIFs: page/per_page pagination with has_next | INTEGRATE | |
| Klipy GIFs: content_filter parameter (fixed medium) | INTEGRATE | |
| Klipy GIFs: file renditions sm.gif (grid) and hd.gif / md.gif (copy) | INTEGRATE | |
| Klipy GIFs: webp, jpg, mp4, webm and xs renditions; format_filter parameter | OPT-OUT | not needed: GIF tiles and GIF copy only; WebP tiles are an optional later optimization (RESEARCH A3) |
| Klipy GIFs: locale parameter | OPT-OUT | not needed yet: Klipy locale is SETT-04 in Phase 4 |
| Klipy GIFs: customer_id parameter | OPT-OUT | not needed: no per-user identity or analytics is sent (privacy, no accounts) |
| Klipy GIFs: get items by slug | OPT-OUT | not needed: favorites/recents render from saved metadata (Phase 3) |
| Klipy GIFs: categories, search suggestions, autocomplete | OPT-OUT | not needed yet: v2 differentiator DIFF-02 |
| Klipy GIFs: share trigger / view reporting | OPT-OUT | explicitly out of scope: no analytics or pingbacks; to be re-checked against Klipy's integration requirements in the Phase 4 Klipy clearance review |
| Klipy GIFs: recent items per customer | OPT-OUT | not needed: Gifsmith keeps its own local Recents (Phase 3) |
| Klipy advertisements (type "ad" items) | OPT-OUT | not needed: ads require a mobile browser user agent and Partner Panel enablement; items with GIF renditions (ads included) stay in place per Klipy no-filter rule; only items with no GIF rendition are skipped and logged (RESEARCH A4, WR-04) |
| Klipy Stickers API | OPT-OUT | explicitly out of scope: stickers are listed under Out of Scope |
| Klipy Clips API | OPT-OUT | explicitly out of scope: Clips come from GIPHY only per the provider list (SRCH-01/02) |
| Klipy Memes API | OPT-OUT | explicitly out of scope: not part of the extension's feature set |
| The Finer Gifs Club: search (/search with q, q.parser=simple, sort, size, start) | INTEGRATE | |
| The Finer Gifs Club: paging via hits.found and hits.start | INTEGRATE | |
| The Finer Gifs Club: media files (media.thefinergifs.club/{fileid}.gif) | INTEGRATE | |
| The Finer Gifs Club: browse by episode order (sort start_time asc with no query) | OPT-OUT | not needed: the provider has no trending list and the UI shows the S8 idle state for an empty query (SRCH-03 applies only to providers with trending) |
