package domain

// Difficulty labels target puzzle generation & grading.
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
	Expert
)

// StrategyTier limits hinting/logic complexity used.
type StrategyTier int

const (
	StrategySingles StrategyTier = iota // singles / sole candidates
	StrategyPairs                       // naked/hidden pairs
	StrategyAdvanced                    // pointing/claiming, triples, etc.
	StrategyXWing                       // advanced fish (placeholder for cap)
)