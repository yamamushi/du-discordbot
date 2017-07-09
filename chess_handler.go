package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"strconv"
	"fmt"
	"errors"
)

type ChessHandler struct {

	db *DBHandler
	logchan chan string
	conf *Config
	wallet *WalletHandler
	chess *ChessGame
	bank *BankHandler
	command *CommandRegistry
	user *UserHandler
}



func (h *ChessHandler) Init() (err error){
	h.chess = &ChessGame{db: h.db }
	h.chess.Init()


	if !h.bank.bank.CheckUserAccount("chessaccount"){
		fmt.Println("Creating Chess Bank Account ")
		err := h.bank.bank.CreateUserAccount("chessaccount")
		if err != nil {
			return err
		}
	}

	return nil
}


func (h *ChessHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){


	if !SafeInput(s, m, h.conf){
		return
	}

	command, payload :=  CleanCommand(m.Content, h.conf)

	if command != "chess" {
		return
	}

	user, err := h.user.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve user record: " + err.Error())
		return
	}

	if h.command.CheckPermission("chess",m.ChannelID, user) {

		if len(payload) < 1 {
			h.DisplayHelp(s, m)
			return
		}
		if len(payload) > 0 {

			if !h.chess.PlayerHasRecord(m.Author.ID) {
				err := h.chess.NewPlayerRecord(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not create new chess player record: "+err.Error())
					return
				}
			}

			command, payload = SplitPayload(payload)

			if command == "newgame" {
				if len(payload) < 1 {
					h.NewGame("white", s, m)
					return
				}
				h.NewGame(payload[0], s, m)
				return
			}
			if command == "plots" {
				if len(payload) < 1 {
					s.ChannelMessageSend(m.ChannelID, "<plots> expects an argument <enable> or <disable>")
					return
				}
				if payload[0] == "enable" {
					err := h.chess.EnablePlots(m.Author.ID)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Could not modify plots: "+err.Error())
						return
					}
					s.ChannelMessageSend(m.ChannelID, "Plots enabled")
					return
				}
				if payload[0] == "disable" {
					err := h.chess.DisablePlots(m.Author.ID)
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Could not modify plots: "+err.Error())
						return
					}
					s.ChannelMessageSend(m.ChannelID, "Plots disabled")
					return
				}
			}
			if command == "bank" && user.Admin {

				account, err := h.bank.bank.GetAccountForUser("chessaccount")
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not find Chess Bank Account")
					return
				}
				s.ChannelMessageSend(m.ChannelID, "Chess Bank Account #"+account.ID)
				return
			}
			if command == "move" {

				if len(payload) < 1 {
					s.ChannelMessageSend(m.ChannelID, "<move> expects an argument")
					return
				}

				err := h.chess.PlayerMove(m.Author.ID, payload[0])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not move: "+err.Error())
					return
				}

				message, err := h.chess.GetBoard(m.Author.ID, "default", "default")
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not display board: "+err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, "Player moved to "+payload[0]+message)
				h.BotMoveCallback(s, m)
				return
			}
			if command == "board" {
				h.SendBoard(s, m)
				return
			}
			if command == "fen" {
				message, err := h.chess.GetFen(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not display board: "+err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, message)
				return
			}
			if command == "styles" {
				h.DisplayStyleMenu(payload, s, m)
				return
			}
			if command == "resign" {
				h.SendBoard(s, m)
				err := h.chess.Resign(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not resign: "+err.Error())
					return
				}
				err = h.chess.ProcessLoss(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error Processing Loss: "+err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, "You have resigned from the game! Better luck next time!")
				return
			}
			if command == "load" {
				s.ChannelMessageSend(m.ChannelID, "Attempting to load saved game, this make take a while depending on the last board state. Attempting to view the board before it is loaded may result in temporary display errors")
				err := h.chess.LoadCurrentGame(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not load game: "+err.Error())
					return
				}
				h.SendBoard(s, m)
				s.ChannelMessageSend(m.ChannelID, "Game has been loaded, you may now continue.")
				return
			}
			if command == "profile" {
				if len(payload) < 1 {
					h.DisplayProfile(m.Author.ID, s, m)
					return
				}
				if len(m.Mentions) < 1 {
					s.ChannelMessageSend(m.ChannelID, "Invalid user selected")
					return
				}

				h.DisplayProfile(m.Mentions[0].ID, s, m)
				return
			}
			if command == "help" {
				h.DisplayHelp(s, m)
				return
			}
			if command == "info" {
				h.DisplayInfo(s, m)
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Invalid command")
			return
		}
	}
}

