package models

type (
	Submission struct {
		ID          uint32
		SourcePath  string
		Language    string
		ProblemID   string
		TestCount   uint16
		Constraints *Constraints
	}

	Constraints struct {
		IsInteractive    bool
		TimeLimit        float32
		MemoryLimit      uint32
		OutputLimit      uint32
		AllowPartial     bool
		AllowedLanguages []string
		ShortCircuit     bool
	}
)
