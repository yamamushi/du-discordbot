package main


/*
The command registry adds the ability for commands to be "protected" by permissions.

 Now I use quotes because there are no guarantees that a command may inadvertently kick off a process if permissions aren't checked properly.

 */


type CommandRecord struct {

	ID string `storm:"id"`
	Command string `storm:"index"`
	Groups []string `storm:"index"`


}