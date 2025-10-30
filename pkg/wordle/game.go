package wordle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Backshifted/wordle-solver/assets"
)

const (
	MaxGuesses           = 6
	WordLength           = 5
	qwertyKeyboardLayout = "q w e r t y u i o p\n a s d f g h j k l\n  z x c v b n m"

	ansiReset          = "\033[0m"
	ansiUnderlined     = "\033[4m"
	ansiNotUnderlined  = "\033[24m"
	ansiFgBlack        = "\033[30m"
	ansiFgWhite        = "\033[97m"
	ansiFgGreen        = "\033[32m"
	ansiFgYellow       = "\033[33m"
	ansiBgGreen        = "\033[42m"
	ansiBgBrightGreen  = "\033[102m"
	ansiBgBrightYellow = "\033[103m"
	ansiBgBrightWhite  = "\033[107m"
	ansiWhite          = ansiBgBrightWhite + ansiFgBlack
	ansiGreen          = ansiBgBrightGreen + ansiFgBlack
	ansiYellow         = ansiBgBrightYellow + ansiFgBlack
)

type (
	// Ordered enum
	Hint  byte
	Guess struct {
		word  string
		hints [5]Hint
	}
	GameState struct {
		guesses  [6]*Guess
		alphabet [26]Hint
	}
)

const (
	LetterPossible Hint = iota
	LetterWrong
	LetterTransposed
	LetterCorrect
)

func NewGuess(word string, hints [5]Hint) *Guess {
	return &Guess{
		word:  word,
		hints: hints,
	}
}

func NewGameState() GameState {
	return GameState{
		guesses:  [6]*Guess{},
		alphabet: [26]Hint{},
	}
}

func (gs *GameState) AddGuess(word string, hints [5]Hint) error {
	if gs.guesses[MaxGuesses-1] != nil {
		return fmt.Errorf("Exceeded maximum number of guesses: %d", MaxGuesses)
	}
	if len(word) != 5 {
		return fmt.Errorf("Invalid word length %d, should be %d", len(word), WordLength)
	}
	if len(hints) != 5 {
		return fmt.Errorf("Invalid hints length %d, should be %d", len(hints), WordLength)
	}
	if !isAscii(word) {
		return fmt.Errorf("Exceeded maximum number of guesses: %d", MaxGuesses)
	}
	if !slices.Contains(assets.Wordles, word) && !slices.Contains(assets.NonWordles, word) {
		return fmt.Errorf("Not a valid word")
	}

	for i, guess := range gs.guesses {
		if guess == nil {
			// Update alphabet
			for j, hint := range hints {
				letterIndex := word[j] - 'a'
				// Make use of ordered enum
				gs.alphabet[letterIndex] = max(hint, gs.alphabet[letterIndex])
			}

			gs.guesses[i] = NewGuess(word, [5]Hint(hints))
			break
		}
	}

	return nil
}

func isAscii(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return len(s) > 0
}

func (gs GameState) String() string {
	const prefix = "      "
	output := strings.Builder{}
	output.WriteString(prefix + "+-----+\n")

	var i int
	for i = 0; i < len(gs.guesses) && gs.guesses[i] != nil; i++ {
		output.WriteString(prefix + "|" + formatGuess(gs.guesses[i]) + "|\n")
	}
	for ; i < MaxGuesses; i++ {
		output.WriteString(prefix + "|     |\n")
	}

	output.WriteString(prefix + "+-----+\n")
	output.WriteString(formatAlphabet(gs.alphabet[:]))
	return output.String()
}

func formatGuess(guess *Guess) string {
	output := strings.Builder{}

	for i := range WordLength {
		formatLetter(&output, guess.word[i], guess.hints[i])
	}

	return output.String()
}

func formatAlphabet(state []Hint) string {
	output := strings.Builder{}

	for _, letter := range qwertyKeyboardLayout {
		if letter >= 'a' && letter <= 'z' {
			formatLetter(&output, byte(letter), state[letter-'a'])
		} else {
			output.WriteRune(letter)
		}
	}

	return output.String()
}

func formatLetter(b *strings.Builder, letter byte, hint Hint) {
	if LetterCorrect == hint {
		b.WriteString(ansiFgGreen)
	} else if LetterTransposed == hint {
		b.WriteString(ansiFgYellow)
	} else if LetterWrong == hint {
		b.WriteString(ansiFgBlack)
	}

	b.WriteByte(letter - 32) // uppercase
	b.WriteString(ansiReset)
}

type Game struct {
	word  string
	state GameState
}

func NewGame(word string) *Game {
	return &Game{
		word:  word,
		state: NewGameState(),
	}
}

func (g *Game) Guess(word string) (bool, error) {
	if len(word) != 5 {
		return false, fmt.Errorf("Invalid word length %d, should be %d", len(word), WordLength)
	}

	hints := [5]Hint{}
	for i := range WordLength {
		hints[i] = LetterWrong
	}
	bag := []byte(g.word)

	// Check correct letter
	for i := range WordLength {
		if word[i] == bag[i] {
			// Prevent duplicate yellows by removing letters from the word/bag
			bag[i] = '?'
			hints[i] = LetterCorrect
		}
	}
	// Check transposed letters
	for i := range WordLength {
		if index := slices.Index(bag, word[i]); index != -1 {
			// Prevent duplicate yellows by removing letters from the word/bag
			bag[index] = '?'
			hints[i] = LetterTransposed
		}
	}

	if err := g.state.AddGuess(word, hints); err != nil {
		return false, err
	}

	for _, hint := range hints {
		if hint != LetterCorrect {
			return false, nil
		}
	}

	return true, nil
}

func (g Game) String() string {
	return g.state.String()
}
