package wa

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaezrr/remy-bot/internal/bot"
	"github.com/kaezrr/remy-bot/internal/config"
	"github.com/kaezrr/remy-bot/internal/store"
	"github.com/rs/zerolog/log"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waStore "go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	qrterminal "github.com/mdp/qrterminal/v3"
)

type BotHandleFunc func(input, prefix string, s store.Store) bot.Response

func sendGroupMessage(client *whatsmeow.Client, jid waTypes.JID, text string) {
	waMsg := &waE2E.Message{
		Conversation: proto.String(text),
	}
	if _, err := client.SendMessage(context.Background(), jid, waMsg); err != nil {
		log.Error().Err(err).Str("jid", jid.String()).Msg("failed to send group message")
	}
}

func Run(cfg *config.Config, s store.Store, handle BotHandleFunc) error {
	container, err := waStore.New(
		context.Background(),
		"sqlite",
		"file:"+cfg.SessionDir+"/whatsmeow.db?_pragma=foreign_keys(1)",
		nil,
	)
	if err != nil {
		return err
	}

	device, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return err
	}

	client := whatsmeow.NewClient(device, nil)

	readyC := make(chan struct{})

	client.AddEventHandler(func(evt any) {
		switch evt.(type) {
		case *events.Connected:
			log.Info().Msg("WhatsApp client connected and ready.")
			close(readyC) // Signal that we can proceed
		case *events.Disconnected:
			log.Error().Msg("WhatsApp client disconnected.")
		}
	})

	if client.Store.ID == nil {
		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			return err
		}

		if err := client.Connect(); err != nil {
			return err
		}

		for evt := range qrChan {
			switch evt.Event {
			case "code":
				log.Info().Msg("Scan the QR with your WhatsApp app")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			default:
				log.Info().Str("event", evt.Event).Msg("QR event")
			}
		}
	} else {
		if err := client.Connect(); err != nil {
			return err
		}
	}

	log.Info().Msg("Waiting for full connection and sync...")
	select {
	case <-readyC:
	case <-time.After(30 * time.Second):
		client.Disconnect()
		return fmt.Errorf("timeout waiting for WhatsApp connection after 30 seconds")
	}

	groups, err := client.GetJoinedGroups(context.Background())
	if err != nil {
		return err
	}

	var targetJID waTypes.JID
	for _, g := range groups {
		if g.GroupName.Name == cfg.TargetGroupName {
			targetJID = g.JID
			break
		}
	}
	if targetJID.IsEmpty() {
		return fmt.Errorf("target group %q not found", cfg.TargetGroupName)
	}

	log.Info().
		Str("group_name", cfg.TargetGroupName).
		Str("group_jid", targetJID.String()).
		Msg("target WhatsApp group resolved")

	sendGroupMessage(client, targetJID, "Remy has entered the chat. Type .h for help!")

	client.AddEventHandler(func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			handleIncomingMessage(client, v, cfg, s, handle, targetJID)
		}
	})

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
	<-sigC

	sendGroupMessage(client, targetJID, "Remy left the chat. See you soon!")

	log.Info().Msg("shutting down Remy, goodbye...")

	client.Disconnect()
	return nil
}

func handleIncomingMessage(
	client *whatsmeow.Client,
	msg *events.Message,
	cfg *config.Config,
	s store.Store,
	handle BotHandleFunc,
	targetJID waTypes.JID,
) {
	if msg.Info.MessageSource.IsFromMe {
		return
	}

	if !msg.Info.IsGroup || msg.Info.Chat != targetJID {
		return
	}

	text := msg.Message.GetConversation()
	if text == "" && msg.Message.ExtendedTextMessage != nil {
		text = msg.Message.ExtendedTextMessage.GetText()
	}
	if text == "" {
		return
	}

	resp := handle(text, cfg.Prefix, s)
	if resp.Text == "" {
		return
	}

	sendGroupMessage(client, targetJID, resp.Text)
}
