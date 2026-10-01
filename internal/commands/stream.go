package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	stdhtml "html"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/types"
	"github.com/dustin/go-humanize"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func getActiveHost() string {
	if envHost := os.Getenv("HOST"); envHost != "" && strings.Contains(envHost, "workers.dev") {
		return envHost
	}
	if config.ValueOf.Host != "" && strings.Contains(config.ValueOf.Host, "workers.dev") {
		return config.ValueOf.Host
	}

	re := regexp.MustCompile(`https://[-a-z0-9]+\.trycloudflare\.com`)
	urlFiles := []string{
		"logs/current_url.txt",
		"/app/logs/current_url.txt",
		"/logs/current_url.txt",
	}
	for _, f := range urlFiles {
		if data, err := os.ReadFile(f); err == nil {
			trimmed := strings.TrimSpace(string(data))
			if re.MatchString(trimmed) {
				return trimmed
			}
		}
	}

	paths := []string{"cloudflared.log", "/app/cloudflared.log", "/app/logs/cloudflared.log"}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			matches := re.FindAllString(string(data), -1)
			if len(matches) > 0 {
				return matches[len(matches)-1]
			}
		}
	}
	if envHost := os.Getenv("HOST"); envHost != "" {
		return envHost
	}
	return config.ValueOf.Host
}

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(
		handlers.NewMessage(nil, sendLink),
	)
}

func supportedMediaFilter(m *types.Message) (bool, error) {
	if not := m.Media == nil; not {
		return false, dispatcher.EndGroups
	}
	switch m.Media.(type) {
	case *tg.MessageMediaDocument:
		return true, nil
	case *tg.MessageMediaPhoto:
		return true, nil
	case tg.MessageMediaClass:
		return false, dispatcher.EndGroups
	default:
		return false, nil
	}
}

func sendLink(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}
	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}
	if !supported {
		ctx.Reply(u, ext.ReplyTextString("Sorry, this message type is unsupported."), nil)
		return dispatcher.EndGroups
	}
	update, err := utils.ForwardMessages(ctx, chatId, config.ValueOf.LogChannelID, u.EffectiveMessage.ID)
	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
		return dispatcher.EndGroups
	}
	if len(update.Updates) < 2 {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected update structure from Telegram"), nil)
		return dispatcher.EndGroups
	}
	msgIDUpdate, ok := update.Updates[0].(*tg.UpdateMessageID)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected update type"), nil)
		return dispatcher.EndGroups
	}
	messageID := msgIDUpdate.ID
	newMsg, ok := update.Updates[1].(*tg.UpdateNewChannelMessage)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected channel message update"), nil)
		return dispatcher.EndGroups
	}
	msg, ok := newMsg.Message.(*tg.Message)
	if !ok {
		ctx.Reply(u, ext.ReplyTextString("Error - unexpected message type"), nil)
		return dispatcher.EndGroups
	}
	doc := msg.Media
	file, err := utils.FileFromMedia(doc)
	if err != nil {
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
		return dispatcher.EndGroups
	}
	fullHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)
	hash := utils.GetShortHash(fullHash)
	host := getActiveHost()
	streamLink := fmt.Sprintf("%s/stream/%d?hash=%s", host, messageID, hash)
	watchLink := fmt.Sprintf("%s/watch/%d?hash=%s", host, messageID, hash)
	downloadLink := streamLink + "&d=true"
	fileSizeStr := humanize.Bytes(uint64(file.FileSize))

	msgText := fmt.Sprintf(`<b>⚡ FILE READY TO STREAM & DOWNLOAD ⚡</b>
━━━━━━━━━━━━━━━━━━━━━
📁 <b>Name:</b> <code>%s</code>
📦 <b>Size:</b> <code>%s</code>
⚙️ <b>Type:</b> <code>%s</code>
🚀 <b>Edge:</b> <i>Cloudflare Anycast (BOM)</i>
━━━━━━━━━━━━━━━━━━━━━
<i>Click buttons below to stream or download up to 65+ MB/s!</i>`,
		stdhtml.EscapeString(file.FileName), fileSizeStr, stdhtml.EscapeString(file.MimeType))

	var rows []tg.KeyboardButtonRow
	if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{
					Text: "▶️ Watch Online (Web Player)",
					URL:  watchLink,
				},
			},
		})
	}
	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "⚡ Fast Download",
				URL:  downloadLink,
			},
			&tg.KeyboardButtonURL{
				Text: "📺 VLC / MX Stream",
				URL:  streamLink,
			},
		},
	})

	markup := &tg.ReplyInlineMarkup{
		Rows: rows,
	}

	_, err = ctx.Reply(u, ext.ReplyTextStyledText(html.String(nil, msgText)), &ext.ReplyOpts{
		Markup:           markup,
		NoWebpage:        true,
		ReplyToMessageId: u.EffectiveMessage.ID,
	})
	if err != nil {
		utils.Logger.Sugar().Error(err)
		ctx.Reply(u, ext.ReplyTextString(fmt.Sprintf("Error - %s", err.Error())), nil)
	}
	return dispatcher.EndGroups
}