func (h *ChessHandler) NewGame(color string, s *discordgo.Session, m *discordgo.MessageCreate){

	if color != "white" && color != "White" && color != "black" && color != "Black"{
		s.ChannelMessageSend(m.ChannelID, color + " is not a valid selection, please choose black or white.")
		return
	}

	if color == "white" || color == "White"{

		account, err := h.bank.bank.GetAccountForUser("chessaccount")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error loading Chess Bank Account " + err.Error())
			return
		}

		err = h.chess.NewGame(m.Author.ID, "white")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not start new game: " + err.Error())
			return
		}
		err = h.ChargeUser(15, m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error charging user wallet " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "New game has been started, costing you 15 credits,"+
			"current prize pool is "+ strconv.Itoa(account.Balance/8)+ " credits, Good Luck!")
		h.SendBoard(s, m)
		return
	}
	if color == "black" || color == "Black" {
		account, err := h.bank.bank.GetAccountForUser("chessaccount")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error loading Chess Bank Account " + err.Error())
			return
		}

		err = h.chess.NewGame(m.Author.ID, "black")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not start new game: " + err.Error())
			return
		}
		err = h.ChargeUser(15, m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error charging user wallet " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "New game has been started, costing you 15 credits,"+
			"current prize pool is "+ strconv.Itoa(account.Balance/8)+ " credits" +
			"\nWhite Moving first, this may take a moment...")
		h.BotMoveCallback(s, m)
		return
	}
}

func (h *ChessHandler) SendBoard(s *discordgo.Session, m *discordgo.MessageCreate){
	message, err := h.chess.GetBoard(m.Author.ID,"default","default")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not display board: " + err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, message)
	return

}


func (h *ChessHandler) BotMoveCallback(s *discordgo.Session, m *discordgo.MessageCreate){

	responsechannel := make(chan string)

	go h.chess.ProcessBotMove(m.Author.ID, responsechannel)

	response := <-responsechannel

	if strings.HasPrefix(response,"botmove") {

		payload := strings.Fields(response)
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Error processing bot response, contact a dev!")
			return
		}

		s.ChannelMessageSend(m.ChannelID,"Bot moved to " + payload[1])
		h.SendBoard(s, m)
		return
	}
	if response == "quit" {

		// Returns 2 if white wins, -2 if black wins, 1 if it's stalemate, 0 if the game is still going.
		gamestatus, err := h.chess.GameStatus(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Game Error: " + err.Error())
			return
		}
		if gamestatus == 2 {
			record, err := h.chess.GetRecordFromDB(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Game Error: " + err.Error())
				return
			}

			if record.CurrentGameColor == "white" {
				err = h.ProcessWin(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error Processing Win: " + err.Error())
					return
				}
				h.SendBoard(s, m)
				s.ChannelMessageSend(m.ChannelID, "White Wins! Congratulations!")
				return
			}
			err = h.chess.ProcessLoss(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error Processing Loss: " + err.Error())
				return
			}
			h.SendBoard(s, m)
			s.ChannelMessageSend(m.ChannelID, "White Wins! Better luck next time!")
			return
		}
		if gamestatus == -2 {
			record, err := h.chess.GetRecordFromDB(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Game Error: " + err.Error())
				return
			}

			if record.CurrentGameColor == "black" {
				err = h.ProcessWin(m.Author.ID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Error Processing Win: " + err.Error())
					return
				}
				h.SendBoard(s, m)
				s.ChannelMessageSend(m.ChannelID, "Black Wins! Congratulations!")
				return
			}
			err = h.chess.ProcessLoss(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error Processing Loss: " + err.Error())
				return
			}
			h.SendBoard(s, m)
			s.ChannelMessageSend(m.ChannelID, "Black Wins! Better luck next time!")
			return
		}
		if gamestatus == 1 {
			err := h.chess.ProcessStalemate(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Game Error: " + err.Error())
				return
			}
			err = h.chess.ProcessStalemate(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error Processing Stalemate: " + err.Error())
				return
			}
			h.SendBoard(s, m)
			s.ChannelMessageSend(m.ChannelID, "Stalemate!")
			return
		}

	}

	s.ChannelMessageSend(m.ChannelID, "Game Error: " + response)
	return

}


