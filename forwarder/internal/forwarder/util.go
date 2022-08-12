package forwarder

func Backtrack(curr []string, ss []string, res *[][]string) {
	backtrack(0, curr, ss, res)
}

func backtrack(pos int, curr []string, ss []string, res *[][]string) {
	currDup := make([]string, len(curr))
	copy(currDup, curr)
	*res = append(*res, currDup)

	for i := pos; i < len(ss); i++ {
		curr = append(curr, ss[i])
		backtrack(i+1, curr, ss, res)
		curr = curr[:len(curr)-1]
	}
}
