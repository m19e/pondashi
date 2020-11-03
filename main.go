package main

import (
	"fmt"
	"io/ioutil"
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

var Sounds = map[string]string{}

func init() {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "dev")
	}

	err := godotenv.Load(fmt.Sprintf(".env%s", os.Getenv("GO_ENV")))
	if err != nil {
		log.Fatal(err)
	}

	files, _ := ioutil.ReadDir("./sounds")

	for _, f := range files {
		filename := f.Name()
		Sounds[strings.Split(filename, ".")[0]] = filename
	}
}

var (
	Token      string
	GuildID    string
	TChannelID string
	VChannelID string
	Folder     string
	err        error

	dgv     *discordgo.VoiceConnection
	playing bool
	jobs    chan string
)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = "sounds"

	jobs = make(chan string, 10)

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
	if m.Author.ID == s.State.User.ID || m.Content == "" {
		return
	}

	log.Printf("[%s]", m.Content)

	if !checkCommand(m.Content[1:]) {
		if dgv == nil {
			return
		}

		sl := strings.Split(m.Content[2:], ":")
		stamp := sl[0]

		filename, ok := Sounds[stamp]
		if !ok {
			return
		}

		jobs <- filename

		if playing {
			return
		}

		playing = true
		for {
			j, ok := <-jobs
			if !ok {
				playing = false
				jobs = make(chan string, 10)
				break
			}
			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, j), make(chan bool))
		}
	} else {
		if m.Content == "!join" {
			dgv, err = s.ChannelVoiceJoin(GuildID, VChannelID, false, true)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		if dgv == nil {
			return
		}

		switch m.Content {
		case "!leave":
			err = dgv.Disconnect()
			if err != nil {
				log.Fatal(err)
			}

		case "!jihou":
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
