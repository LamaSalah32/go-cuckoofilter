package cuckoofilter

func swap[T any](s []T, i, j uint) {
    s[i], s[j] = s[j], s[i]
}