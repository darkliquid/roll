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

	// Extra rules
	tFAILURES
	tEXPLODE
	tCOMPOUND
	tPENETRATE
	tKEEPHIGH
	tKEEPLOW
	tDROPHIGH
	tDROPLOW
	tREROLL
	tSORT

	// Tests
	tGREATER
	tLESS
	tEQUAL

	// Grouping
	tGROUPSTART
	tGROUPEND
	tGROUPSEP

	// Math operators and functions
	tMULT
	tDIV
	tMOD
	tPOW
	tLPAREN
	tRPAREN
	tFUNC
	tMATCH
)
