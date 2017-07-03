package main

import (
	"math/rand"
	"time"

	"errors"
	"strings"
	"strconv"
	"fmt"

	"github.com/yamamushi/chess/engine"
	"github.com/yamamushi/chess/search"
)


type ChessGame struct {

	boardlist []*ChessGameSession
	db *DBHandler

}


type ChessGameSession struct {
	board	*engine.Board
	UserID	string
	SecondPlayerID	string
	AIGame	bool
	AIMove	bool
	InProgress	bool
}


type ChessPlayerRecord struct {
	UserID           string `storm:"id"`
	Games            int
	Wins             int
	Losses           int
	LastGameFEN      string
	CurrentGame      string
	CurrentGameColor string
	LastColorMoved   string
	WhitePieceStyle  string
	BlackPieceStyle  string
	BoardStyle       string
	Plots            string  // Workaround because we can't store a bool in the db apparently
}


type ChessPiece struct {
	Type	string
	Symbol	string
}


type PieceStyle struct {

	King	ChessPiece
	Queen	ChessPiece
	Bishop	ChessPiece
	Knight	ChessPiece
	Rook	ChessPiece
	Pawn	ChessPiece
}


type BoardStyle struct {

	WhiteChessSquare string
	BlackChessSquare string

}



var (

	incmoves = make(chan *engine.Move, 1)
	outmoves = make(chan *engine.Move, 1)
	quit     = make(chan int, 1)

)

const LOG = false



/*

Functions Related to piece and board styles



 */


func (h *ChessGame) DefaultBoardStyle() (style BoardStyle){
	style.WhiteChessSquare = ":white_large_square:"
	style.BlackChessSquare = ":black_large_square:"
	return style
}


func (h *ChessGame) DefaultWhitePieces() (style PieceStyle) {

	style.King = ChessPiece{Type: "king", Symbol: ":prince::skin-tone-1:"}
	style.Queen = ChessPiece{Type: "queen", Symbol: ":princess::skin-tone-1:"}
	style.Bishop = ChessPiece{Type: "bishop", Symbol: ":construction_worker::skin-tone-1:"}
	style.Knight = ChessPiece{Type: "knight", Symbol: ":cop::skin-tone-1:"}
	style.Rook = ChessPiece{Type: "rook", Symbol: ":guardsman::skin-tone-1:"}
	style.Pawn = ChessPiece{Type: "pawn", Symbol: ":baby::skin-tone-1:"}
	return style

}

func (h *ChessGame) DefaultBlackPieces() (style PieceStyle) {

	style.King = ChessPiece{Type: "king", Symbol: ":prince::skin-tone-5:"}
	style.Queen = ChessPiece{Type: "queen", Symbol: ":princess::skin-tone-5:"}
	style.Bishop = ChessPiece{Type: "bishop", Symbol: ":construction_worker::skin-tone-5:"}
	style.Knight = ChessPiece{Type: "knight", Symbol: ":cop::skin-tone-5:"}
	style.Rook = ChessPiece{Type: "rook", Symbol: ":guardsman::skin-tone-5:"}
	style.Pawn = ChessPiece{Type: "pawn", Symbol: ":baby::skin-tone-5:"}
	return style
}

func (h *ChessGame) StyleChurchPieces() (style PieceStyle) {

	style.King = ChessPiece{Type: "king", Symbol: ":man::skin-tone-1:"}
	style.Queen = ChessPiece{Type: "queen", Symbol: ":woman::skin-tone-1:"}
	style.Bishop = ChessPiece{Type: "bishop", Symbol: ":older_man::skin-tone-1:"}
	style.Knight = ChessPiece{Type: "knight", Symbol: ":cop::skin-tone-1:"}
	style.Rook = ChessPiece{Type: "rook", Symbol: ":church:"}
	style.Pawn = ChessPiece{Type: "pawn", Symbol: ":bow::skin-tone-1:"}
	return style

}

func (h *ChessGame) StyleMosquePieces() (style PieceStyle) {

	style.King = ChessPiece{Type: "king", Symbol: ":man::skin-tone-4:"}
	style.Queen = ChessPiece{Type: "queen", Symbol: ":woman::skin-tone-4:"}
	style.Bishop = ChessPiece{Type: "bishop", Symbol: ":older_man::skin-tone-4:"}
	style.Knight = ChessPiece{Type: "knight", Symbol: ":man_with_turban::skin-tone-4:"}
	style.Rook = ChessPiece{Type: "rook", Symbol: ":mosque:"}
	style.Pawn = ChessPiece{Type: "pawn", Symbol: ":bow::skin-tone-4:"}
	return style
}

