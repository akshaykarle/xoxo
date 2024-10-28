package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_BOARD_SIZE     = 100 // Maximum supported board size
	DEFAULT_WIN_LENGTH = 3
	EMPTY_CELL         = '_'
	PLAYER_X           = 'X'
	PLAYER_O           = 'O'
)

// Pre-calculate direction checks for win conditions
var directions = [4][2]int{
	{0, 1},  // horizontal
	{1, 0},  // vertical
	{1, 1},  // diagonal
	{1, -1}, // anti-diagonal
}

// Position cache for board sizes
type PositionCache struct {
	cache map[int][][]string
}

func NewPositionCache() *PositionCache {
	return &PositionCache{
		cache: make(map[int][][]string),
	}
}

func (pc *PositionCache) GetPosition(size, row, col int) string {
	if positions, exists := pc.cache[size]; exists {
		return positions[row][col]
	}

	// Create position map for this board size
	positions := make([][]string, size)
	for i := range positions {
		positions[i] = make([]string, size)
		for j := range positions[i] {
			colStr := ""
			if j >= 26 {
				colStr = string('a'+j/26-1) + string('a'+j%26)
			} else {
				colStr = string('a' + j)
			}
			positions[i][j] = fmt.Sprintf("%s%d", colStr, i+1)
		}
	}

	pc.cache[size] = positions
	return positions[row][col]
}

type Game struct {
	board         [][]rune
	currentPlayer rune
	boardSize     int
	winLength     int
	posCache      *PositionCache
}

func NewGame(boardSize, winLength int) *Game {
	board := make([][]rune, boardSize)
	for i := range board {
		board[i] = make([]rune, boardSize)
		for j := range board[i] {
			board[i][j] = EMPTY_CELL
		}
	}
	return &Game{
		board:     board,
		boardSize: boardSize,
		winLength: winLength,
		posCache:  NewPositionCache(),
	}
}

// Parse a position like "a1" or "aa15"
func parsePosition(pos string, boardSize int) (row, col int, err error) {
	if len(pos) < 2 {
		return 0, 0, fmt.Errorf("invalid position")
	}

	// Parse column
	colPart := strings.ToLower(pos[:len(pos)-1])
	if len(colPart) == 1 {
		col = int(colPart[0] - 'a')
	} else if len(colPart) == 2 {
		col = (int(colPart[0]-'a'+1) * 26) + int(colPart[1]-'a')
	}

	// Parse row (1-based to 0-based)
	rowStr := pos[len(colPart):]
	rowNum, err := strconv.Atoi(rowStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid row number")
	}
	row = rowNum - 1

	if row < 0 || row >= boardSize || col < 0 || col >= boardSize {
		return 0, 0, fmt.Errorf("position out of bounds")
	}

	return row, col, nil
}

