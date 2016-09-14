package roll

// Token is an input token
type Token int

const (
	tILLEGAL Token = iota
	tEOF
	tWS

	// Literals
	tNUM
	tDIE

	// Modifiers
	tPLUS
	tMINUS
)
