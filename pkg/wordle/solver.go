package wordle

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode"

	"overberne.com/worlde/assets"
)

type (
	Pattern [5]Hint
	// Each bit is a unique letter
	// bit 1 = a, bit 2 = b, etc.
	LetterConstraint   uint32
	MinCountConstraint struct {
		Char  byte
		Count byte
	}
	Constraint struct {
		Letters [5]LetterConstraint
		// A static length array saves spaces over a map, 10 vs 48 bytes.
		Counts [5]MinCountConstraint
	}
	WordleTreeNode struct {
		Options  LetterConstraint
		Children map[byte]*WordleTreeNode
	}
	WordleTree struct {
		WordleTreeNode
		WordCount int
		Wordles   []string
	}
	WordUtility struct {
		Word    string
		Utility float64
	}
	ConstraintMap map[string][]Constraint
)

var (
	Hints              = [3]Hint{LetterWrong, LetterCorrect, LetterTransposed}
	NumPatterns        = int(math.Pow(float64(len(Hints)), WordLength))
	AllPatterns        = generateAllPatterns()
	SolutionTree       = assets.Load[*WordleTree](assets.SolutionTreeFile)
	GuessTree          = assets.Load[*WordleTree](assets.GuessTreeFile)
	GuessConstraintMap = assets.Load[ConstraintMap](assets.ConstraintMapFile)
	// SolutionTree       = NewWordleTree(Wordles, NewConstraint())
	// GuessTree          = NewWordleTree(WordlesAndNonWordles, NewConstraint())
	// GuessConstraintMap = NewConstraintMap(WordlesAndNonWordles, SolutionTree)
)

func NewLetterConstraint(char byte) LetterConstraint {
	return 1 << LetterConstraint(char-'a')
}

func (lc LetterConstraint) Include(char byte) LetterConstraint {
	return lc | NewLetterConstraint(char)
}

func (lc LetterConstraint) Exclude(char byte) LetterConstraint {
	return lc & ^NewLetterConstraint(char)
}

func (lc LetterConstraint) Includes(letter LetterConstraint) bool {
	return letter&lc > 0
}

func (lc LetterConstraint) ToChars() []byte {
	options := make([]byte, 0)

	for char := byte('a'); char <= 'z'; char++ {
		bitmask := NewLetterConstraint(char)
		if lc&bitmask > 0 {
			options = append(options, char)
		}
	}

	return options
}

func NewConstraint() Constraint {
	return Constraint{
		Letters: [5]LetterConstraint{
			LetterConstraint(math.MaxUint32),
			LetterConstraint(math.MaxUint32),
			LetterConstraint(math.MaxUint32),
			LetterConstraint(math.MaxUint32),
			LetterConstraint(math.MaxUint32),
		},
		Counts: [5]MinCountConstraint{},
	}
}

func (c Constraint) And(other Constraint) Constraint {
	combinedCounts := [5]MinCountConstraint(slices.Clone(c.Counts[:]))
	for _, counts := range other.Counts {
		for i := range WordLength {
			if counts.Char == combinedCounts[i].Char {
				combinedCounts[i].Count = max(combinedCounts[i].Count, counts.Count)
				break
			}
			if combinedCounts[i].Char == 0 {
				combinedCounts[i] = counts
				break
			}
		}
	}

	return Constraint{
		Letters: [5]LetterConstraint{
			c.Letters[0] & other.Letters[0],
			c.Letters[1] & other.Letters[1],
			c.Letters[2] & other.Letters[2],
			c.Letters[3] & other.Letters[3],
			c.Letters[4] & other.Letters[4],
		},
		Counts: combinedCounts,
	}
}

func (c Constraint) Dec(char byte) Constraint {
	for i := range len(c.Counts) {
		if c.Counts[i].Char == char {
			c.Counts[i].Count--
		}
	}

	return c
}

func (c Constraint) Matches(word string) bool {
	for i := range WordLength {
		letter := NewLetterConstraint(word[i])
		if !c.Letters[i].Includes(letter) {
			return false
		}
	}

	for _, countConstraint := range c.Counts {
		count := byte(0)
		for i := range WordLength {
			if word[i] == countConstraint.Char {
				count++
			}
		}

		if count < countConstraint.Count {
			return false
		}
	}

	return true
}

func NewWordleTreeNode() *WordleTreeNode {
	return &WordleTreeNode{
		Options:  LetterConstraint(0),
		Children: make(map[byte]*WordleTreeNode),
	}
}

