package main


func Italics(message string) string {
	message = "*"+message+"*"
	return message
}

func Bold(message string) string {
	message = "**"+message+"**"
	return message
}

func BoldItalics(message string) string {
	message = "***"+message+"***"
	return message
}

func Underline(message string) string {
	message = "__"+message+"__"
	return message
}

func UnderlineItalics(message string) string {
	message = "__"+Italics(message)+"__"
	return message
}

func UnderlineBold(message string) string {
	message = "__"+Bold(message)+"__"
	return message
}

func UnderlineBoldItalics(message string) string {
	message = "__"+BoldItalics(message)+"__"
	return message
}

func CodeBlock(message string) string {
	return "`"+message+"`"
}

func CodeBox(message string) string {
	return "```\n"+message+"\n```"
}