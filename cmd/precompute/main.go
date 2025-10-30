package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/Backshifted/wordle-solver/assets"
	"github.com/Backshifted/wordle-solver/pkg/wordle"
)

func main() {
	solutionTree := wordle.NewWordleTree(assets.Wordles, wordle.NewConstraint())
	writeObject(solutionTree, "assets/solution-tree.bin")

	guessTree := wordle.NewWordleTree(assets.WordlesAndNonWordles, wordle.NewConstraint())
	writeObject(guessTree, "assets/guess-tree.bin")

	guessConstraintMap := wordle.NewConstraintMap(assets.WordlesAndNonWordles, wordle.SolutionTree)
	writeObject(guessConstraintMap, "assets/constraint-map.bin")
}

func writeObject(obj any, path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to open file '%s': %v", path, err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(obj); err != nil {
		log.Fatal("Encoding error:", err)
	}

	fmt.Printf("Wrote object to '%s'\n", path)
}
