package service

import "github.com/opcotech/elemo/internal/repository"

type (
	CursorPage  = repository.CursorPage
	PageInfo    = repository.PageInfo
	Page[T any] = repository.Page[T]
)

func mapPage[In any, Out any](page repository.Page[In], fn func(In) Out) Page[Out] {
	items := make([]Out, len(page.Items))
	for i, item := range page.Items {
		items[i] = fn(item)
	}

	return Page[Out]{
		Items:    items,
		PageInfo: page.PageInfo,
	}
}
