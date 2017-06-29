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

	// Create / open our embedded database
	db, err := storm.Open(conf.DBConfig.DBFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()


	// Run a quick first time db configuration to verify that it is working properly
	fmt.Println("Checking Database")
	dbhandler := DBHandler{conf: &conf, rawdb: db}
	err = dbhandler.FirstTimeSetup()
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


	// Create a callback handler and add it to our Handler Queue
	fmt.Println("Adding Callback Handler")
	callbackhandler := CallbackHandler{dg: dg}
	dg.AddHandler(callbackhandler.Read)

	// Create our user handler
	fmt.Println("Adding User Handler")
	userhandler := UserHandler{conf: &conf, db: &dbhandler}
	userhandler.Init()
	dg.AddHandler(userhandler.Read)

	// Create our permissions handler
	fmt.Println("Adding Permissions Handler")
	permissionshandler := PermissionsHandler{dg: dg, conf: &conf, callback: &callbackhandler, db: &dbhandler, user: &userhandler}
	dg.AddHandler(permissionshandler.Read)

	// Create our command handler
	fmt.Println("Add Command Registry Handler")
	commandhandler := CommandHandler{dg: dg, db: &dbhandler, callback: &callbackhandler, user: &userhandler, conf: &conf, perm: &permissionshandler}
	commandhandler.Init()
	dg.AddHandler(commandhandler.Read)

	// Now we create and initialize our main handler
	handler := MainHandler{db: &dbhandler, conf: &conf, dg: dg, callback: &callbackhandler, perm: &permissionshandler, command: &commandhandler}
	err = handler.Init()
	if err != nil {
		fmt.Println("error in mainHandler.init", err)
		return
	}


	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
