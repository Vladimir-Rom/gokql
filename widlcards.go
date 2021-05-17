package gokql

import (
	"strings"
)

type wildcard struct {
	parts     []string
	firstStar bool
	lastStar  bool
}

func newWildcard(wildcards string) wildcard {
	if len(wildcards) == 0 {
		return wildcard{}
	}
	parts := strings.Split(wildcards, "*")
	firstStar := wildcards[0] == '*'
	lastStar := wildcards[len(wildcards)-1] == '*'
	if firstStar && len(parts) > 0 {
		parts = parts[1:]
	}
	if lastStar && len(parts) > 0 {
		parts = parts[0 : len(parts)-1]
	}

	return wildcard{
		parts:     parts,
		firstStar: firstStar,
		lastStar:  lastStar,
	}
}

func (w wildcard) Match(str string) bool {
	lenStr := len(str)
	if len(w.parts) == 0 {
		return true
	}
	parts := w.parts
	strIndex := 0
	if !w.firstStar {
		if !strings.HasPrefix(str, parts[0]) {
			return false
		}
		strIndex = len(parts[0])
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return w.lastStar || strIndex == lenStr
	}

	lastPart := parts[len(parts)-1]

	if !w.lastStar {
		parts = parts[0 : len(parts)-1]
	}

	for partIndex, part := range parts {
		findResult := strings.Index(str[strIndex:], part)
		if findResult == -1 {
			return false
		}

		strIndex += findResult + len(part)
		if strIndex >= lenStr {
			return partIndex == len(parts)-1
		}
	}

	if !w.lastStar {
		return strings.HasSuffix(str[strIndex:], lastPart)
	} else {
		return true
	}
}
