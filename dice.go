package roll

import (
	"fmt"
	"math/rand"
	"strconv"
)

// DieRoll is the result of a die roll
type DieRoll struct {
	Result int
	Symbol string
}

// Die is the interface allDice must confirm to
type Die interface {
	Roll() DieRoll
	String() string
}

// FateDie is a die representing the typical Fate/Fudge die
type FateDie int

const (
	// FateBlank is the symbol for a blank Fate die face
	FateBlank = "☐"
	// FateMinus is the symbol for a minus Fate die face
	FateMinus = "⊟"
	// FatePlus is the symbol for a plus Fate die face
	FatePlus = "⊞"
)

// Roll generates a random number and the appropriate symbol
func (d FateDie) Roll() DieRoll {
	val := rand.Intn(3) - 1
	sym := FateBlank

	switch val {
	case -1:
		sym = FateMinus
	case 1:
		sym = FatePlus
	}

	return DieRoll{
		Result: val,
		Symbol: sym,
	}
}

// String returns the string representation of the FateDie type
func (d FateDie) String() string {
	return "dF"
}

// NormalDie is a die representing an N-sided die
type NormalDie int

// Roll generates a random number and the appropriate symbol
func (d NormalDie) Roll() DieRoll {
	val := rand.Intn(int(d)) + 1
	sym := strconv.Itoa(val)
	return DieRoll{
		Result: val,
		Symbol: sym,
	}
}

// String returns the string representation of the NormalDie type
func (d NormalDie) String() string {
	return fmt.Sprintf("d%d", d)
}
