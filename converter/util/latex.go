package util

import (
	"log"
	"strings"
)

/*
Breaks the given string into slices of 'linesPerPage' lines based on the
given separator 'sep'.
*/
func SliceContent(str string, linesPerPage int, sep string) []string {
	contentLines := strings.SplitAfter(str, sep)
	pages := []string{}

	for i := range (len(contentLines) / linesPerPage) + 1 {
		start := i * linesPerPage
		end := min(start+linesPerPage, len(contentLines))
		page := strings.Join(contentLines[start:end], "")
		pages = append(pages, page)
	}

	return pages
}

// assumes the given string is in the third column (considered the last column) of the array environment
func BreakColumn(line string, maxLength int, sep string) string {
	divisions, err := DivideLatexString(line, maxLength)
	if err != nil {
		log.Println("Failed to break Latex string. Keeping the string unmodified...")
		return line
	}

	return strings.Join(divisions, sep)
}

/*
Divides the given string into multiple divisions of at most 'divLen'
characters consider Latex symbols escaped with a backslash.
Each division ends on a Latex symbol.

Whitespaces and OS new lines are not counted.
Latex new lines are treated as one character and do NOT reset the char
count of the current division.
*/
func DivideLatexString(str string, divLen int) ([]string, error) {
	if divLen < 1 {
		return []string{}, NewError(ErrZeroOrNegDivisionLength)
	}

	divisions, charCount, divStart, isInsideSymbol, toAdd := []string{}, 0, 0, false, 0

	i := 0
	for i < len(str) {
		toAdd, isInsideSymbol = countSymbol(str[i], isInsideSymbol)
		charCount += toAdd

		if charCount >= divLen && !isInsideSymbol {
			lastSymbol := findLastSymbol(str, i)
			divisions = append(divisions, str[divStart:lastSymbol+1])
			divStart = lastSymbol + 1
			i, charCount = divStart, 0
			continue
		}
		i++
	}

	if divStart < len(str) {
		divisions = append(divisions, str[divStart:])
	}
	return divisions, nil
}

func countSymbol(char byte, isInsideSymbol bool) (int, bool) {
	if char == ' ' || char == '\n' {
		return 0, false
	}

	if char == '\\' && !isInsideSymbol {
		return 1, true
	}

	if !isInsideSymbol {
		return 1, false
	}

	return 0, isInsideSymbol
}

func findLastSymbol(str string, currPos int) int {
	if currPos <= 0 {
		return 0
	}

	lastSpace := currPos
	for i := currPos; i >= 0; i-- {
		if str[i] == ' ' {
			lastSpace = i
			continue
		}

		if str[i] == '\\' {
			return lastSpace
		}
	}

	return currPos
}