func NewWordleTree(wordles []string, constraint Constraint) *WordleTree {
	root := &WordleTree{WordleTreeNode: *NewWordleTreeNode()}
	var node *WordleTreeNode

	for _, word := range wordles {
		if !constraint.Matches(word) {
			continue
		}

		root.Wordles = append(root.Wordles, word)
		root.WordCount++
		node = &root.WordleTreeNode

		for i := range WordLength {
			char := word[i]
			if child, ok := node.Children[char]; ok {
				node = child
			} else {
				child := NewWordleTreeNode()
				node.Children[char] = child
				node.Options = node.Options.Include(char)
				node = child
			}
		}
	}

	return root
}

func (wt *WordleTree) HasMatches(constraint Constraint) bool {
	return hasMatches(&wt.WordleTreeNode, constraint, 0)
}

func hasMatches(node *WordleTreeNode, constraint Constraint, depth int) bool {
	if depth >= WordLength {
		for _, countConstraint := range constraint.Counts {
			if countConstraint.Count > 0 {
				return false
			}
		}
		return true
	}

	options := node.Options & constraint.Letters[depth]
	for _, char := range options.ToChars() {
		if hasMatches(node.Children[char], constraint.Dec(char), depth+1) {
			return true
		}
	}

	return false
}

func (wt *WordleTree) CountMatches(constraint Constraint) int {
	return countMatches(&wt.WordleTreeNode, constraint, 0)
}

func countMatches(node *WordleTreeNode, constraint Constraint, depth int) (matches int) {
	if depth >= WordLength {
		for _, countConstraint := range constraint.Counts {
			if countConstraint.Count > 0 {
				return 0
			}
		}
		return 1
	}

	options := node.Options & constraint.Letters[depth]
	for _, char := range options.ToChars() {
		matches += countMatches(node.Children[char], constraint.Dec(char), depth+1)
	}

	return matches
}

func (wt *WordleTree) utility(constraints []Constraint) (utility float64) {
	wordCount := float64(wt.WordCount)

	for _, constraint := range constraints {
		matches := wt.CountMatches(constraint)
		if matches > 0 {
			p := float64(matches) / wordCount
			// p * log(1/p) = p * -log(p)
			utility -= p * math.Log2(p)
		}
	}

	return utility
}

func NewConstraintMap(wordles []string, worldeTree *WordleTree) ConstraintMap {
	constraintMap := make(map[string][]Constraint, len(wordles))

	for _, word := range wordles {
		constraints := make([]Constraint, 0, NumPatterns)

		for _, pattern := range AllPatterns {
			if !isValidPattern(word, pattern) {
				continue
			}

			// Only include constraints with matches
			c := constraintFromPattern(word, pattern)
			if worldeTree.HasMatches(c) {
				constraints = append(constraints, c)
			}
		}

		constraintMap[word] = constraints
	}

	return constraintMap
}

func isValidPattern(word string, pattern Pattern) bool {
	for i := range WordLength {
		if pattern[i] == LetterWrong {
			for j := i + 1; j < WordLength; j++ {
				// Transpositions(Yellow) of the same letter may not
				// occur after a wrong letter, they must be first.
				if word[i] == word[j] && pattern[j] == LetterTransposed {
					return false
				}
			}
		}
	}

	return true
}

func constraintFromPattern(word string, pattern Pattern) Constraint {
	constraint := NewConstraint()
	counts := make(map[byte]byte, WordLength)

	// Letter constraints
	for i := range WordLength {
		char := word[i]

		switch pattern[i] {
		case LetterCorrect:
			constraint.Letters[i] = NewLetterConstraint(char)
			counts[char]++
		case LetterTransposed:
			constraint.Letters[i] = constraint.Letters[i].Exclude(char)
			counts[char]++
		case LetterWrong:
			constraint.Letters[i] = constraint.Letters[i].Exclude(char)

			var hasTransposedDuplicate bool
			// Duplicate yellows can only occur before wrong letters.
			for j := range i {
				hasTransposedDuplicate = hasTransposedDuplicate || (word[j] == char && pattern[j] == LetterTransposed)
			}
			// A duplicate yellow of the same char means we
			// cannot exclude 'char' from the rest of the pattern.
			if !hasTransposedDuplicate {
				for j := range WordLength {
					// Only exclude in places which are not same char & not correct
					if word[j] != char || pattern[j] != LetterCorrect {
						constraint.Letters[j] = constraint.Letters[j].Exclude(char)
					}
				}
			}
		}
	}

	// Count constraints
	var i int
	for char, count := range counts {
		constraint.Counts[i] = MinCountConstraint{char, count}
		i++
	}

	return constraint
}

