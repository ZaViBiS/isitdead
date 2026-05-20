package bot

import (
	"context"
	"errors"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

type Client struct {
	bot *tgbot.Bot
}

func New(token string) (*Client, error) {
	b, err := tgbot.New(token)
	if err != nil {
		return nil, err
	}

	client := &Client{bot: b}
	client.registerHandlers()

	return client, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, message string) error {
	_, err := c.bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	})

	return err
}

func (c *Client) Start(ctx context.Context) {
	c.bot.Start(ctx)
}

func (c *Client) registerHandlers() {
	c.bot.RegisterHandler(tgbot.HandlerTypeMessageText, "start", tgbot.MatchTypeCommand, c.handleStart)
}

func (c *Client) handleStart(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	parts := strings.Fields(update.Message.Text)
	if len(parts) < 2 {
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Open the Telegram connection link from isitdead.cc to connect this chat.",
		})
		if err != nil {
			log.Err(err).Msg("error while sending message to user")
		}
		return
	}
	token := parts[1]
	if linkErr := sendToken(token, update.Message.Chat.ID); linkErr != nil {
		text := "Could not connect Telegram right now. Try again in a minute."
		if errors.Is(linkErr, ErrLinkRejected) {
			text = "This connection link is invalid or expired. Create a new Telegram link on the dashboard."
		}
		if errors.Is(linkErr, ErrMissingBaseURL) {
			text = "Telegram integration is not configured correctly. BASE_URL is missing."
		}
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
		})
		if err != nil {
			log.Err(err).Msg("error while sending message to user")
		}
		log.Err(linkErr).Msg("telegram link failed")
		return
	}

	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Telegram is connected to isitdead.cc. Alerts will appear here.",
	})
}
