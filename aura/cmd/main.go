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
	opt := option.WithCredentialsFile("../internal/config/neweden-admin.json")
	config := &firebase.Config{ProjectID: os.Getenv("PROJECT_ID"), DatabaseURL: os.Getenv("DATABASE_URL")}

	app, error := firebase.NewApp(ctx, config, opt)
	if error != nil {
		fmt.Println("Error initializing firebase app: ", error)
		return
	}

	database, error = app.Database(ctx)
	if error != nil {
		fmt.Println("Error fetching firebase client: ", error)
		return
	}

	discord, error = discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
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

func getMemberRoles(member *discordgo.GuildMemberAdd, user *User) []string {
	updatedList := []string{}

	roles, error := discord.GuildRoles(member.Member.GuildID)
	if error != nil {
		fmt.Println("Error getting guild roles", error)
		return updatedList
	}

	for _, role := range roles {
		if (user == nil || user.id == "") && role.Name == "Guest" {
			updatedList = append(updatedList, role.ID)
		}
	}

	return updatedList
}

func ready(session *discordgo.Session, ready *discordgo.Ready) {
	fmt.Println("Aura has started! All systems green.")
}

func memberAdded(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	var user User

	userID := member.Member.User.ID
	error := database.NewRef("discord/"+userID).Get(ctx, &user)

	if error != nil || user.id != userID {
		roles := getMemberRoles(member, &user)
		discord.GuildMemberEdit(member.Member.GuildID, userID, roles)
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
