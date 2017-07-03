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

	_ "net/http/pprof"
	"net/http"
)

// Variables used for command line parameters
var (
	ConfPath string
)

func init() {
	// Read our command line options
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

	fmt.Println("\n\n|| Starting du-discordbot ||\n")

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

	logger := Logger{}

	// Create a callback handler and add it to our Handler Queue
	fmt.Println("Adding Callback Handler")
	callbackhandler := CallbackHandler{dg: dg, logger: &logger}
	dg.AddHandler(callbackhandler.Read)

	// Create our user handler
	fmt.Println("Adding User Handler")
	userhandler := UserHandler{conf: &conf, db: &dbhandler, logger: &logger}
	userhandler.Init()
	dg.AddHandler(userhandler.Read)

	// Create our permissions handler
	fmt.Println("Adding Permissions Handler")
	permissionshandler := PermissionsHandler{dg: dg, conf: &conf, callback: &callbackhandler, db: &dbhandler,
		user: &userhandler, logger: &logger}
	dg.AddHandler(permissionshandler.Read)

	// Create our command handler
	fmt.Println("Adding Command Registry Handler")
	commandhandler := CommandHandler{dg: dg, db: &dbhandler, callback: &callbackhandler,
		user: &userhandler, conf: &conf, perm: &permissionshandler, logger: &logger}


	// Create our permissions handler
	fmt.Println("Adding Channel Permissions Handler")
	channelhandler := ChannelHandler{db: &dbhandler, conf: &conf, registry: commandhandler.registry,
		user: &userhandler, logger: &logger}
	channelhandler.Init()
	dg.AddHandler(channelhandler.Read)

	// Don't forget to initialize the command handler -AFTER- the Channel Handler!
	commandhandler.Init(&channelhandler)
	dg.AddHandler(commandhandler.Read)

	// Setup and initialize our Central Bank
	fmt.Println("Setting Up Bank")
	centralbank := Bank{db: &dbhandler, conf: &conf, user: &userhandler}
	centralbank.Init()

	// Create our Wallet Handler
	fmt.Println("Adding User Wallet Handler")
	wallethandler := WalletHandler{db: &dbhandler, conf: &conf, user: &userhandler, logger: &logger}
	dg.AddHandler(wallethandler.Read)

	// Create our Bank handler
	fmt.Println("Adding Bank Handler")
	bankhandler := BankHandler{db: &dbhandler, conf: centralbank.conf, com: &commandhandler, logger: &logger,
		user: &userhandler, callback: &callbackhandler, bank: &centralbank, wallet: &wallethandler}
	dg.AddHandler(bankhandler.Read)

	// Initalize our Logger
	fmt.Println("Initializing Logger")
	logger.Init(&channelhandler)

	// Now we create and initialize our main handler
	fmt.Println("\n|| Initializing Main Handler ||\n")
	handler := MainHandler{db: &dbhandler, conf: &conf, dg: dg, callback: &callbackhandler, perm: &permissionshandler,
		command: &commandhandler, logger: &logger, bankhandler: &bankhandler, user: &userhandler}
	err = handler.Init()
	if err != nil {
		fmt.Println("error in mainHandler.init", err)
		return
	}
	fmt.Println("\n|| Main Handler Initialized ||\n")


	if conf.DUBotConfig.Profiler {
		http.ListenAndServe(":8080", http.DefaultServeMux)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