// Print the board in a readable format
func (g *Game) printBoard() string {
	var sb strings.Builder

	// Print column headers
	sb.WriteString("  ") // Space for row numbers
	for col := 0; col < g.boardSize; col++ {
		if col >= 26 {
			sb.WriteString(string('a' + col/26 - 1))
			sb.WriteString(string('a' + col%26))
			sb.WriteString(" ")
		} else {
			sb.WriteString(string('a' + col))
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	// Print rows with numbers and cells
	for row := 0; row < g.boardSize; row++ {
		// Add row number
		sb.WriteString(fmt.Sprintf("%2d", row+1))

		// Add cells
		for col := 0; col < g.boardSize; col++ {
			sb.WriteString(" ")
			sb.WriteRune(g.board[row][col])
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *Game) parseBoardState(state string) error {
	rows := strings.Split(state, "/")

	for i, row := range rows {
		col := 0
		for pos := 0; pos < len(row); pos++ {
			ch := rune(row[pos])

			if ch >= '0' && ch <= '9' {
				// Handle multi-digit numbers
				numStr := string(ch)
				for pos+1 < len(row) && row[pos+1] >= '0' && row[pos+1] <= '9' {
					pos++
					numStr += string(row[pos])
				}
				count, err := strconv.Atoi(numStr)
				if err != nil {
					return fmt.Errorf("invalid number in board state")
				}
				col += count
				break
			} else if ch == EMPTY_CELL {
				// Handle single empty cell
				col++
			} else if ch == PLAYER_X || ch == PLAYER_O {
				// Handle player marks
				if col >= g.boardSize {
					return fmt.Errorf("invalid row length")
				}
				g.board[i][col] = ch
				col++
			} else {
				return fmt.Errorf("invalid character in board state")
			}
		}

		if col != g.boardSize {
			return fmt.Errorf("invalid row length, got %d expected %d", col, g.boardSize)
		}
	}

	return nil
}

func (g *Game) checkWin(row, col int, player rune) bool {
	for _, dir := range directions {
		count := 1
		dx, dy := dir[0], dir[1]

		// Check forward
		for i := 1; i < g.winLength; i++ {
			r, c := row+dx*i, col+dy*i
			if r < 0 || r >= g.boardSize || c < 0 || c >= g.boardSize || g.board[r][c] != player {
				break
			}
			count++
		}

		// Check backward
		for i := 1; i < g.winLength; i++ {
			r, c := row-dx*i, col-dy*i
			if r < 0 || r >= g.boardSize || c < 0 || c >= g.boardSize || g.board[r][c] != player {
				break
			}
			count++
		}

		if count >= g.winLength {
			return true
		}
	}
	return false
}

func (g *Game) findBestMove() string {
	center := g.boardSize / 2
	radius := g.winLength
	centerStart := max(0, center-radius)
	centerEnd := min(g.boardSize-1, center+radius)

	// Quick check for first move
	if g.board[center][center] == EMPTY_CELL {
		return g.posCache.GetPosition(g.boardSize, center, center)
	}

	// Check center region first (most likely area for threats)
	for i := centerStart; i <= centerEnd; i++ {
		for j := centerStart; j <= centerEnd; j++ {
			if g.board[i][j] == EMPTY_CELL {
				// Check for winning move
				g.board[i][j] = g.currentPlayer
				if g.checkWin(i, j, g.currentPlayer) {
					g.board[i][j] = EMPTY_CELL
					return g.posCache.GetPosition(g.boardSize, i, j)
				}
				g.board[i][j] = EMPTY_CELL
			}
		}
	}

	// Check for blocking moves
	opponent := PLAYER_O
	if g.currentPlayer == PLAYER_O {
		opponent = PLAYER_X
	}

	for i := centerStart; i <= centerEnd; i++ {
		for j := centerStart; j <= centerEnd; j++ {
			if g.board[i][j] == EMPTY_CELL {
				g.board[i][j] = opponent
				if g.checkWin(i, j, opponent) {
					g.board[i][j] = EMPTY_CELL
					return g.posCache.GetPosition(g.boardSize, i, j)
				}
				g.board[i][j] = EMPTY_CELL
			}
		}
	}

	// Spiral out from center
	for r := 1; r <= center; r++ {
		for i := -r; i <= r; i++ {
			for j := -r; j <= r; j++ {
				row, col := center+i, center+j
				if row >= 0 && row < g.boardSize && col >= 0 && col < g.boardSize && g.board[row][col] == EMPTY_CELL {
					return g.posCache.GetPosition(g.boardSize, row, col)
				}
			}
		}
	}

	return g.posCache.GetPosition(g.boardSize, 0, 0)
}

func parseCommandLine(line string) (boardState string, player rune, timeLimit time.Duration, winLength int, err error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return "", 0, 0, 0, fmt.Errorf("invalid command format")
	}

	boardState = parts[1]
	player = rune(parts[2][0])
	winLength = DEFAULT_WIN_LENGTH // Default win length

	// Parse additional parameters
	for i := 3; i < len(parts); i++ {
		switch parts[i] {
		case "time":
			if i+1 < len(parts) && strings.HasPrefix(parts[i+1], "ms:") {
				ms, err := strconv.Atoi(parts[i+1][3:])
				if err != nil {
					continue
				}
				timeLimit = time.Duration(ms) * time.Millisecond
				i++
			}
		case "win-length":
			if i+1 < len(parts) {
				wl, err := strconv.Atoi(parts[i+1])
				if err == nil && wl > 0 {
					winLength = wl
				}
				i++
			}
		}
	}

	return boardState, player, timeLimit, winLength, nil
}

func main() {
	var game *Game
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "st3p version"):
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Printf("st3p version %s ok\n", parts[2])
			}

		case line == "identify":
			fmt.Println("identify name optimized-tictactoe")
			fmt.Println("identify author chat-assistant")
			fmt.Println("identify version 2.0.0")
			fmt.Println("identify ok")

		case strings.HasPrefix(line, "move"):
			boardState, player, timeLimit, winLength, err := parseCommandLine(line)
			if err != nil {
				continue
			}

			// Create new game with board size determined from state
			rows := strings.Split(boardState, "/")
			boardSize := len(rows)
			game = NewGame(boardSize, winLength)

			if err := game.parseBoardState(boardState); err != nil {
				fmt.Println("Error when parsing board: ", err)
				continue
			}

			// println("Current board:")
			// println(game.printBoard())

			game.currentPlayer = player

			if timeLimit > 0 {
				timer := time.NewTimer(timeLimit)
				moveChan := make(chan string)

				go func() {
					moveChan <- game.findBestMove()
				}()

				select {
				case move := <-moveChan:
					fmt.Printf("best %s\n", move)
				case <-timer.C:
					fmt.Printf("best %s\n", game.findBestMove()) // Fallback to quick move
				}
			} else {
				fmt.Printf("best %s\n", game.findBestMove())
			}

		case line == "quit":
			os.Exit(0)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
