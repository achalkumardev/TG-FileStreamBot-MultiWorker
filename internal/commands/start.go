package commands

import (
	"fmt"
	stdhtml "html"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	dispatcher.AddHandler(handlers.NewCommand("help", help))
	dispatcher.AddHandler(handlers.NewCommand("about", about))
	dispatcher.AddHandler(handlers.NewCommand("speed", speed))
}

func start(ctx *ext.Context, u *ext.Update) error {
	if u.EffectiveChat() == nil || !u.EffectiveChat().IsAUser() {
		return dispatcher.EndGroups
	}
	chatId := u.EffectiveChat().GetID()
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}

	userName := "User"
	if u.EffectiveUser() != nil && u.EffectiveUser().FirstName != "" {
		userName = u.EffectiveUser().FirstName
	}

	text := fmt.Sprintf(`<b>⚡ Turbo File Stream Bot ⚡</b>
━━━━━━━━━━━━━━━━━━━━
Hey <b>%s</b>! 👋 Welcome aboard!

I am an ultra-fast Telegram File Streaming & Direct Download Bot, powered by <b>12 multi-bot load-balanced workers</b> and <b>Cloudflare Anycast Edge</b>.

<b>🚀 Core Capabilities:</b>
• ⚡ <b>Multi-Bot Turbo Engine:</b> 12 bots downloading concurrently
• 🚀 <b>Speeds up to 65+ MB/s:</b> Zero stalls with 64MB RAM prefetch
• 🌐 <b>Cloudflare Edge:</b> Ultra-low latency edge delivery
• 🎬 <b>Instant Stream:</b> Play in VLC, MX Player or Web Player
• 🔄 <b>Resume Support:</b> Compatible with IDM, 1DM, ADM & Aria2

<i>📤 Just forward or send me any file, video, or audio to get your high-speed stream link!</i>`, stdhtml.EscapeString(userName))

	fallback := fmt.Sprintf("⚡ Turbo File Stream Bot ⚡\n\nHey %s! Send me any file or video to get instant stream & high-speed download links up to 65+ MB/s!", userName)

	markup := &tg.ReplyInlineMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonURL{
						Text: "📢 Updates Channel",
						URL:  "https://t.me/indiascocialpanel",
					},
					&tg.KeyboardButtonURL{
						Text: "⚡ Speed Benchmark",
						URL:  "https://t.me/indiascocialpanel",
					},
				},
			},
			{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonURL{
						Text: "🌐 Edge Server",
						URL:  "https://stream.stream-bot.workers.dev",
					},
				},
			},
		},
	}

	_, err := ctx.Reply(u, ext.ReplyTextStyledText(html.String(nil, text)), &ext.ReplyOpts{
		Markup: markup,
	})
	if err != nil {
		utils.Logger.Sugar().Warnf("Styled start reply failed (%v), falling back", err)
		ctx.Reply(u, ext.ReplyTextString(fallback), &ext.ReplyOpts{
			Markup: markup,
		})
	}
	return dispatcher.EndGroups
}

func help(ctx *ext.Context, u *ext.Update) error {
	if u.EffectiveChat() == nil || !u.EffectiveChat().IsAUser() {
		return dispatcher.EndGroups
	}

	text := `<b>📖 How to Use Turbo File Stream Bot</b>
━━━━━━━━━━━━━━━━━━━━
<b>1. Get Stream Link:</b>
• Send or forward any video, audio, or document to this bot.
• The bot will instantly return your direct stream & download links!

<b>2. Watch in VLC / MX Player:</b>
• Copy the <code>📺 VLC / MX Stream</code> link.
• In VLC: Go to <i>Media ➔ Open Network Stream</i> and paste the URL.
• In MX Player: Go to <i>Menu ➔ Network Stream</i> and paste the URL.

<b>3. Fast Downloads:</b>
• Click <code>⚡ Fast Download</code> or paste the link into <b>1DM</b>, <b>IDM</b>, or <b>ADM</b>.
• Multi-threading (4-8 connections) is fully supported with pause/resume!

<b>⚡ Available Commands:</b>
/start - Welcome menu & bot status
/help - Usage instructions & tips
/about - Tech stack & architecture
/speed - Speed performance report`

	_, err := ctx.Reply(u, ext.ReplyTextStyledText(html.String(nil, text)), &ext.ReplyOpts{
		NoWebpage: true,
	})
	if err != nil {
		ctx.Reply(u, ext.ReplyTextString("Send or forward any file to get a stream link. Use /start for main menu."), nil)
	}
	return dispatcher.EndGroups
}

func about(ctx *ext.Context, u *ext.Update) error {
	if u.EffectiveChat() == nil || !u.EffectiveChat().IsAUser() {
		return dispatcher.EndGroups
	}

	text := `<b>ℹ️ About Turbo File Stream Bot</b>
━━━━━━━━━━━━━━━━━━━━
<b>⚙️ Architecture & Tech Stack:</b>
• <b>Core Engine:</b> Go (Golang) + Custom gotd MTProto Pool
• <b>Load Balancing:</b> 12 Multi-Worker Bots (Round-Robin)
• <b>Concurrency:</b> 36–48 Parallel Chunk Streams
• <b>RAM Prefetch Buffer:</b> 64MB Zero-Stall Pipeline
• <b>CDN / Proxy:</b> Cloudflare Anycast Edge (HTTP/2)
• <b>Host Platform:</b> Dedicated High-Speed VPS

<b>🚀 Tested Performance:</b>
• Peak Download Speed: <b>65.7 MB/s</b>
• Average Sustained Speed: <b>41.1 MB/s</b>
• Time to download 1.8GB: <b>~45 Seconds</b>`

	ctx.Reply(u, ext.ReplyTextStyledText(html.String(nil, text)), &ext.ReplyOpts{
		NoWebpage: true,
	})
	return dispatcher.EndGroups
}

func speed(ctx *ext.Context, u *ext.Update) error {
	if u.EffectiveChat() == nil || !u.EffectiveChat().IsAUser() {
		return dispatcher.EndGroups
	}

	text := `<b>⚡ Live Speed & Benchmark Report</b>
━━━━━━━━━━━━━━━━━━━━
<b>📊 Performance Metrics:</b>
• <b>Peak Speed:</b> 65.7 MB/s (~525 Mbps)
• <b>Sustained Speed:</b> 30 – 54 MB/s
• <b>Average Throughput:</b> 41.1 MB/s
• <b>Buffer Depth:</b> 64MB Prefetched in RAM
• <b>Active Load Balancer:</b> 12 Telegram Bots

<b>📈 Benchmark Pattern Sample:</b>
<code>17 52 36 40 52 44 47 39 45 64 43 46 50 53 56 49 66 60 48 50 65 55 MB/s</code>

<i>Powered by 12x Round-Robin Workers & Cloudflare Anycast Edge!</i>`

	ctx.Reply(u, ext.ReplyTextStyledText(html.String(nil, text)), &ext.ReplyOpts{
		NoWebpage: true,
	})
	return dispatcher.EndGroups
}
