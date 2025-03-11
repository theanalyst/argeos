package list

import "strings"

type StringList []string

func (l *StringList) DecodeFromString(raw, separator string) {
	if separator == "" {
		panic("separator cannot be empty")
	}

	*l = strings.Split(raw, separator)
}

func (l StringList) Contains(s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}