func (h *ChessHandler) DisplayHelp(s *discordgo.Session, m *discordgo.MessageCreate){

	prompt := new(discordgo.MessageEmbed)

	account, err := h.bank.bank.GetAccountForUser("chessaccount")
	if err != nil {
		account = AccountRecord{ID: "Error", Balance: 0}
	}

	chessbalance := account.Balance

	prompt.URL = ""
	prompt.Type = ""
	prompt.Title = "Chess Help Page"
	prompt.Description = "Commands for <chess> \n :rotating_light:|| Curent Prize Pool - "+
		strconv.Itoa(chessbalance/8) +" Credits! ||:rotating_light:"
	prompt.Timestamp = ""
	prompt.Color = ColorDark_Red()

	footer := new(discordgo.MessageEmbedFooter)
	footer.Text = "When you see a good move, look for a better one - (Emanuel Lasker)"
	//	footer.IconURL = h.conf.BankConfig.BankIconURL
	prompt.Footer = footer
	/*
		image := new(discordgo.MessageEmbedImage)
		image.URL = h.conf.BankConfig.BankIconURL
		image.Height = 5
		image.Width = 5
		prompt.Image = image

	*/
	/*
	thumbnail := new(discordgo.MessageEmbedThumbnail)
	thumbnail.URL = ""
	thumbnail.Height = 10
	thumbnail.Width = 10
	prompt.Thumbnail = thumbnail
	*/
	/*
		video := new(discordgo.MessageEmbedVideo)
		video.URL = ""
		video.Height = 10
		video.Width = 10
		prompt.Video = video

		provider := new(discordgo.MessageEmbedProvider)
		provider.URL = ""
		provider.Name = ""
		prompt.Provider = provider
	*/
	author := new(discordgo.MessageEmbedAuthor)
	author.Name = "DU-DiscordBot Chess"
	author.URL = h.conf.BankConfig.BankURL
	author.IconURL = "https://discordapp.com/api/users/"+s.State.User.ID+"/avatars/"+s.State.User.Avatar+".jpg"
	prompt.Author = author


	fields := []*discordgo.MessageEmbedField{}

	infofield := discordgo.MessageEmbedField{}
	infofield.Name = "help"
	infofield.Value = "Display this page"
	infofield.Inline = true
	fields = append(fields, &infofield)

	newgamefield := discordgo.MessageEmbedField{}
	newgamefield.Name = "newgame <color>"
	newgamefield.Value = "Start a new game, white or black"
	newgamefield.Inline = true
	fields = append(fields, &newgamefield)

	loadfield := discordgo.MessageEmbedField{}
	loadfield.Name = "load"
	loadfield.Value = "Load recent unfinished game"
	loadfield.Inline = true
	fields = append(fields, &loadfield)

	resignfield := discordgo.MessageEmbedField{}
	resignfield.Name = "resign"
	resignfield.Value = "Resign from a game"
	resignfield.Inline = true
	fields = append(fields, &resignfield)

	profilefield := discordgo.MessageEmbedField{}
	profilefield.Name = "profile"
	profilefield.Value = "Display a player profile"
	profilefield.Inline = true
	fields = append(fields, &profilefield)

	stylesfield := discordgo.MessageEmbedField{}
	stylesfield.Name = "styles"
	stylesfield.Value = "Manage/purchase Piece and Board styles"
	stylesfield.Inline = true
	fields = append(fields, &stylesfield)

	boardfield := discordgo.MessageEmbedField{}
	boardfield.Name = "board"
	boardfield.Value = "Display your current game board"
	boardfield.Inline = true
	fields = append(fields, &boardfield)

	plotsfield := discordgo.MessageEmbedField{}
	plotsfield.Name = "plots"
	plotsfield.Value = "enable or disable plots"
	plotsfield.Inline = true
	fields = append(fields, &plotsfield)

	fenfield := discordgo.MessageEmbedField{}
	fenfield.Name = "fen"
	fenfield.Value = "Get your current game in FEN format"
	fenfield.Inline = true
	fields = append(fields, &fenfield)

	helpfield := discordgo.MessageEmbedField{}
	helpfield.Name = "info"
	helpfield.Value = "Display instructions about how to play DU-DiscordBot Chess"
	helpfield.Inline = true
	fields = append(fields, &helpfield)

	prompt.Fields = fields

	s.ChannelMessageSendEmbed(m.ChannelID, prompt)

}


