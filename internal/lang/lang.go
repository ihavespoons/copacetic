/*
Copyright © 2024 Ben Gittins
*/

package lang

type LanguageSet map[string]int

func (l LanguageSet) Has(k string) bool {
	_, ok := l[k]
	return ok
}
