package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var token string
var lock sync.Mutex
var messages int = 0

func init() {

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func initiLogger() {

	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	Formatter.ForceColors = true

	log.SetFormatter(Formatter)

	path := "logs/"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	writer, err := rotatelogs.New(
		fmt.Sprintf("%s%s.log", path, "%Y-%m-%d"),
		rotatelogs.WithMaxAge(time.Hour*72),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		log.Fatalf("Failed to Initialize Log File %s", err)
	}

	mw := io.MultiWriter(os.Stdout, writer)
	log.SetOutput(mw)
}

func main() {

	initiLogger()

	if token == "" {
		log.Fatal("No token provided. Please provide the argument: -t <bot token>")
		return

	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Info("Ready.")
	s.UpdateListeningStatus("speak")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	messages++
	defer func() {
		messages--
	}()

	lock.Lock()
	defer lock.Unlock()

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) < 5 {
		return
	}

	if len(m.Content) > 500 {
		SendMessage("can not send more than 500 characters", m.ChannelID, s)
		return
	}

	if m.Content[:5] == "speak" {
		msg := m.Content[5:]
		if strings.TrimSpace(msg) == "" {
			SendMessage("speak <your message>", m.ChannelID, s)
			return
		}

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Error("could not find the channel that the message came from", err.Error())
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.Error("could not find the guild for the channel", err.Error())
			return
		}

		connected := false
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				connected = true
			}
		}

		if !connected {
			log.Warn("not connected to a voice channel")
			SendMessage("not connected to a voice channel", m.ChannelID, s)
			return
		}

		log.Info("received text:", msg)
		resp, err := SynthesizeText(msg)
		if err != nil {
			log.Error("error obtaining voice from text", err.Error())
			return
		}

		file := path.Base(resp)

		// Convert mp3 to opus.
		c1 := exec.Command("ffmpeg", "-i", file, "-f", "wav", "-")
		c2 := exec.Command("opusenc", "--bitrate", "256", "-", "output.opus")

		c2.Stdin, _ = c1.StdoutPipe()
		c2.Stdout = os.Stdout
		_ = c2.Start()
		_ = c1.Run()
		_ = c2.Wait()

		// Delete downloaded file.
		err = os.Remove(file)

		if err != nil {
			log.Error("could not delete the downloaded file: ", file, err.Error())
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				vc, err := playSound(s, g.ID, vs.ChannelID)

				if messages-1 == 0 {
					vc.Disconnect()
				}

				if err != nil {
					log.Error("error playing sound", err.Error())
					return
				}
				return
			}
		}
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (voice *discordgo.VoiceConnection, err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	dgvoice.PlayAudioFile(vc, "output.opus", make(chan bool))

	// Delete opus file.
	err = os.Remove("output.opus")

	if err != nil {
		log.Error("could not delete output.opus file: ", err.Error())
		return vc, err
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specified amount of time before next message.
	time.Sleep(1 * time.Second)

	return vc, nil
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			log.Info("a new guild joined", channel.ID)
			return
		}
	}
}

func SendMessage(message string, channelId string, s *discordgo.Session) {
	err := s.ChannelTyping(channelId)
	if err != nil {
		log.Error("error triggering typing", err.Error())
		return
	}

	s.ChannelMessageSend(channelId, message)
}
