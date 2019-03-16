package hocr

type stack []string

func (s stack) push(str string) stack {
	return append(s, str)
}

func (s stack) pop() stack {
	// do not check for errors
	return s[0 : len(s)-1]
}

// match the given strings with the top of the stack
func (s stack) match(strs ...string) bool {
	if len(strs) > len(s) {
		return false
	}
	for i, j := len(s), len(strs); i > 0 && j > 0; i, j = i-1, j-1 {
		if s[i-1] != strs[j-1] {
			return false
		}
	}
	return true
}
