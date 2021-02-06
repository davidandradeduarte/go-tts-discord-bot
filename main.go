package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var token string
var buffer = make([][]byte, 0)

func init() {

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	if token == "" {
		fmt.Println("No token provided. Please provide the argument: -t <bot token>")
		return
		
	}
	// Load the sound file.
	//err := loadSound()
	//if err != nil {
	//	fmt.Println("Error loading sound: ", err)
	//	fmt.Println("Please copy $GOPATH/src/github.com/bwmarrin/examples/airhorn/airhorn.dca to this directory.")
	//	return
	//}
	
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
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
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Airhorn is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateGameStatus(0, "!airhorn")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	
	if m.Content[:6] == "speak " {
		resp, err := getVoiceFromText(m.Content[6:])

		if err != nil {
			fmt.Println(err.Error())
		}

		s.ChannelMessageSend(m.ChannelID, resp.URL)
		
		// Download mp3
		cmd := exec.Command("wget", resp.URL)

		err = cmd.Run()

		if err != nil {
			log.Fatal(err)
			return
		}

		myUrl, err := url.Parse(resp.URL)
		if err != nil {
			log.Fatal(err)
		}
		
		// TODO: convert mp3 to opus
		c1 := exec.Command("ffmpeg", "-i", path.Base(myUrl.Path), "-f", "wav", "-")
		
		c2 := exec.Command("opusenc", "--bitrate", "256", "-", "output.opus")

		c2.Stdin, _ = c1.StdoutPipe()
		c2.Stdout = os.Stdout
		_ = c2.Start()
		_ = c1.Run()
		_ = c2.Wait()
		
		// ffmpeg -i e7ad8c3b4325406e9d235618bf8d6ab9.mp3 -f wav - | opusenc --bitrate 256 - output.opus

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					fmt.Println("Error playing sound:", err)
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

	// Send the buffer data.
	//for _, buff := range buffer {
	//	vc.OpusSend <- buff
	//}

	// Start loop and attempt to play all files in the given folder
	//fmt.Println("Reading Folder: ", ".")
	//files, _ := ioutil.ReadDir(".")
	//for _, f := range files {
	//	fmt.Println("PlayAudioFile:", f.Name())
	//	//s.UpdateStatus(0, f.Name())
	//
	//	
	//}
	dgvoice.PlayAudioFile(vc, "output.opus", make(chan bool))

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
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
			//_, _ = s.ChannelMessageSend(channel.ID, "guild create func")
			return
		}
	}
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound() error {

	file, err := os.Open("output.opus")
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}
