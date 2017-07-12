package main

type Permissions struct{}

type CommandPermissions struct {
	ID      string `storm:"id"`
	command string `storm:"index"`
	channel string `storm:"index"`
	groups  []string
}
