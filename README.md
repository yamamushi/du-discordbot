# du-discordbot

A Dual Universe bot being developed for the unofficial Dual Universe discord.

## Discord

Join us on Discord @ [http://discord.me/dualuniverse](http://discord.me/dualuniverse)


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



## Features

- [X] Embedded Database
- [X] Docker Support
- [ ] Internal Permissions System
- [ ] RSS Subscriptions
- [ ] Currency System
- [ ] More Games!
- [ ] Reminders / Notifications
- [ ] Dual Universe Wiki Integration
- [ ] Dual Universe Resource Guide