func (h *ChessHandler) DisplayInfo( s *discordgo.Session, m *discordgo.MessageCreate){

	message := string("|| DU-DiscordBot Chess Help || \n" + "```" + "\n")
	message = message +"To start a new game: <chess> newgame\n\n"
	message = message +"If you get an error about a chess game record found, you can load it with: <chess> load\n\n"
	message = message +"The syntax for moving is: <piece><from><to>, for example - move pe2-e4 \n\n"
	message = message +"Valid pieces are: p-pawn, r-rook, n-knight, b-bishop, q-queen, k-king\n\n"
	message = message +"If you find yourself stuck, you can always resign!"
	message = message + "```"

	s.ChannelMessageSend(m.ChannelID, message)

}


func (h *ChessHandler) DisplayStyleMenu(payload []string, s *discordgo.Session, m *discordgo.MessageCreate){

	if len(payload) < 1 {
		stylesmenu := "Styles || \nAvailable Options are : "
		stylesmenu = stylesmenu + "```"
		stylesmenu = stylesmenu + "buy - list - inventory - set"
		stylesmenu = stylesmenu + "\n```\n"

		s.ChannelMessageSend(m.ChannelID, stylesmenu)
		return
	}
	if len(payload) > 0 {
		command, payload := SplitPayload(payload)

		if command == "buy" {

			if len(payload) < 1 {
				s.ChannelMessageSend(m.ChannelID, "<buy> expects an argument <stylename>")
				return
			}

			valid := false
			for _, style := range chessstylelist {
				if payload[0] == style {
					valid = true
				}
			}

			if !valid {
				s.ChannelMessageSend(m.ChannelID, "Invalid item selected")
				return
			}

			record, err := h.chess.GetRecordFromDB(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error retrieving record from db: " + err.Error())
				return
			}
			for _, ownedstyle := range record.PieceStyles {
				if ownedstyle == payload[0] {
					s.ChannelMessageSend(m.ChannelID, "Style already owned!")
					return
				}
			}

			err = h.ChargeUser(h.chess.GetStyleForName(payload[0]).Price, m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not charge wallet: " + err.Error())
				return
			}

			err = h.chess.AddPieceStyleByName(m.Author.ID, payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error adding piece style: " + err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " Purchased " + payload[0]+
				" for " +strconv.Itoa(h.chess.GetStyleForName(payload[0]).Price)+" credits!")
			return
		}
		if command == "list" {

			listmessage := h.DisplayStylesByList(chessstylelist)

			s.ChannelMessageSend(m.ChannelID, listmessage)
			return

		}
		if command == "inventory" || command == "inv" {

			record, err := h.chess.GetRecordFromDB(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not retrieve inventory: " + err.Error())
				return
			}

			message := ":\nStyles In " + m.Author.Mention() +"'s Inventory "
			message = message + h.DisplayStylesByList(record.PieceStyles)

			s.ChannelMessageSend(m.ChannelID, message)
			return
		}
		if command == "set" {
			if len(payload) < 1 {
				s.ChannelMessageSend(m.ChannelID, "<set> expects an argument <black||white> <stylename>")
				return
			}

			command, payload := SplitPayload(payload)
			if command == "help"{
				s.ChannelMessageSend(m.ChannelID, "<set> expects an argument <black||white> <stylename>")
				return
			}
			if command == "black"{
				if len(payload) < 1 {
					s.ChannelMessageSend(m.ChannelID, "<set black> expects an argument")
					return
				}
				h.ChangePieceStyle("black", payload[0], s, m)
				return
			}
			if command == "white"{
				if len(payload) < 1 {
					s.ChannelMessageSend(m.ChannelID, "<set white> expects an argument")
					return
				}
				h.ChangePieceStyle("white", payload[0], s, m)
				return
			}
		}
	}
}

func (h *ChessHandler) ChangePieceStyle(piece string, style string, s *discordgo.Session, m *discordgo.MessageCreate){

	if piece != "white" && piece != "black" {
		s.ChannelMessageSend(m.ChannelID, style + " is not a valid piece")
		return
	}

	isvalid := false
	for _, stylename := range chessstylelist {
		if stylename == style {
			isvalid = true
		}
	}
	if !isvalid {
		s.ChannelMessageSend(m.ChannelID, style + " is not a valid style")
		return
	}


	record, err := h.chess.GetRecordFromDB(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error loading chess account record: "+err.Error())
		return
	}

	owned := false

	for _, ownedstyle := range record.PieceStyles {
		if style == ownedstyle {
			owned = true
		}
	}

	if !owned {
		s.ChannelMessageSend(m.ChannelID, "You do not own this style!")
		return
	}
	if piece == "black" {
		if record.WhitePieceStyle == style && style != "default" {
			s.ChannelMessageSend(m.ChannelID, "White and Black piece styles cannot match! (except for default style)")
			return
		}
		if record.BlackPieceStyle == style {
			s.ChannelMessageSend(m.ChannelID, "Style for Black is already set!")
			return
		}
		record.BlackPieceStyle = style
	}
	if piece == "white" {
		if record.BlackPieceStyle == style && style != "default" {
			s.ChannelMessageSend(m.ChannelID, "White and Black piece styles cannot match! (except for default style)")
			return
		}
		if record.WhitePieceStyle == style {
			s.ChannelMessageSend(m.ChannelID, "Style for White is already set!")
			return
		}
		record.WhitePieceStyle = style
	}

	err = h.chess.SaveRecordToDB(record)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error: could not save chess record: "+err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Style for "+ piece+ " set!")
	return
}

