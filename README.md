**This readme can also be found (arguably in a better format) on the [github wiki](https://github.com/yamamushi/du-discordbot/wiki).**

Table of Contents
=================

   * [du-discordbot](#du-discordbot)
      * [Features](#features)
      * [Commands](#commands)
         * [Owner Commands](#owner-commands)
         * [Admin Commands](#admin-commands)
         * [Moderator Commands](#senior-moderator-commands)
         * [Moderator Commands](#moderator-commands)
         * [User Commands](#user-commands)
      * [Permissions](#permissions)
         * [Ranks](#ranks)
         * [User Permission Commands](#user-permission-commands)
         * [Command Permissions](#command-permissions)
         * [Channel Permissions](#channel-permissions)
      * [Discord](#discord)
   * [Developers Guide](Developers.md)


# du-discordbot

A Dual Universe bot being developed for the [unofficial Dual Universe discord](https://discord.me/dualuniverse).

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


### Owner Commands


### Admin Commands


| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| rss add  | Adds a new feed to the channel subscriptions  | ~rss add http://example.com/feed.rss |
| rss get  | Retrieves the latest RSS Items for the current channel  | ~rss get |
| rss list  | Lists the current channel subscriptions | ~rss list  |


### Senior Moderator Commands


### Moderator Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| groups <user>| Shows groups the user belongs to | groups @yamamushi |
| command enable | Enables a command in the current channel or the target channel | command enable ping |
| command disable | Disables a command in the current channel or the target channel | command enable ping #general|
| command list | Lists enabled commands for the current channel or the target channel, accepts optional page number | command list #general 2|
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
 
 **Remember that if you want to enable a command in discord, you have to enable it for the channel and for a user group.**
 
 **Users must belong to this group to run a given command, even if it is enabled for a channel. Therefore it is useful to assign the "citizen" group to a command if you would like everyone to have access to it.**
 

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

### Permission Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| promote | Promotes a user to the selected group  | ~promote @yamamushi admin |
| demote | Demotes a user to the selected group  | ~demote @yamamushi moderator |
| command | Used to manage permissions on commands | [command permissions](#command-permissions) |
| channel | Used to manage permissions on commands | [channel permissions](#channel-permissions) |

## Command Permissions

 **By Default command permission commands are allowed by moderators and above**

The command registry is an internal permissions system for commands, which allows for limiting their usage as configured.

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
 

## Channel Permissions

 **By Default channel permission commands are allowed by smoderators and above**

The command registry is an internal permissions system for commands, which allows for limiting their usage as configured.

Permissions on commands can be registered in one of three ways:

| Permission | Description | Notes |
| ---------- | ----------- | ----- |
| botlog | Sets the channel to serve as the bot log channel | Unique Flag |
| permissionlog | Sets the channel to serve as the permission log channel | Unique Flag |
| banklog | Sets the channel to serve as the bank log channel | Unique Flag |
| hq | Sets the channel to serve as the hq channel | Unique Flag, Can only be set by owner |
| groups | restricting a command's usage to specific groups | |
| users | restricting a command's usage to specific users | |



The three whitelist permissions can be managed with the following commands:

| Command       | Description   | Example Usage  |
| ------------- | ------------- | -------------  |
| channel info | Displays permission info about the channel | channel info #general |
| channel group | Manages the group settings of the channel | channel group add admin #admin-channel |
| channel set | Sets a flag on the channel | channel set botlog #bot-channel |
| channel unset| Unsets a flag from the channel | channel unset botlog #general |
 

## Discord

Join us on Discord @ [http://discord.me/dualuniverse](http://discord.me/dualuniverse)


