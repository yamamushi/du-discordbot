package main

// Italics function
func Italics(message string) string {
	message = "*" + message + "*"
	return message
}

// Bold function
func Bold(message string) string {
	message = "**" + message + "**"
	return message
}

// BoldItalics function
func BoldItalics(message string) string {
	message = "***" + message + "***"
	return message
}

// Underline function
func Underline(message string) string {
	message = "__" + message + "__"
	return message
}

// UnderlineItalics function
func UnderlineItalics(message string) string {
	message = "__" + Italics(message) + "__"
	return message
}

// UnderlineBold function
func UnderlineBold(message string) string {
	message = "__" + Bold(message) + "__"
	return message
}

// UnderlineBoldItalics function
func UnderlineBoldItalics(message string) string {
	message = "__" + BoldItalics(message) + "__"
	return message
}

// CodeBlock function
func CodeBlock(message string) string {
	return "`" + message + "`"
}

// CodeBox function
func CodeBox(message string) string {
	return "```\n" + message + "\n```"
}