func (h *ChessGame) StyleAnimalPieces() (style PieceStyle) {

	style.King = ChessPiece{Type: "king", Symbol: ":lion_face:"}
	style.Queen = ChessPiece{Type: "queen", Symbol: ":tiger:"}
	style.Bishop = ChessPiece{Type: "bishop", Symbol: ":snake:"}
	style.Knight = ChessPiece{Type: "knight", Symbol: ":dolphin:"}
	style.Rook = ChessPiece{Type: "rook", Symbol: ":elephant:"}
	style.Pawn = ChessPiece{Type: "pawn", Symbol: ":fox:"}
	return style
}

func (h *ChessGame) GetStyleForName(name string) (style PieceStyle) {
	if name == "default" {
		style = h.DefaultBlackPieces()
	} else if name == "church"{
		style = h.StyleChurchPieces()
	} else if name == "mosque"{
		style = h.StyleMosquePieces()
	} else if name == "animal"{
		style = h.StyleAnimalPieces()
	} else {
		style = h.DefaultBlackPieces()
	}
	return style
}

func (h *ChessGame) GetStyles(black string, white string, board string) (blackstyle PieceStyle, whitestyle PieceStyle, boardstyle BoardStyle){


	if black == "default" {
		blackstyle = h.DefaultBlackPieces()
	} else if black == "church"{
		blackstyle = h.StyleChurchPieces()
	} else if black == "mosque"{
		blackstyle = h.StyleMosquePieces()
	} else if black == "animal"{
		blackstyle = h.StyleAnimalPieces()
	} else {
		blackstyle = h.DefaultBlackPieces()
	}


	if white == "default" {
		whitestyle = h.DefaultWhitePieces()
	} else if white == "church"{
		whitestyle = h.StyleChurchPieces()
	} else if white == "mosque"{
		whitestyle = h.StyleMosquePieces()
	} else if white == "animal"{
		whitestyle = h.StyleAnimalPieces()
	} else {
		whitestyle = h.DefaultBlackPieces()
	}

	if board == "default" {
		boardstyle = h.DefaultBoardStyle()
	} else {
		boardstyle = h.DefaultBoardStyle()
	}

	return blackstyle, whitestyle, boardstyle
}



func (h *ChessGame) EnablePlots(userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	record.Plots = "true"

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessGame) DisablePlots(userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	record.Plots = "false"

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	return nil
}



/*

Functions related to managing Chess Player Records


 */



