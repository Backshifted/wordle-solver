package assets

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"log"
)

var (
	//go:embed wordles.json
	WordlesFile []byte
	//go:embed nonwordles.json
	NonwordlesFile []byte
	//go:embed solution-tree.bin
	SolutionTreeFile []byte
	//go:embed guess-tree.bin
	GuessTreeFile []byte
	//go:embed constraint-map.bin
	ConstraintMapFile []byte

	Wordles              = LoadJsonStringArray(WordlesFile)
	NonWordles           = LoadJsonStringArray(NonwordlesFile)
	WordlesAndNonWordles = append(Wordles, NonWordles...)
)

func LoadJsonStringArray(file []byte) []string {
	var wordles []string
	if err := json.Unmarshal(file, &wordles); err != nil {
		log.Fatalf("Failed to parse embedded JSON: %v", err)
	}

	return wordles
}

func Load[T any](data []byte) T {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)

	var obj T
	if err := dec.Decode(&obj); err != nil {
		log.Fatalf("Failed to decode struct: %v", err)
	}

	return obj
}