func (h *ChessHandler) DisplayStylesByList(stylelist []string) (message string){

	listmessage := ":"

	listmessage = listmessage + "\n\n"
	listmessage = listmessage + "|  Name  |  Pawn  |  Rook  |  Knight  |  Bishop  |  Queen  |  King  |  Price  |\n"
	listmessage = listmessage + "|-----------------------------------------------------------------------------|\n"
	for _, stylename := range stylelist {
		style := h.chess.GetStyleForName(stylename)

		listmessage = listmessage + "| " +stylename + " | "
		listmessage = listmessage + style.Pawn.Symbol + "  |  "
		listmessage = listmessage + style.Rook.Symbol + "    |  "
		listmessage = listmessage + style.Knight.Symbol + "    |   "
		listmessage = listmessage + style.Bishop.Symbol + "    |   "
		listmessage = listmessage + style.Queen.Symbol + "    |   "
		listmessage = listmessage + style.King.Symbol + " |   "
		listmessage = listmessage + strconv.Itoa(style.Price) + " | \n"
	}
	listmessage = listmessage + ""
	return listmessage
}


func (h *ChessHandler) DisplayProfile(userid string, s *discordgo.Session, m *discordgo.MessageCreate){

	record, err := h.chess.GetRecordFromDB(userid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not display profile: " + err.Error())
		return
	}

	userstate, err := s.User(userid)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not display profile: " + err.Error())
		return
	}


	prompt := new(discordgo.MessageEmbed)

	prompt.URL = ""
	prompt.Type = ""
	prompt.Title = "Chess Profile"
	prompt.Description = "Player Record for " + userstate.Mention()
	prompt.Timestamp = ""
	prompt.Color = ColorDark_Red()

	//footer := new(discordgo.MessageEmbedFooter)
	//footer.Text = strconv.Itoa(record.Games)
	//	footer.IconURL = h.conf.BankConfig.BankIconURL
	//prompt.Footer = footer
	/*
		image := new(discordgo.MessageEmbedImage)
		image.URL = h.conf.BankConfig.BankIconURL
		image.Height = 5
		image.Width = 5
		prompt.Image = image

	*/

	thumbnail := new(discordgo.MessageEmbedThumbnail)
	thumbnail.URL = "https://discordapp.com/api/users/"+userid+"/avatars/"+userstate.Avatar+".jpg"
	thumbnail.Height = 10
	thumbnail.Width = 10
	prompt.Thumbnail = thumbnail

	/*
		video := new(discordgo.MessageEmbedVideo)
		video.URL = ""
		video.Height = 10
		video.Width = 10
		prompt.Video = video

		provider := new(discordgo.MessageEmbedProvider)
		provider.URL = ""
		provider.Name = ""
		prompt.Provider = provider
	*/
	author := new(discordgo.MessageEmbedAuthor)
	author.Name = "DU-DiscordBot Chess"
	author.URL = h.conf.BankConfig.BankURL
	author.IconURL = "https://discordapp.com/api/users/"+s.State.User.ID+"/avatars/"+s.State.User.Avatar+".jpg"
	prompt.Author = author


	fields := []*discordgo.MessageEmbedField{}

	gamesfield := discordgo.MessageEmbedField{}
	gamesfield.Name = "Games Played"
	gamesfield.Value = strconv.Itoa(record.Games)
	gamesfield.Inline = true
	fields = append(fields, &gamesfield)

	winsfield := discordgo.MessageEmbedField{}
	winsfield.Name = "Wins"
	winsfield.Value = strconv.Itoa(record.Wins)
	winsfield.Inline = true
	fields = append(fields, &winsfield)

	lossesfield := discordgo.MessageEmbedField{}
	lossesfield.Name = "Losses"
	lossesfield.Value = strconv.Itoa(record.Losses)
	lossesfield.Inline = true
	fields = append(fields, &lossesfield)

	if h.chess.CheckGameInProgress(record.UserID){
		currentgamefield := discordgo.MessageEmbedField{}
		currentgamefield.Name = "Current Game FEN"
		currentgamefield.Value, err = h.chess.GetFen(record.UserID)
		if err != nil {
			currentgamefield.Value = "No game in progress"
		}
		currentgamefield.Inline = true
		fields = append(fields, &currentgamefield)
	}

	previousgamefield := discordgo.MessageEmbedField{}
	previousgamefield.Name = "Last Game FEN"
	previousgamefield.Value = record.LastGameFEN
	if previousgamefield.Value == "" {
		previousgamefield.Value = "No game history found"
	}
	previousgamefield.Inline = true
	fields = append(fields, &previousgamefield)

	currentgamecolorfield := discordgo.MessageEmbedField{}
	currentgamecolorfield.Name = "Current Game Color"
	currentgamecolorfield.Value = record.CurrentGameColor
	currentgamecolorfield.Inline = true
	fields = append(fields, &currentgamecolorfield)


	// Get our Styles

	blackstyle, whitestyle, boardstyle := h.chess.GetStyles(record.BlackPieceStyle, record.WhitePieceStyle, record.BoardStyle)

	boardstyledescription := boardstyle.WhiteChessSquare + boardstyle.BlackChessSquare + "\n"
	boardstyledescription = boardstyledescription + boardstyle.BlackChessSquare + boardstyle.WhiteChessSquare

	boardstylefield := discordgo.MessageEmbedField{}
	boardstylefield.Name = "Board Style"
	boardstylefield.Value = boardstyledescription
	boardstylefield.Inline = true
	fields = append(fields, &boardstylefield)

	whitestyledescription := whitestyle.Pawn.Symbol
	whitestyledescription = whitestyledescription + whitestyle.Rook.Symbol
	whitestyledescription = whitestyledescription + whitestyle.Knight.Symbol
	whitestyledescription = whitestyledescription + whitestyle.Bishop.Symbol
	whitestyledescription = whitestyledescription + whitestyle.Queen.Symbol
	whitestyledescription = whitestyledescription + whitestyle.King.Symbol


	whitepiecestyle := discordgo.MessageEmbedField{}
	whitepiecestyle.Name = "White Pieces"
	whitepiecestyle.Value = whitestyledescription
	whitepiecestyle.Inline = true
	fields = append(fields, &whitepiecestyle)


	blackstyledescription := blackstyle.Pawn.Symbol
	blackstyledescription = blackstyledescription + blackstyle.Rook.Symbol
	blackstyledescription = blackstyledescription + blackstyle.Knight.Symbol
	blackstyledescription = blackstyledescription + blackstyle.Bishop.Symbol
	blackstyledescription = blackstyledescription + blackstyle.Queen.Symbol
	blackstyledescription = blackstyledescription + blackstyle.King.Symbol

	blackpiecestyle := discordgo.MessageEmbedField{}
	blackpiecestyle.Name = "Black Pieces"
	blackpiecestyle.Value = blackstyledescription
	blackpiecestyle.Inline = true
	fields = append(fields, &blackpiecestyle)

	prompt.Fields = fields

	fmt.Println("Sending Message")
	s.ChannelMessageSendEmbed(m.ChannelID, prompt)

}


func (h *ChessHandler) ProcessWin(userid string) (err error){
	err = h.chess.ProcessWin(userid)
	if err != nil {
		return err
	}

	account, err := h.bank.bank.GetAccountForUser("chessaccount")
	if err != nil {
		return err
	}

	err = h.PayUser(account.Balance/8, userid)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessHandler) PayUser(amount int, userid string) (err error){

	wallet, err := h.wallet.GetWallet(userid)
	if err != nil {
		return err
	}

	err = h.bank.Withdraw(amount, "chessaccount", wallet)
	return err
}


func (h *ChessHandler) ChargeUser(amount int, userid string) (err error){

	wallet, err := h.wallet.GetWallet(userid)
	if err != nil {
		return err
	}

	if wallet.Balance < amount {
		return errors.New("Insufficient Wallet Funds")
	}

	err = h.bank.Deposit(amount, "chessaccount", wallet)
	return err
}