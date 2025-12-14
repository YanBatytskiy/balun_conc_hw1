package compute

func IsAnyLetter(symbol rune) bool {
	return (symbol >= LetterRangeLower[0] &&
		symbol <= LetterRangeLower[1]) ||
		(symbol >= LetterRangeUpper[0] &&
			symbol <= LetterRangeUpper[1])
}

func IsUpperLetter(symbol rune) bool {
	return symbol >= LetterRangeUpper[0] && symbol <= LetterRangeUpper[1]
}

func IsDigit(symbol rune) bool {
	return symbol >= DigitRange[0] && symbol <= DigitRange[1]
}

func IsPunctuation(symbol rune) bool {
	for _, pct := range Punctuation {
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
