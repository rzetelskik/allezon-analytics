package forwarder

func Backtrack(pos int, curr []string, ss []string, res *[][]string) {
	currDup := make([]string, len(curr))
	copy(currDup, curr)
	*res = append(*res, currDup)

	for i := pos; i < len(ss); i++ {
		curr = append(curr, ss[i])
		Backtrack(i+1, curr, ss, res)
		curr = curr[:len(curr)-1]
	}
}
