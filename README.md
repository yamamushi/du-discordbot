Table of Contents
=================

   * [Table of Contents](#table-of-contents)
   * [du-discordbot](#du-discordbot)
      * [Features](#features)
      * [Commands](#commands)
         * [Admin Commands](#admin-commands)
         * [User Commands](#user-commands)
      * [Permissions](#permissions)
         * [Ranks](#ranks)
         * [User Permission Commands](#user-permission-commands)
         * [Command Permission Commands](#command-permission-commands)
      * [Discord](#discord)
   * [Developers Guide](#developers-guide)
      * [Docker](#docker)
      * [Adding Commands](#adding-commands)
         * [Hello Handler](#hello-handler)
            * [Enabling HelloHandler](#enabling-hellohandler)
         * [Hello Handler Sub-Callbacks](#hello-handler-sub-callbacks)
            

# du-discordbot

A Dual Universe bot being developed for the unofficial Dual Universe discord.

## Features

- [X] Embedded Database
- [X] Docker Support
- [X] Internal User Permissions System
- [ ] Internal Command Permissions System
- [X] RSS Subscriptions
- [X] Currency System
- [ ] More Games!
- [ ] Reminders / Notifications
- [ ] Dual Universe Wiki Integration
- [ ] Dual Universe Resource Guide


## Commands

_du-discordbot_ maintains its own internal permissions system. It is important to note that these commands are not attributed to discord based roles. Ranks can therefore be assigned through _du-discordbot_.


### Admin Commands


| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| rss add  | Adds a new feed to the channel subscriptions  | ~rss add http://example.com/feed.rss |
| rss get  | Retrieves the latest RSS Items for the current channel  | ~rss get |
| rss list  | Lists the current channel subscriptions | ~rss list  |


### User Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| balance  | Lists your current credits balance  | ~balance |
| balance <user> | Gets balance for selected user  | ~balance @yamamushi |
| transfer <amount> <user> | Transfers credits to selected user | ~transfer 100 @yamamushi  |
| ping | Pings the bot (not a latency ping!) | ~ping |
| pong | Pongs the bot (not a latency pong!) | ~pong |


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

### Command Permission Commands

-Planned-
 

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

I won't go through the process of explaining the details of golang or the discordgo library, and you should definitely have a grasp of golang before proceeding. 

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

#### Enabling HelloHandler

Cool, we've got a simple handler built, but how do we actually get it to listen to incoming messages?


Open up `func (h *MainHandler) Init() error`, which is defined within ``main_handler.go``, our entry point for adding handlers into our _discordgo_ session.

You will see several examples of handlers being added to the queue, but let's skip those for now and find the line that reads

``// Add new handlers below this line //``

Below this line (obviously), let's create our `HelloHandler` and add it to the _discordgo_ session. 

```go
    // Add new handlers below this line //

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


Now back in ``main_handler.go``, we'll update the section you changed before to look like the following:

```go
    // Add new handlers below this line //
    
    // Here we add the global callback_handler as needed by HelloHandler 
    // If you get errors registering sub callbacks, make sure you've 
    // Constructed the parent handlers appropriately.
    hello := HelloHandler{conf: h.conf, callback: h.callback}
    h.dg.AddHandler(hello.Read)
    
```


Now when a user enters (for example) ``~prompt``, they will get prompted for input (Y or N in this example). The callback will get queued, and be triggered the next time that user talks in the channel. 

These callbacks will be reset when the bot is restarted, so no need to worry about prompts sticking around in the queue for eternity.
 
 