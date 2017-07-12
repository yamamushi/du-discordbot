# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/yamamushi/du-discordbot

# Create our shared volume
RUN mkdir /du-bot

# Run our dependency installation for Opus Encoding/Decoding
RUN apt-get update && \
        DEBIAN_FRONTEND=noninteractive apt-get install -y libav-tools opus-tools -f && \
        apt-get clean && \
        rm -rf /var/lib/apt/lists/


# Get the du-discordbot dependencies inside the container.
RUN go get github.com/bwmarrin/discordgo
RUN go get github.com/BurntSushi/toml
RUN go get github.com/asdine/storm
RUN go get gopkg.in/oleiade/lane.v1
RUN go get github.com/satori/go.uuid
RUN go get github.com/anaskhan96/soup
RUN go get github.com/lunixbochs/vtclean
RUN go get github.com/yuin/gopher-lua
RUN go get github.com/jonas747/ogg
RUN go get -u github.com/rylio/ytdl/...

# This is a fork of gofeed that allows for custom user-agent strings in requests to work with sites that filter these
RUN go get github.com/yamamushi/gofeed

# This is a fork of https://github.com/JacobRoberts/chess, to try and resolve issues with castling
RUN go get github.com/yamamushi/chess


# Install and run du-discordbot
RUN go install github.com/yamamushi/du-discordbot

# Run the outyet command by default when the container starts.
WORKDIR /du-bot
ENTRYPOINT /go/bin/du-discordbot

VOLUME /du-bot