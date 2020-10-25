package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func init() {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "dev")
	}

	err := godotenv.Load(fmt.Sprintf(".env%s", os.Getenv("GO_ENV")))
	if err != nil {
		log.Fatal(err)
	}
}

var (
	Token      string
	GuildID    string
	TChannelID string
	VChannelID string
	Folder     string
	err        error

	dgv *discordgo.VoiceConnection
)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = "sounds"

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal(err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(messageCreate)

	fmt.Println("Pondashi is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// log.Printf("[%s]", m.Content)

	if !checkCommand(m.Content[1:]) {
		sl := strings.Split(m.Content[2:], ":")
		stamp := sl[0]
		if dgv == nil || !checkStamp(stamp) {
			return
		}
		dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, attachCodec(stamp)), make(chan bool))
	} else {
		switch m.Content {
		case "!join":
			dgv, err = s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
			if err != nil {
				log.Fatal(err)
			}

		case "!leave":
			if dgv == nil {
				return
			}
			err = dgv.Disconnect()
			if err != nil {
				log.Fatal(err)
			}

		case "!jihou":
			if dgv == nil {
				return
			}

			for _, num := range getCountsRing() {
				dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, fmt.Sprintf("Bell_use%d.ogg", num)), make(chan bool))
			}
		}
	}
}

func checkCommand(m string) bool {
	switch m {
	case "join", "leave", "jihou":
		return true
	}
	return false
}

func checkStamp(m string) bool {
	switch m {
	case "hnn_aha", "kit_pya", "hgc_empty", "askr_help", "lifesaver":
		return true
	}
	return false
}

func attachCodec(m string) string {
	switch m {
	case "hnn_aha", "kit_pya", "hgc_empty", "askr_help"	:
		return fmt.Sprintf("%s.%s", m, "m4a")
	case "lifesaver":
		return fmt.Sprintf("%s.%s", m, "mp3")
	}
	return "Bell_use1.ogg"
}

func getCountsRing() []int {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	hour := time.Now().In(jst).Hour()

	if hour == 0 {
		hour = 12
	}
	if hour > 12 {
		hour = hour - 12
	}

	slice := make([]int, 12)
	for i := range make([]int, hour) {
		slice = append(slice, (i%2)+1)
	}

	return slice
}
