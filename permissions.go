package main

// Permissions struct
type Permissions struct{}

// CommandPermissions struct
type CommandPermissions struct {
	ID      string `storm:"id"`
	command string `storm:"index"`
	channel string `storm:"index"`
	groups  []string
}