func (h *ChessGame) SaveRecordToDB(record ChessPlayerRecord) (err error) {

	db := h.db.rawdb.From("Games").From("Chess")

	tmprecord := ChessPlayerRecord{}
	err = db.One("UserID", record.UserID, &tmprecord)
	if err == nil {
		fmt.Println("Updating Record")
		err = db.Update(&record)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Println("Creating New Record")
	err = db.Save(&record)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessGame) NewPlayerRecord(userid string) (err error){

	record := ChessPlayerRecord{UserID: userid, Games: 0, Wins: 0, Losses: 0, LastGameFEN: "", CurrentGame: "", WhitePieceStyle: "default",BlackPieceStyle: "default",BoardStyle: "default",Plots: "false",CurrentGameColor: "white"}

	err = h.SaveRecordToDB(record)
	return err

}


func (h *ChessGame) GetRecordFromDB(userid string) (record ChessPlayerRecord, err error) {
	db := h.db.rawdb.From("Games").From("Chess")

	userrecord := ChessPlayerRecord{}
	err = db.One("UserID", userid, &userrecord)
	if err != nil {
		return userrecord, err
	}
	return userrecord, nil
}


func (h *ChessGame) PlayerHasRecord(userid string) (bool){
	db := h.db.rawdb.From("Games").From("Chess")

	record := ChessPlayerRecord{}
	err := db.One("UserID", userid, &record)
	if err != nil {
		return false
	}
	return true
}


func (h *ChessGame) ProcessWin(userid string) (err error){

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.Wins = record.Wins + 1
	record.Games = record.Games + 1

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}


func (h *ChessGame) ProcessLoss(userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.Losses = record.Losses + 1
	record.Games = record.Games +1

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}


func (h *ChessGame) ProcessStalemate(userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.Games = record.Games +1

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

func (h *ChessGame) UpdateBlackStyle(style string, userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	if record.WhitePieceStyle == style && style != "default" {
		return errors.New("Cannot set White and Black pieces to the same styles")
	}
	record.BlackPieceStyle = style

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

func (h *ChessGame) UpdateWhiteStyle(style string, userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	if record.BlackPieceStyle == style && style != "default" {
		return errors.New("Cannot set White and Black pieces to the same styles")
	}

	record.WhitePieceStyle = style

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

func (h *ChessGame) UpdateBoardStyle(style string, userid string) (err error){
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.BoardStyle = style

	err= h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}


/*

Functions related to storing, managing, and interacting with live dubot-games

 */


func (h *ChessGame) Init() {

	h.boardlist = make([]*ChessGameSession, 0)

}


func (h *ChessGame) GetFen(userid string) (board string, err error) {

	game, err := h.GetGame(userid)
	if err != nil {
		return board, err
	}
	board = game.board.ToFen()
	return board, nil
}


func (h *ChessGame) GetGame(userid string) (game *ChessGameSession, err error){

	if !h.CheckGameInProgress(userid) {
		return game, errors.New("No Games in Progress")
	}

	for _, game := range h.boardlist {
		if game.UserID == userid {
			return game, nil
		}
	}
	return game, errors.New("No Game Found")
}


func (h *ChessGame) CheckGameInProgress(userid string) bool {
	for _, game := range h.boardlist {
		if game.UserID == userid {
			//fmt.Println("Game Found: " + game.UserID)
			return game.InProgress
		}
	}
	return false
}


func (h *ChessGame) NewGame(userid string, color string) (err error) {

	if h.CheckGameInProgress(userid){
		return errors.New("Game in progress, you must forfeit first!")
	}

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return errors.New("Could not find player record!")
	}

	if record.CurrentGame != "" && record.CurrentGame != " " {
		return errors.New("Record has game in progress")
	}

	if color == "white" {
		record.CurrentGameColor = color
	} else if color == "black" {
		record.CurrentGameColor = color
	} else {
		return errors.New("Invalid color selected")
	}

	gameinprogress := ChessGameSession{UserID: userid, board: &engine.Board{Turn: 1}, AIGame: true, AIMove: false, InProgress: true}
	gameinprogress.board.SetUpPieces()


	for i, game := range h.boardlist {
		if game.UserID == userid {
			//fmt.Println("Removing")
			h.boardlist = append(h.boardlist[:i], h.boardlist[i+1:]...)
		}
	}

	h.boardlist = append(h.boardlist, &gameinprogress)

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessGame) EndGame(userid string) (err error) {
	return h.Resign(userid)
}


// This doesn't work as expected, need to refactor
func (h *ChessGame) LoadGame(game string, board *engine.Board) (err error){

	log := strings.Fields(game)
	if len(log) < 1 {
		return errors.New("No game to load!")
	}

	// Reset our board
	board = &engine.Board{Turn: 1}
	board.SetUpPieces()

	//fmt.Println(game)

	// Unpack our log by replaying each move on the board
	for i, move := range log {
		movetype := h.StringToMove(move)
		err := board.Move(movetype)
		//fmt.Println("Moving " + movetype.ToString())
		if err != nil{
			return errors.New("Error loading game at move " + strconv.Itoa(i) + " : "+ err.Error())
		}
	}
	return nil
}


func (h *ChessGame) LoadCurrentGame(userid string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	if record.CurrentGame == "" {
		return errors.New("No Game Found!")
	}

	log := strings.Fields(record.CurrentGame)
	if len(log) < 1 {
		return errors.New("No game to load!")
	}

	//fmt.Println("Getting Game")

	for i, game := range h.boardlist {
		if game.UserID == userid {
			//fmt.Println("Removing")
			h.boardlist = append(h.boardlist[:i], h.boardlist[i+1:]...)
		}
	}

	session := &ChessGameSession{UserID: userid, AIGame: true, AIMove: false, InProgress: true}
	h.boardlist = append(h.boardlist, session)
	//fmt.Println("Adding")


	// Reset our board
	session.board = &engine.Board{Turn: 1}
	session.board.SetUpPieces()

	//fmt.Println(record.CurrentGame)

	// Unpack our log by replaying each move on the board
	for i, move := range log {
		movetype := h.StringToMove(move)
		err := session.board.Move(movetype)
		//fmt.Println("Moving " + movetype.ToString())
		if err != nil{
			return errors.New("Error loading game at move " + strconv.Itoa(i) + " : "+ err.Error())
		}
	}

	h.boardlist = append(h.boardlist, session)
	if err != nil {
		return err
	}

	if record.LastColorMoved == record.CurrentGameColor {

		responsechannel := make(chan string)

		go h.ProcessBotMove(userid, responsechannel)

		response := <-responsechannel

		if strings.HasPrefix(response,"botmove") {
			payload := strings.Fields(response)
			if len(payload) < 2 {
				return errors.New("Error Loading Game: Processing bot response payload is nil, contact a dev!")
			}
			return
		}
		if response == "quit" {
			// Returns 2 if white wins, -2 if black wins, 1 if it's stalemate, 0 if the game is still going.
			gamestatus, err := h.GameStatus(userid)
			if err != nil {
				return errors.New("Error Loading Game: " + err.Error())
			}
			if gamestatus == 2 {
				return errors.New("White Wins!")
			}
			if gamestatus == -2 {
				return errors.New("Black Wins!")
			}
			if gamestatus == 1 {
				err := h.ProcessStalemate(userid)
				if err != nil {
					return errors.New("Error Loading Game: " + err.Error())
				}
				return errors.New("Stalemate!")
			}
		}
	}

	return nil
}


func (h *ChessGame) Resign(userid string) (err error){

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	game, err := h.GetGame(userid)
	if err != nil {
		return err
	}

	game.InProgress = false
	record.LastGameFEN = game.board.ToFen()
	record.CurrentGame = " "
	record.Losses = record.Losses + 1
	record.Games = record.Games + 1

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	record, err = h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	//fmt.Println(record.CurrentGame)


	/*
	This used to be wrapped in ChessGame.EndGame() but to avoid
	a race condition happening, it now happens here in one go.
	 */

	for i, game := range h.boardlist {
		if game.UserID == userid {
			h.boardlist = append(h.boardlist[:i], h.boardlist[i+1:]...)
			return nil
		}
	}

	return nil
}



/*

Functions related to interacting with live chess boards

 */

// This will not updating the player record, call PlayerMove instead!
func (h *ChessGame) Move(message string, board *engine.Board) (err error){

	if len(message) < 6 {
		return errors.New("Invalid Move")
	}

	move := h.StringToMove(message)
	err = board.Move(move)
	if err != nil{
		return err
	}

	return nil
}


// Returns 2 if white wins, -2 if black wins, 1 if it's stalemate, 0 if the game is still going.
func (h *ChessGame) GameStatus(userid string) (status int, err error) {

	board, err := h.GetGame(userid)
	if err != nil {
		return 0, err
	}
	return board.board.IsOver(), nil
}


func (h *ChessGame) PlayerMove(userid string, move string) (err error) {


	record, err:= h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	game, err := h.GetGame(userid)
	if err != nil {
		return err
	}

	if game.AIMove{
		return errors.New("It is not your turn!")
	}

	err = h.Move(move, game.board)
	if err != nil {
		return err
	}

	record.CurrentGame = record.CurrentGame + move + " "

	if record.CurrentGameColor == "white" {
		record.LastColorMoved = "white"
	} else {
		record.LastColorMoved = "black"
	}

	if game.AIGame {
		game.AIMove = true
	}

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessGame) BotMove(userid string, move string) (err error) {
	db := h.db.rawdb.From("Games").From("Chess")

	record := ChessPlayerRecord{}
	err = db.One("UserID", userid, &record)
	if err != nil {
		return err
	}

	game, err := h.GetGame(userid)
	if err != nil {
		return err
	}
	if !game.AIGame{
		return errors.New("Not an AI Game!")
	}

	h.Move(move, game.board)
	record.CurrentGame = record.CurrentGame + move + " "

	if record.CurrentGameColor == "white" {
		record.LastColorMoved = "black"
	} else {
		record.LastColorMoved = "white"
	}

	if game.AIGame{
		game.AIMove = false
	}
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	return nil
}


func (h *ChessGame) ProcessBotMove(userid string, response chan string){

	db := h.db.rawdb.From("Games").From("Chess")

	record := ChessPlayerRecord{}
	err := db.One("UserID", userid, &record)
	if err != nil {
		response <- err.Error()
		return
	}

	game, err := h.GetGame(userid)
	if err != nil {
		response <- err.Error()
		return
	}
	if !game.AIGame{
		response <- "Not an AI Game!"
		return
	}

	board := game.board

	rand.Seed(time.Now().UTC().UnixNano())

	var mymove *engine.Move
	if moves, ok := search.Book[board.ToFen()]; ok {
		mymove = h.StringToMove(moves[rand.Intn(len(moves))])
	} else {
		if m := search.AlphaBeta(board, 4, search.BLACKWIN, search.WHITEWIN); m != nil {
			mymove = m
		} else {
			response <- "quit"
			return
		}
	}

	movestring := mymove.ToString()
	err = h.BotMove( userid, movestring)
	if err != nil {
		response <- err.Error()
		return
	}

	response <- "botmoved "+movestring
	return
}


func (h *ChessGame) GetBoard(userid string, piecestyle string, boardstyle string) (board string, err error){

	board = ":\n"
	board = board + ":earth_americas: :earth_africa: :earth_asia:|| Chess ||:earth_asia: :earth_africa: :earth_americas:"
	board = board + "\n----------------------------------------\n"

	blackstyle := PieceStyle{}
	whitestyle := PieceStyle{}
	bstyle := BoardStyle{}

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return board, err
	}

	blackstyle, whitestyle, bstyle = h.GetStyles(record.BlackPieceStyle, record.WhitePieceStyle, record.BoardStyle)

	game, err := h.GetGame(userid)
	if err != nil {
		return "", errors.New("Could not find game in progress!")
	}


	boardarray := game.board.ToArray()
	for y := 7; y >= 0; y-- {
		for x := 0; x < 8; x++ {
			if boardarray[y][x] == "" {
				if y % 2 == 1 && x % 2 == 1{
					board = board + bstyle.BlackChessSquare + " "
				}
				if y % 2 == 1 && x % 2 == 0{
					board = board + bstyle.WhiteChessSquare + " "
				}
				if y % 2 == 0 && x % 2 == 0{
					board = board + bstyle.BlackChessSquare + " "
				}
				if y % 2 == 0 && x % 2 == 1{
					board = board + bstyle.WhiteChessSquare + " "
				}

			} else {
				if boardarray[y][x] == "p"{
					board = board + blackstyle.Pawn.Symbol + " "
				}
				if boardarray[y][x] == "P"{
					board = board + whitestyle.Pawn.Symbol + " "
				}
				if boardarray[y][x] == "r"{
					board = board + blackstyle.Rook.Symbol + " "
				}
				if boardarray[y][x] == "R"{
					board = board + whitestyle.Rook.Symbol + " "
				}
				if boardarray[y][x] == "n"{
					board = board + blackstyle.Knight.Symbol + " "
				}
				if boardarray[y][x] == "N"{
					board = board + whitestyle.Knight.Symbol + " "
				}
				if boardarray[y][x] == "b"{
					board = board + blackstyle.Bishop.Symbol + " "
				}
				if boardarray[y][x] == "B"{
					board = board + whitestyle.Bishop.Symbol + " "
				}
				if boardarray[y][x] == "q"{
					board = board + blackstyle.Queen.Symbol + " "
				}
				if boardarray[y][x] == "Q"{
					board = board + whitestyle.Queen.Symbol + " "
				}
				if boardarray[y][x] == "k"{
					board = board + blackstyle.King.Symbol + " "
				}
				if boardarray[y][x] == "K"{
					board = board + whitestyle.King.Symbol + " "
				}
			}
		}

		if record.Plots == "true" {

			if y == 0{
				board = board + ":one:"
			}
			if y == 1{
				board = board + ":two:"
			}
			if y == 2{
				board = board + ":three:"
			}
			if y == 3{
				board = board + ":four:"
			}
			if y == 4{
				board = board + ":five:"
			}
			if y == 5{
				board = board + ":six:"
			}
			if y == 6{
				board = board + ":seven:"
			}
			if y == 7{
				board = board + ":eight:"
			}
		}
		board = board + "\n"
	}
	if record.Plots == "true" {
		board = board + ":regional_indicator_a: :regional_indicator_b: :regional_indicator_c: :regional_indicator_d:" +
			" :regional_indicator_e: :regional_indicator_f: :regional_indicator_g: :regional_indicator_h:"
	}
	return board, nil
}


// Accepts a string such as "pe2-e4" and converts it to the Move struct.
func (h *ChessGame) StringToMove(s string) *engine.Move {
	var move engine.Move
	move.Begin = h.StringToSquare(s[1:3])
	move.End = h.StringToSquare(s[4:])
	move.Piece = s[0]
	return &move
}


// Accepts a string such as "e4'"and converts it to the Square struct.
func (h *ChessGame) StringToSquare(s string) engine.Square {
	var square engine.Square
	for i, b := range engine.Files {
		if b == s[0] {
			square.X = i + 1
			break
		}
	}
	for i, b := range engine.Ranks {
		if b == s[1] {
			square.Y = i + 1
			break
		}
	}
	return square
}
