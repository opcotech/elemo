package http

import (
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

const DefaultPageSize = 100

func cursorPageFromParams(pageSize *int, pageToken *string) (service.CursorPage, error) {
	size := DefaultPageSize
	if pageSize != nil {
		size = *pageSize
	}
	page, err := service.CursorPage{
		Size:  size,
		Token: pageToken,
	}.Normalize()
	if err != nil {
		return service.CursorPage{}, err
	}
	if page.Token != nil {
		if _, err := repository.DecodeCursor(*page.Token); err != nil {
			return service.CursorPage{}, err
		}
	}
	return page, nil
}

func pageInfoToDTO(info service.PageInfo) api.PageInfo {
	return api.PageInfo{
		HasMore:       info.HasMore,
		NextPageToken: info.NextPageToken,
		TotalCount:    info.TotalCount,
	}
}
