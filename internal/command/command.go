package command

const (
	CommandSet = "SET"
	CommandGet = "GET"
	CommandDel = "DEL"
)

var (
	Punctuation      = []rune{'*', '/', '_', '.'}
	LetterRangeLower = [2]rune{'a', 'z'}
	LetterRangeUpper = [2]rune{'A', 'Z'}
	DigitRange       = [2]rune{'0', '9'}

	CommandSetQ = 2
	CommandGetQ = 1
	CommandDelQ = 1
)