// Generates all patterns (i.e. permutations of hints)
func generateAllPatterns() []Pattern {
	patterns := make([]Pattern, NumPatterns)
	p := Pattern{LetterWrong, LetterWrong, LetterWrong, LetterWrong, LetterWrong}

	// Find all permutations by counting in base 3
	for i := range NumPatterns {
		patterns[i] = p
		p[0]++

		for j := range WordLength - 1 {
			// Carry over
			if p[j] > LetterCorrect {
				p[j] = LetterWrong
				p[j+1]++
			}
		}
	}

	return patterns
}

type Solver struct {
	state         GameState
	constraints   Constraint
	solutionTree  *WordleTree
	guessTree     *WordleTree
	constraintMap map[string][]Constraint
	numGuesses    int
}

func NewSolver() Solver {
	return Solver{
		state:         NewGameState(),
		constraints:   NewConstraint(),
		solutionTree:  SolutionTree,
		guessTree:     GuessTree,
		constraintMap: GuessConstraintMap,
	}
}

// Pad string with spaces, takes into account unprintable ANSI control sequences
func padRightLines(lines []string) {
	maxPrintLength := 0
	invisibleChars := make([]int, len(lines))
	for i, line := range lines {
		printLength := 0
		inControl := false
		for _, r := range line {
			if inControl {
				invisibleChars[i]++
				if r == 'm' {
					inControl = false
				}
			} else if unicode.IsControl(r) {
				inControl = true
				invisibleChars[i]++
			} else {
				printLength++
			}
		}
		maxPrintLength = max(maxPrintLength, printLength)
	}

	for i, line := range lines {
		lines[i] = fmt.Sprintf("%-*s", maxPrintLength+invisibleChars[i], line)
	}
}

func (s Solver) String() string {
	lines := []string{
		fmt.Sprintf("%sUncertainty: %.3f bits", ansiUnderlined, math.Log2(float64(s.solutionTree.WordCount))),
		"",
	}
	lines = append(lines, strings.Split(s.state.String(), "\n")...)
	padRightLines(lines)

	lines[0] += "  |  Expected utility in bits     " + ansiNotUnderlined
	lines[1] += "  |                |"
	numCols := 2
	tableLines := lines[2:]
	utilities := s.topNWords(len(tableLines) * numCols)
	for i := range tableLines {
		if i < len(utilities) {
			tableLines[i] = fmt.Sprintf("  %s|  %s  %.3f", tableLines[i], utilities[i].Word, utilities[i].Utility)
		} else {
			tableLines[i] = fmt.Sprintf("  %s|  %s  %s", tableLines[i], "     ", "     ")
		}
	}
	for i := len(tableLines); i < len(tableLines)*2; i++ {
		if i < len(utilities) {
			tableLines[i%len(tableLines)] += fmt.Sprintf("  |  %s  %.3f", utilities[i].Word, utilities[i].Utility)
		} else {
			tableLines[i%len(tableLines)] += fmt.Sprintf("  |  %s  %s", "     ", "     ")
		}
	}

	return strings.Join(lines, "\n")
}

func (s *Solver) AddGuess(word string, hints [5]Hint) (bool, error) {
	if s.numGuesses >= MaxGuesses {
		return true, nil
	}

	if err := s.state.AddGuess(word, hints); err != nil {
		return false, err
	}

	allCorrect := true
	for _, s := range hints {
		allCorrect = allCorrect && s == LetterCorrect
	}
	if allCorrect {
		return true, nil
	}

	s.constraints = s.constraints.And(constraintFromPattern(word, Pattern(hints)))
	s.solutionTree = NewWordleTree(s.solutionTree.Wordles, s.constraints)
	s.guessTree = NewWordleTree(s.guessTree.Wordles, s.constraints)
	s.constraintMap = NewConstraintMap(s.guessTree.Wordles, s.solutionTree)
	s.numGuesses++
	return s.numGuesses >= MaxGuesses, nil
}

func (s *Solver) topNWords(n int) []WordUtility {
	if s.solutionTree.WordCount == 1 {
		return []WordUtility{{s.solutionTree.Wordles[0], 0.0}}
	}

	utilities := make([]WordUtility, len(s.guessTree.Wordles))

	maxRoutines := 1000
	sem := make(chan struct{}, maxRoutines)
	var wg sync.WaitGroup
	for i, word := range s.guessTree.Wordles {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, word string) {
			defer wg.Done()
			defer func() { <-sem }()
			utilities[i].Word = word
			utilities[i].Utility = s.solutionTree.utility(s.constraintMap[word])
		}(i, word)
	}
	wg.Wait()

	sort.Slice(utilities, func(i int, j int) bool {
		return utilities[i].Utility > utilities[j].Utility
	})

	n = min(n, len(utilities))
	return utilities[:n]
}
