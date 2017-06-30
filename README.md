Table of Contents
=================

   * [Table of Contents](#table-of-contents)
   * [du-discordbot](#du-discordbot)
      * [Features](#features)
      * [Commands](#commands)
         * [Admin Commands](#admin-commands)
         * [Moderator Commands](#moderator-commands)
         * [User Commands](#user-commands)
      * [Permissions](#permissions)
         * [Ranks](#ranks)
         * [User Permission Commands](#user-permission-commands)
         * [Command Permissions](#command-permissions)
      * [Discord](#discord)
   * [Developers Guide](#developers-guide)
      * [Docker](#docker)
      * [Adding Commands](#adding-commands)
         * [Hello Handler](#hello-handler)
         * [Enabling HelloHandler](#enabling-hellohandler)
         * [Hello Handler Sub-Callbacks](#hello-handler-sub-callbacks)
         * [Adding Hello Handler To The Command Registry](#adding-hello-handler-to-the-command-registry)
            

# du-discordbot

A Dual Universe bot being developed for the unofficial Dual Universe discord.

## Features

- [X] Embedded Database
- [X] Docker Support
- [X] Internal User Permissions System
- [X] Internal Command Permissions System
- [ ] Internal Channels System
- [ ] Dual Universe Wiki Integration
- [ ] Dual Universe Resource Guide
- [ ] Dual Universe Forum Integration
- [X] RSS Subscriptions
- [ ] Twitter Subscriptions
- [X] Currency System
- [ ] Bank System
- [ ] More Games!
- [ ] A Prize system for spending the credits from winning games
- [ ] Reminders / Notifications


## Commands

_du-discordbot_ maintains its own internal permissions system. It is important to note that these commands are not attributed to discord based roles. Ranks can therefore be assigned through _du-discordbot_.


### Admin Commands


| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| rss add  | Adds a new feed to the channel subscriptions  | ~rss add http://example.com/feed.rss |
| rss get  | Retrieves the latest RSS Items for the current channel  | ~rss get |
| rss list  | Lists the current channel subscriptions | ~rss list  |


### Moderator Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| groups <user>| Shows groups the user belongs to | ~groups @yamamushi |
| command enable | Enables a command in the current channel | command enable ping |
| command disable | Disables a command in the current channel | command enable ping |
| command list | Lists enabled commands for the current channel | command list |
| command usage | Displays usage for the supplied command | command usage ping | 
| command description | Displays description for the supplied command | command description ping | 



### User Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| balance  | Lists your current credits balance  | ~balance |
| balance <user> | Gets balance for selected user  | ~balance @yamamushi |
| transfer <amount> <user> | Transfers credits to selected user | ~transfer 100 @yamamushi  |
| ping | Pings the bot (not a latency ping!) | ~ping |
| pong | Pongs the bot (not a latency pong!) | ~pong |
| groups | Shows groups the user belongs to | ~groups |


## Permissions 

_du-discordbot_ uses an internal permission system for group assignment. 

Users are separated into groups that match the basic permissions system of the discord the bot was built for (see: [Discord](#discord)). 

Commands, likewise, can be registered with the command registry along with a permission level (defaults to "_citizen_")

If a command is run that is in the registry (the command registry is an extra feature that is not necessary when registering callbacks see: [Adding Commands](#adding-commands)), it will validate that a user has access to that command corresponding to their group, and return true or false. 
 
 The outcome of that check will determine the result of the attempted command. 
 
 tl;dr - _The things check the DB for permissions_

### Ranks

| Command       | Description   | Notes  |
| ------------- | ------------- | ------------- |
| owner | Configured owner of the bot, not the discord.  | Cannot be currently assigned |
| admin | Bot Administrators | |
| smoderator | Bot Senior Moderators  | |
| moderator | Bot Moderators | | 
| editor | Editors | -planned- |
| agora | Fans who make things | -planned- | 
| streamer | Streamers (Twitch/etc.) | -planned- | 
| recruiter | Recruiters | -planned-  | 
| citizen | Discord Citizen | All users whom the bot has seen speak have this role. |

### User Permission Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| promote | Promotes a user to the selected group  | ~promote @yamamushi admin |

### Command Permissions

The command registry is an internal permissions system for commands, which allows for limiting their usage x    as configured.

Permissions on commands can be registered in one of three ways:

| Permission | Description |
| ---------- | ----------- |
| channels | restricting a command's usage to specific channels | 
| groups | restricting a command's usage to specific groups |
| users | restricting a command's usage to specific users |


The three whitelist permissions can be managed with the following commands:

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| list | Lists the permissions for the chosen command | ~command users list ping |
| add | Adds a permission to the chosen command  | ~command groups add admin ping |
| remove | Removes a permission from the chosen command | ~command channels remove #general ping |


There is an additional command for listing the commands available in your current channel:

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| list | Lists the permissions for the current channel | ~command list |
 

## Discord

Join us on Discord @ [http://discord.me/dualuniverse](http://discord.me/dualuniverse)




# Developers Guide
**This library is under early stage active development so this guide subject to change**

## Docker

Launching this bot in docker is fairly straightforward.

1) Clone the repository

```git clone https://github.com/yamamushi/du-discordbot && cd du-discordbot```

2) Configure your configuration file as necessary.

```cp du-bot.conf.example du-bot.conf && vi du-bot.conf```

2) Create the docker container named du-discordbot

```docker build -t du-discordbot .```

3) Start the container with the name "du-discordbot"

```docker run --name dubot --rm du-discordbot```

4) To stop the container, open another console and run

```docker stop du-discordbot```


## Adding Commands

I won't go through the process of explaining the details of [_golang_](https://tour.golang.org/welcome/1) or the [_discordgo_](https://github.com/bwmarrin/discordgo) library, and you should definitely have a grasp of golang before proceeding. 

This should serve as a guide for how the control flow of the callback system works.

To do this, I'll walk through the process of adding a "Hello" handler, and then the process of adding sub callback handlers.


### Hello Handler

First, lets create our HelloHandler, which will listen for the string `hello` in every channel that our bot has access to. 

```go
package main

import (
	"fmt"
	
	"github.com/bwmarrin/discordgo"
)

type HelloHandler struct {
    conf *Config
}

func (h *HelloHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

    // Set our command prefix to the default one within our config file
	cp := h.conf.DUBotConfig.CP
	
	// If the message read is "hello" reply in the same channel with "Hi!"
	if m.Content == cp + "hello" {
		s.ChannelMessageSend( m.ChannelID, "Hi!")
	}	
}
```

That's it! 

As you can see, the process of adding more strings to listen for is fairly easy, just remember to take the command prefix into account.

You can save this file as `hello_handler.go`, I like the underscore in the filename because it's easy for me to distinguish handler files in a directory listing.

### Enabling HelloHandler

Cool, we've got a simple handler built, but how do we actually get it to listen to incoming messages?


Open up `func (h *MainHandler) Init() error`, which is defined within ``main_handler.go``, our entry point for adding handlers into our _discordgo_ session.

You will see several examples of handlers being added to the queue, but let's skip those for now and find the line that reads

``// Add new handlers below this line //``

Below this line (obviously), let's create our `HelloHandler` and add it to the _discordgo_ session. 

```go
    // Add new handlers below this line //

    // If you get errors registering handlers, make sure the function
    // Is defined correctly -> func(s *discordgo.Session, m *discordgo.MessageCreate)
    
    hello := HelloHandler{conf: h.conf}
    h.dg.AddHandler(hello.Read)
```

That's it!

After the bot starts up, your handler will be added to the message reader queue, and when it encounters the configured string `~hello`, it will respond appropriately.



### Hello Handler Sub-Callbacks


You may be wondering, "I've got a handler ready but how do I prompt for user input?". Well _du-discordbot_ has a callback handler built for this purpose. It's rough, but don't worry it's not the worst thing in the world.

Open up wherever you saved your HelloHandler, and lets modify that to look like this:

```go
package main

import (
	"fmt"
	"strings"
	
	"github.com/bwmarrin/discordgo"
)

type HelloHandler struct {

    callback *CallbackHandler
    conf *Config
}


func (h *HelloHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

    // Set our command prefix to the default one within our config file
	cp := h.conf.DUBotConfig.CP
	
	// If the message read is "hello" reply in the same channel with "Hi!"
	if m.Content == cp + "hello" {
		s.ChannelMessageSend( m.ChannelID, "Hi!")
	}
	if m.Content == cp + "prompt" {
	
		s.ChannelMessageSend(m.ChannelID, "Y or N?")
		
		// CallbackHandler.Watch expects a message, which it could
		// Also use to pass in arguments directly to a callback
		// That needs some setup or other parameters
		// In this example, our sub callback doesn't do anything
		// With the message it receives, so we can leave it blank
		message := "" 
		
		// (callback, uuid, message, session, messagecreate)
	    h.callback.Watch( h.Prompt, GetUUID(), message, s, m)	    
	}
}


// All sub-callbacks MUST have this function signature 
// func( string, *discordgo.Session, *discordgo.MessageCreate)
func (h *HelloHandler) Prompt(message string, s *discordgo.Session, m *discordgo.MessageCreate) {

    // Setup our command prefix
    cp := h.conf.DUBotConfig.CP
    // We want to cancel our command if another one is called by our user
    // We do this to avoid having duplicate/similar commands overrunning each other
    if strings.HasPrefix(m.Content, cp){
        s.ChannelMessageSend(m.ChannelID, "Prompt Cancelled")
        return
    }
	
	// Check 
    if m.Content == "Y" || m.Content == "y" {
        s.ChannelMessageSend(m.ChannelID, "You Selected Yes" )
        return
    }
    if m.Content == "N" || m.Content == "n" {
        s.ChannelMessageSend(m.ChannelID, "You Selected No" )
        return
    }

    s.ChannelMessageSend(m.ChannelID, "Invalid Response")

}

```


Now back in `main_handler.go`, we'll update the section you changed before to look like the following:

```go
    // Add new handlers below this line //
    
    // Here we add the global callback_handler as needed by HelloHandler 
    // If you get errors registering sub callbacks, make sure you've 
    // Constructed the parent handlers appropriately.
    hello := HelloHandler{conf: h.conf, callback: h.callback}
    h.dg.AddHandler(hello.Read)
    
```


Now when a user enters (for example) `~prompt`, they will get prompted for input (Y or N in this example). The callback will get queued, and be triggered the next time that user talks in the channel. 

These callbacks will be reset when the bot is restarted, so no need to worry about prompts sticking around in the queue for eternity.
 
 
 
### Adding Hello Handler To The Command Registry 

It is not required to use the command registry when adding handlers, however if you want to limit your command's usage to specific groups, channels, or users, you will need to register your command accordingly.

Please refer to [Command Permission Commands](#command-permission-commands) for further information about the permission groups and how to manage them from within discord. 

Once again, we're going to use `HelloHandler` as our example Handler, open it up and refer to the following:

```go
package main

import (
	"fmt"
	"strings"
	
	"github.com/bwmarrin/discordgo"
)

type HelloHandler struct {

    callback *CallbackHandler
    conf *Config
	user *UserHandler    
    registry *CommandRegistry
    db *DBHandler
}


func (h *HelloHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

    // Set our command prefix to the default one within our config file
	cp := h.conf.DUBotConfig.CP
	
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		//fmt.Println("Error finding user")
		return
	}	
	
	// If the message read is "hello", and permissions are valid, reply in the same channel with "Hi!"
	if m.Content == cp + "hello" {
	    if h.registry.CheckPermission("ping", m.ChannelID, user){
            s.ChannelMessageSend( m.ChannelID, "Hi!")
        }
	}
	if m.Content == cp + "prompt" {
	    if h.registry.CheckPermission("prompt", m.ChannelID, user){
            s.ChannelMessageSend(m.ChannelID, "Y or N?")
            
            // CallbackHandler.Watch expects a message, which it could
            // Also use to pass in arguments directly to a callback
            // That needs some setup or other parameters
            // In this example, our sub callback doesn't do anything
            // With the message it receives, so we can leave it blank
            message := "" 
            
            // (callback, uuid, message, session, messagecreate)
            h.callback.Watch( h.Prompt, GetUUID(), message, s, m)	    
	    }
	}
}


// All sub-callbacks MUST have this function signature 
// func( string, *discordgo.Session, *discordgo.MessageCreate)
func (h *HelloHandler) Prompt(message string, s *discordgo.Session, m *discordgo.MessageCreate) {

    // Setup our command prefix
    cp := h.conf.DUBotConfig.CP
    // We want to cancel our command if another one is called by our user
    // We do this to avoid having duplicate/similar commands overrunning each other
    if strings.HasPrefix(m.Content, cp){
        s.ChannelMessageSend(m.ChannelID, "Prompt Cancelled")
        return
    }
	
	// Check 
    if m.Content == "Y" || m.Content == "y" {
        s.ChannelMessageSend(m.ChannelID, "You Selected Yes" )
        return
    }
    if m.Content == "N" || m.Content == "n" {
        s.ChannelMessageSend(m.ChannelID, "You Selected No" )
        return
    }

    s.ChannelMessageSend(m.ChannelID, "Invalid Response")

}

```

In our example above we added three new pointers, one to `user` which is a pointer to the global User Handler, and the other for `registry` as a pointer to the global Command Registry (**not** the Command Handler!). 

We setup and retrieve our **current** user from the database (DB interaction will be a future tutorial) with :
```go
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil{
		//fmt.Println("Error finding user")
		return
	}	
```


For our `hello` command, we have added the necessary function call to the command registry `h.registry.CheckPermission(<command>, <channel>, <user>)` to check whether or not the given command has permission to run in the current channel:

```go
	if m.Content == cp + "hello" {
	    // <command>, <channelid>, <user>
	    if h.registry.CheckPermission("ping", m.ChannelID, user){
            s.ChannelMessageSend( m.ChannelID, "Hi!")
        }
	}
```



Open up `main_handler.go` and scroll down to where you added your `HelloHandler` before, modify it slightly to the following:


```go
    // Add new handlers below this line //
    
    // Here we add the global command registry (h.registry) as needed by HelloHandler
    // As well as the global user handler (h.user) and finally the db handler (h.db).
    // If you get errors registering sub callbacks, make sure you've 
    // Constructed the parent handlers appropriately.
    hello := HelloHandler{conf: h.conf, callback: h.callback, registry: h.registry, user: h.user, db: h.db}
    h.dg.AddHandler(hello.Read)
    
```


Now in `main_handler.go`, scroll down to `RegisterCommands()` and modify it to the following (it may look different as other commands were likely registered between the time of writing and the time you're looking at the file):

```go
func (h *MainHandler) RegisterCommands() (err error) {

	h.registry.Register("follow", "Follow a DU forum user. Updates will be sent via pm", "follow <forum name>")
	h.registry.Register("ping", "Ping command", "ping")
	h.registry.Register("pong", "Pong command", "pong")
	h.registry.Register("hello", "Hello command", "hello")
	h.registry.Register("prompt", "prompt command", "prompt")

	return nil

}
```

`h.registry.Register()` expects 3 arguments:
 
 `Command` - `Description` - `Usage Example`
 
 So in our above example for the `hello` command, we've added the description `Hello command` and the example usage has been set as `hello` (we omit our command prefix here as a rule of thumb).
 
 