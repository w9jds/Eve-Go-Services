package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"

	"google.golang.org/api/option"

	"github.com/bwmarrin/discordgo"
)

var (
	ctx      = context.Background()
	discord  *discordgo.Session
	database *db.Client
)

type User struct {
	accessToken  string
	accountId    string
	email        string
	expiresAt    uint64
	id           string
	refreshToken string
	scope        string
	tokenType    string
	username     string
}

func main() {
	opt := option.WithCredentialsFile("path/to/serviceAccountKey.json")

	app, error := firebase.NewApp(ctx, nil, opt)
	if error != nil {
		fmt.Println("Error initializing firebase app: ", error)
	}

	database, error = app.Database(ctx)
	if error != nil {
		fmt.Println("Error fetching firebase client: ", error)
	}

	discord, error := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if error != nil {
		fmt.Println("Error creating discord client: ", error)
		return
	}

	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(memberAdded)

	error = discord.Open()
	if error != nil {
		fmt.Println("Error opening connection: ", error)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func ready(session *discordgo.Session, ready *discordgo.Ready) {
	fmt.Println("Aura has started! All systems green.")
}

func memberAdded(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	guildID := member.Member.GuildID
	userID := member.Member.User.ID

	var user User
	if error := database.NewRef("discord/"+userID).Get(ctx, &user); error != nil {
		var guestRoleID string
		roles, error := discord.GuildRoles(guildID)
		if error != nil {
			fmt.Println("Error getting guild roles", error)
		}

		for _, role := range roles {
			if role.Name == "Guest" {
				guestRoleID = role.ID
			}
		}

		discord.GuildMemberEdit(guildID, userID, []string{guestRoleID})
	}

}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	if strings.HasPrefix(strings.ToLower(message.Content), "!time") {
		utc := time.Now().UTC().Format("Monday, 02 January, 2006 15:04:05")
		session.ChannelMessageSend(message.ChannelID, utc)
	}
}
