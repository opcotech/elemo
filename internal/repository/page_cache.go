package repository

import "fmt"

func normalizedPage(page CursorPage) (CursorPage, error) {
	return page.Normalize()
}

func pageTokenValue(token *string) string {
	if token == nil {
		return ""
	}

	return *token
}

func projectionCacheValue(proj any) string {
	return fmt.Sprintf("%+v", proj)
}
