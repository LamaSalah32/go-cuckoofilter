package cuckoofilter

func GenerateCombinations(n, k byte) [][]byte {
	result := [][]byte{}
	comb := make([]byte, k)

	var backtrack func(pos, start byte)
	backtrack = func(pos, start byte) {
		if pos == k {
			tmp := make([]byte, k)
			copy(tmp, comb)
			result = append(result, tmp)
			return
		}

		for i := start; i < n; i++ {
			comb[pos] = i
			backtrack(pos+1, i)
		}
	}

	backtrack(0, 0)
	return result
}
