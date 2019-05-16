package commandargs

import (
	"regexp"
)

var (
	shellsplitRegexp = regexp.MustCompile(`\s*(([^\s\\\'\"]+)|'([^\']*)'|"((?:[^\"\\]|\\.)*)")`)
)

func splitSshArgs(command string) []string {
	matches := shellsplitRegexp.FindAllStringSubmatch(command, -1)

	var words []string
	for _, match := range matches {
		word := match[2]
		sqWord := match[3]
		dqWord := match[4]

		if dqWord != "" {
			words = append(words, dqWord)
		} else if sqWord != "" {
			words = append(words, sqWord)
		} else {
			words = append(words, word)
		}
	}

	return words
}
