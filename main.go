package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/asdine/storm"

)

// Variables used for command line parameters
var (
	Token string
	ConfPath string
)

func init() {
	flag.StringVar(&ConfPath, "c", "du-bot.conf", "Path to Config File")
	flag.Parse()

	_, err := os.Stat(ConfPath)
	if err != nil {
		log.Fatal("Config file is missing: ", ConfPath)
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	// Verify we can actually read our config file
	conf, err := ReadConfig(ConfPath)
	if err != nil {
		fmt.Println("error reading config file at: ", ConfPath)
		return
	}

	// Create or open our embedded database
	db, err := storm.Open(conf.DBConfig.DBFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()


	// Run a quick first time db configuration to verify that it is working properly
	dbhandler := DBHandler{DB: db, conf: &conf}
	err = dbhandler.Configure()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + conf.DiscordConfig.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	defer dg.Close()

	// Now we create and add our message handlers
	// Register the reader func as a callback for MessageCreate events.
	reader := MessageReader{db: &dbhandler, conf: &conf}
	dg.AddHandler(reader.read)

	rss := RSSHandler{db: &dbhandler, conf: &conf}
	dg.AddHandler(rss.menu)



	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}


	// Update our default playing status
	err = dg.UpdateStatus(0, conf.DUBotConfig.Playing)
	if err != nil {
		fmt.Println("error updating now playing,", err)
		return
	}


	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
