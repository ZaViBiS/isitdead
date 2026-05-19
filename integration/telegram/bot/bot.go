package bot

import (
	"context"
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
			Text:   "no code?",
		})
		if err != nil {
			log.Err(err).Msg("error while sending messege to user")
		}
		return
	}
	token := parts[1]
	if err := sendToken(token, update.Message.Chat.ID); err != nil {
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "error 500, try again later",
		})
		if err != nil {
			log.Err(err).Msg("error while sending messege to user")
		}
	}

	_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hello! Bot is connected.",
	})
}
