# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/yamamushi/du-discordbot
ADD du-bot.conf /go

# Build the du-discordbot inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/bwmarrin/discordgo
RUN go get github.com/BurntSushi/toml
RUN go get github.com/asdine/storm
RUN go install github.com/yamamushi/du-discordbot

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/du-discordbot


