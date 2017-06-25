# du-discordbot

A Dual Universe bot being developed for the unofficial Dual Universe discord.

## Features

- [X] Embedded Database
- [X] Docker Support
- [ ] Internal Permissions System
- [ ] RSS Subscriptions
- [X] Currency System
- [ ] More Games!
- [ ] Reminders / Notifications
- [ ] Dual Universe Wiki Integration
- [ ] Dual Universe Resource Guide

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

To do this, I'll walk through the process of adding a "Hello" handler, and then the process of adding secondary handlers.


### Hello Handler

First, lets create our HelloHandler, which will listen for the string `hello` in every channel that our bot has access to. 

```go
package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type HelloHandler struct {}

func (h *HelloHandler) Read(s *discordgo.Session, m *discordgo.Message) {

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


Open up `func (h *MainHandler) Init() error`, which is defined within ``main_handler.go``, is our entry point for adding handlers into our _discordgo_ session.

You will see several examples of handlers being added to the queue, but let's skip those for now and find the line that reads

``// Add new handlers below this line //``

Below this line (obviously), let's create our `HelloHandler` and add it to the _discordgo_ session. 

```go
    // Add new handlers below this line //

    hello := HelloHandler{}
    h.dg.AddHandler(hello.Read)
```

That's it!

After the bot starts up, your handler will be added to the message reader queue, and when it encounters the configured string `~hello`, it will respond appropriately.







