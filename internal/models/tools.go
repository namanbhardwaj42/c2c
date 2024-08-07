// tools.go

package models

import "strings"

func contains(element string, data []string) bool {
	for _, v := range data {
		if element == v {
			return true
		}
	}
	return false
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func clean_genres(genres []string, max_length int) string {

	all_genres := []string{}
	for _, g := range genres {
		all_genres = append(all_genres, strings.Split(g, " / ")...)
	}
	genres_as_string := strings.Join(all_genres[:min(len(all_genres), max_length)], "/")
	// genres_as_string = strings.Replace(genres_as_string, " / ", "/", 1)

	return genres_as_string
}
