package compute

import (
	"lesson1/internal/command"
)

func IsAnyLetter(symbol rune) bool {
	return (symbol >= command.LetterRangeLower[0] &&
		symbol <= command.LetterRangeLower[1]) ||
		(symbol >= command.LetterRangeUpper[0] &&
			symbol <= command.LetterRangeUpper[1])
}

func IsUpperLetter(symbol rune) bool {
	return symbol >= command.LetterRangeUpper[0] && symbol <= command.LetterRangeUpper[1]
}

func IsDigit(symbol rune) bool {
	return symbol >= command.DigitRange[0] && symbol <= command.DigitRange[1]
}

func IsPunctuation(symbol rune) bool {
	for _, pct := range command.Punctuation {
		if symbol == pct {
			return true
		}
	}
	return false
}

func ValidateCommand(raw string) bool {
	for _, symbol := range raw {
		ok := IsUpperLetter(symbol)
		if !ok {
			return false
		}
	}
	return true
}

func ValidateArgument(raw string) bool {
	for _, symbol := range raw {
		ok := IsAnyLetter(symbol) || IsDigit(symbol) || IsPunctuation(symbol)
		if !ok {
			return false
		}
	}
	return true
}
