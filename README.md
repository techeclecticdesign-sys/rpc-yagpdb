# 🐉 RPC YAGPDB Bot

This project holds new bot code for the RPC YAGPDB bot. I am striving to automate some common staff tasks so that our current staff can focus on other aspects of managing the server.

## 🗺️ Layout

Each folder contains:

- 🖼️ **Screenshot images** showing what settings are used in the bot dashboard.
- 📜 **A text file** describing how to set up the code in the dashboard.
- 🪄 **A Go template file** containing the code for the bot, which is pasted into the dashboard.

## 📚 Commands

- ⚔️ **[quick_reactions](quick_reactions/)** — When a new post is made in the quick-search advert channels, a timer starts and, after a short delay, the bot checks that the post has at least three unique approved roleplay reactions. If it doesn't, the author is pinged in #rule_infractions and asked to add them. This keeps quick adverts properly tagged without staff hand-checking every post.

- 🔗 **[link_alerts](link_alerts/)** — Watches the quick-search advert channels for messages containing http/https links. When one is found, the bot pings the author in #rule_infractions and asks them to remove the link from their ad.

- 📜 **[header_alerts](header_alerts/)** — Watches the one-on-one advert channels for Discord header text (`#`, `##`, `###`), which isn't allowed there. When a header is used, the bot pings the author in #rule_infractions to remove it (regular bold is still fine).

- 🖼️ **[no_images_in_1x1](no_images_in_1x1/)** — Watches the 1x1 advert channels for posts with an uploaded image or file attachment, which aren't allowed there. When one is found, the bot pings the author in #rule_infractions and asks them to remove it from their ad.

- 📏 **[group_header_check](group_header_check/)** — Enforces the header rules in the group advert channels: only one header line is allowed per post, and if that is exceeded the author is pinged in #rule_infractions to fix it.

- 📨 **[alert_sender](alert_sender/)** — Shared helper used by `link_alerts`, `header_alerts`, `no_images_in_1x1`, and `group_header_check`. Instead of pinging directly, those commands hand their message to this one with a short delay; it then pings #rule_infractions only if the post still exists. This stops the author being pinged about a post that the length/cooldown/duplicate command already deleted, and keeps the #rule_infractions channel ID configured in one place.

- 🚫 **[banned_words](banned_words/)** — Watches the advert channels for a configurable list of banned words. If any word in a post contains a banned word as a substring (case-insensitive), the bot pings the author in #rule_infractions, links to the channel, and shows the offending word(s) in ||spoiler tags||. Routes through `alert_sender`.

- 🪪 **[cross_channel_dupes](cross_channel_dupes/)** — Enforces the "one advert, one channel" rule (o7). When someone posts, it compares the new ad against that user's current ads in the other advert channels (found in one database query via their stored `lastMsg_` keys) and, if it finds a verbatim/copy-paste match, pings the author in #rule_infractions to differentiate or remove one. Routes through `alert_sender`, so a post already deleted for length/cooldown won't trigger a duplicate ping.

- 🧹 **[autoremove_reactions](autoremove_reactions/)** — Extending an existing command to cover all advert channels (1x1 and group).

- 🛡️ **[2000_char_advert_fix](2000_char_advert_fix/)** — A fix (not a new command) for the long-form advert length check, which was wrongly deleting posts that were actually under Discord's 2,000-character limit. It switches the check to character counting with `toRune`.

- 🛡️ **[2000_char_group_advert_fix](2000_char_group_advert_fix/)** — The same length-check fix applied to the group advert command, which had the identical byte-counting bug.

- 🩹 **[member_intros_fix](member_intros_fix/)** — A fix (not a new command) for the member-intros command: it switches the length check to character counting, and fixes a bug where the duplicate-intro lock stopped working after a single collision.
