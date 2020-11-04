package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
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

var rep = regexp.MustCompile(`<:([^:]+):\d+>`)

func main() {
	Token = os.Getenv("TOKEN")
	GuildID = os.Getenv("GUILD_ID")
	TChannelID = os.Getenv("TEXT_CHANNEL_ID")
	VChannelID = os.Getenv("VOICE_CHANNEL_ID")
	Folder = os.Getenv("FOLDER")

	files, _ := ioutil.ReadDir(Folder)

	for _, f := range files {
		filename := f.Name()
		Sounds[strings.Split(filename, ".")[0]] = filename
	}

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

	if !checkCommand(m.Content) {
		if dgv == nil {
			return
		}

		filename, ok := Sounds[checkStamp(m.Content)]
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

		case "!kaboom":
			gs, err := s.Guild(GuildID)

			var data discordgo.GuildChannelCreateData
			vc, err := s.State.Channel(VChannelID)

			data.Name = "爆破予定地"
			data.Type = 2
			data.ParentID = vc.ParentID

			c, _ := s.GuildChannelCreateComplex(GuildID, data)

			defer func() {
				dgv.Disconnect()

				_, err = s.ChannelDelete(c.ID)
				if err != nil {
					return
				}
			}()

			target, err := s.ChannelMessageSend(TChannelID, "Please follow me.")

			time.Sleep(3 * time.Second)

			s.ChannelVoiceJoin(GuildID, c.ID, false, true)

			time.Sleep(3 * time.Second)

			var convicts []*discordgo.User

			for _, vs := range gs.VoiceStates {
				if vs.ChannelID == VChannelID {
					member, _ := s.GuildMember(GuildID, vs.UserID)
					convicts = append(convicts, member.User)
					s.GuildMemberMove(GuildID, member.User.ID, c.ID)
					time.Sleep(250 * time.Millisecond)
				}
			}

			count, _ := s.ChannelMessageEdit(TChannelID, target.ID, "Start a countdown.")

			time.Sleep(3 * time.Second)

			for i := 5; i > 0; i-- {
				count, err = s.ChannelMessageEdit(TChannelID, count.ID, fmt.Sprintf("%d", i))
				time.Sleep(1 * time.Second)
			}

			count, _ = s.ChannelMessageEdit(TChannelID, count.ID, "https://media.giphy.com/media/146BUR1IHbM6zu/giphy.gif")

			_, err = s.ChannelMessageSend(TChannelID, fmt.Sprintf("See you, %s", createMentions(convicts)))

			dgvoice.PlayAudioFile(dgv, fmt.Sprintf("%s/%s", Folder, "kaboom.m4a"), make(chan bool))
		}
	}
}

func checkCommand(m string) bool {
	switch m {
	case "!join", "!leave", "!kaboom":
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

func createMentions(users []*discordgo.User) string {
	var mentions []string
	for _, u := range users {
		mentions = append(mentions, u.Mention())
	}
	return strings.Join(mentions, " ")
}

func checkStamp(m string) string {
	match := rep.FindStringSubmatch(m)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}
