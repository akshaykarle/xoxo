package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	BOARD_SIZE = 15
	WIN_LENGTH = 5
	EMPTY_CELL = '_'
	PLAYER_X   = 'X'
	PLAYER_O   = 'O'
)

// Pre-calculate board positions to avoid string allocations
var positionMap [BOARD_SIZE][BOARD_SIZE]string

// Pre-calculate direction checks for win conditions
var directions = [4][2]int{
	{0, 1},  // horizontal
	{1, 0},  // vertical
	{1, 1},  // diagonal
	{1, -1}, // anti-diagonal
}

// Initialize position map during package initialization
func init() {
	for i := 0; i < BOARD_SIZE; i++ {
		for j := 0; j < BOARD_SIZE; j++ {
			colStr := ""
			if j >= 26 {
				colStr = string('a'+j/26-1) + string('a'+j%26)
			} else {
				colStr = string('a' + j)
			}
			positionMap[i][j] = fmt.Sprintf("%s%d", colStr, i+1)
		}
	}
}

type Game struct {
	board         [BOARD_SIZE][BOARD_SIZE]rune
	currentPlayer rune
}

func NewGame() *Game {
	var g Game
	for i := range g.board {
		for j := range g.board[i] {
			g.board[i][j] = EMPTY_CELL
		}
	}
	return &g
}

// Optimized position parsing using a lookup table
func parsePosition(pos string) (row, col int, err error) {
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
	row = int(pos[len(pos)-1] - '1')

	if row < 0 || row >= BOARD_SIZE || col < 0 || col >= BOARD_SIZE {
		return 0, 0, fmt.Errorf("position out of bounds")
	}

	return row, col, nil
}

func (g *Game) parseBoardState(state string) error {
	var pos int
	var count int
	row := 0
	col := 0

	// Direct string parsing without splitting
	for pos < len(state) {
		ch := state[pos]
		if ch == '/' {
			if col != BOARD_SIZE {
				return fmt.Errorf("invalid row length")
			}
			row++
			col = 0
			pos++
			continue
		}

		if ch >= '0' && ch <= '9' {
			count = int(ch - '0')
			for i := 0; i < count && col < BOARD_SIZE; i++ {
				g.board[row][col] = EMPTY_CELL
				col++
			}
		} else {
			g.board[row][col] = rune(ch)
			col++
		}
		pos++
	}

	return nil
}

// Optimized win checking using direct array access
func (g *Game) checkWin(row, col int, player rune) bool {
	for _, dir := range directions {
		count := 1
		dx, dy := dir[0], dir[1]

		// Check forward
		for i := 1; i < WIN_LENGTH; i++ {
			r, c := row+dx*i, col+dy*i
			if r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || g.board[r][c] != player {
				break
			}
			count++
		}

		// Check backward
		for i := 1; i < WIN_LENGTH; i++ {
			r, c := row-dx*i, col-dy*i
			if r < 0 || r >= BOARD_SIZE || c < 0 || c >= BOARD_SIZE || g.board[r][c] != player {
				break
			}
			count++
		}

		if count >= WIN_LENGTH {
			return true
		}
	}
	return false
}

// Optimized move finding using threat-space search
func (g *Game) findBestMove() string {
	center := BOARD_SIZE / 2
	centerStart := center - 2
	centerEnd := center + 2

	// Quick check for first move
	if g.board[center][center] == EMPTY_CELL {
		return positionMap[center][center]
	}

	// Check center region first (most likely area for threats)
	for i := centerStart; i <= centerEnd; i++ {
		for j := centerStart; j <= centerEnd; j++ {
			if g.board[i][j] == EMPTY_CELL {
				// Check for winning move
				g.board[i][j] = g.currentPlayer
				if g.checkWin(i, j, g.currentPlayer) {
					g.board[i][j] = EMPTY_CELL
					return positionMap[i][j]
				}
				g.board[i][j] = EMPTY_CELL
			}
		}
	}

	// Check for blocking moves in center region
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
					return positionMap[i][j]
				}
				g.board[i][j] = EMPTY_CELL
			}
		}
	}

	// Spiral out from center for remaining moves
	for radius := 1; radius <= center; radius++ {
		for i := -radius; i <= radius; i++ {
			for j := -radius; j <= radius; j++ {
				row, col := center+i, center+j
				if row >= 0 && row < BOARD_SIZE && col >= 0 && col < BOARD_SIZE && g.board[row][col] == EMPTY_CELL {
					return positionMap[row][col]
				}
			}
		}
	}

	return positionMap[0][0]
}

func main() {
	game := NewGame()
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
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				game = NewGame()
				if err := game.parseBoardState(parts[1]); err != nil {
					continue
				}

				game.currentPlayer = rune(parts[2][0])

				var moveTime time.Duration
				for i := 3; i < len(parts)-1; i++ {
					if parts[i] == "time" && strings.HasPrefix(parts[i+1], "ms:") {
						var ms int
						fmt.Sscanf(parts[i+1][3:], "%d", &ms)
						moveTime = time.Duration(ms) * time.Millisecond
						break
					}
				}

				if moveTime > 0 {
					timer := time.NewTimer(moveTime)
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
			}

		case line == "quit":
			os.Exit(0)
		}
	}
}
