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

- 🧹 **[autoremove_reactions](autoremove_reactions/)** — Extending an existing command to cover all advert channels (1x1 and group).

- 🛡️ **[2000_char_advert_fix](2000_char_advert_fix/)** — A fix (not a new command) for the long-form advert length check, which was wrongly deleting posts that were actually under Discord's 2,000-character limit. It switches the check to character counting with `toRune`.
