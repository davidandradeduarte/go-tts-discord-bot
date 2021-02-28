package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var token string
var buffer = make([][]byte, 0)
var lock sync.Mutex

func init() {

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

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
	fmt.Println("TTS Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Ready.")
	s.UpdateGameStatus(0, "speak")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	lock.Lock()
	defer lock.Unlock()

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) < 6 {
		return
	}

	if m.Content[:6] == "speak " {
		//resp, err := getVoiceFromText(m.Content[6:])

		resp, err := SynthesizeText(m.Content[6:])
		if err != nil {
			log.Println("error obtaining voice from text", err.Error())
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
			log.Println("could not delete the downloaded file: ", file, err.Error())
			return
		}

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Println("could not find the channel that the message came from", err.Error())
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.Println("could not find the guild for the channel", err.Error())
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					log.Println("error playing sound", err.Error())
				}
				return
			}
		}
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	dgvoice.PlayAudioFile(vc, "output.opus", make(chan bool))

	// Delete opus file.
	err = os.Remove("output.opus")

	if err != nil {
		log.Println("could not delete output.opus file: ", err.Error())
		return
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			log.Println("a new guild joined", channel.ID)
			return
		}
	}
}
