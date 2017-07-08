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
      * [Wallet](#Wallet)         
      * [Bank](#Bank)
      * [Lua](#Lua)
      * [Chess](#Chess)
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
- [X] Internal Channels System
- [ ] Dual Universe Wiki Integration
- [ ] Dual Universe Resource Guide
- [ ] Dual Universe Forum Integration (half working with rss)
- [X] RSS Subscriptions
- [ ] Twitter Subscriptions (semi-working, needs polishing)
- [X] Currency System
- [ ] Bank System with Loans (bank works, no loans yet!)
- [ ] A Casino
- [X] Chess with selectable piece and board styles
- [ ] More Games!
- [ ] A Prize system for spending the credits from winning games
- [ ] Reminders / Notifications 
- [X] A Discord Lua Interpreter (is this the first discord lua parser?)
- [ ] MyCroft.ai Integration
- [ ] Random Utilities (too many to list here)


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
| groups | Shows groups the user belongs to | groups |
| ping | Pings the bot (not a latency ping!) | ping |
| pong | Pongs the bot (not a latency pong!) | pong |
| moon | Displays the current view of the moon from earth | moon |
| unshorten | Unshortens a bit.ly or similar url | unshorten <url> | 




## Wallet

Each user has a `wallet` that they carry with them. When users interact with games or other commands that accept credits, the funds will be taken from the wallet and never the users bank account.


### Wallet Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| wallet balance  | Lists your current credits balance  | wallet balance |
| wallet balance <user> | Gets balance for selected user  | wallet balance @yamamushi |
| wallet transfer <amount> <user> | Transfers credits to selected user | wallet transfer 100 @yamamushi  |



## Bank 

Running the `bank` command will launch the bank menu within a private chat.

The bank interface offers several commands for interacting with your bank account. While not currently an issue, banking your funds will protect them from being taken out of your wallet in the future. It also prevents users from seeing how much you have. 


### Bank Commands

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| bank | Opens the banking menu | bank |
| balance | Displays your bank account balance | balance |
| deposit | Deposits funds from your wallet | deposit 100 |
| withdraw | Withdraws funds from your bank account | withdraw 100 |
| transfer | Transfers funds to the provided bank account # | transfer 100 c9cfaa51-50b6-4f90-ada8-f938aadbfaba |
| rewards | Opens the credit rewards menu | rewards |
| loans | Opens the loans menu | loans |


You can also bypass the banking menu by running commands as arguments to the bank command, ie:

`bank deposit 15`



## Chess 

Chessbot will display a chess board and play a game of chess against users using the  [chess engine written by Jacob Roberts](https://github.com/JacobRoberts/chess).

Each game of chess costs 15 credits, and the funds go towards the total chess award pool for winning a game. 

Credits can be spent on chess piece styles as outlined in the chess menu.

| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| chess | Displays the chess program help menu| chess |
| chess styles | displays the chess styles help menu | chess styles | 



## Lua

Du-Discordbot has a builtin lua interpreter for testing lua code snippets. This is driven by [gopher-lua](https://github.com/yuin/gopher-lua).

It is important to follow the formatting guidelines of the lua interpreter, or your script will not be read correctly.


| Command       | Description   | Example Usage  |
| ------------- | ------------- | ------------- |
| lua | Displays the lua interpreter help menu| lua |


Lua input must be formatted as follows:

````go
~lua read ```
  function fib(n)
    if n == 0 then
      return 0
    elseif n == 1 then
      return 1
    end
    return fib(n-1) + fib(n-2)
  end
print(fib(10))
```
````

Note that the first line has three backticks included with the command, and the last line has three backticks on a line by itself. This formatting allows the parser to pull your code from the input and run it.




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


