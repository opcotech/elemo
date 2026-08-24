package service_test

import (
	"go.uber.org/mock/gomock"

	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

func mockSearchIndex(ctrl *gomock.Controller) *mocksvc.MockSearchService {
	svc := mocksvc.NewMockSearchService(ctrl)
	svc.EXPECT().EnqueueIndex(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}

func mockSearchDelete(ctrl *gomock.Controller) *mocksvc.MockSearchService {
	svc := mocksvc.NewMockSearchService(ctrl)
	svc.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}

func mockSearchDeleteByScope(ctrl *gomock.Controller) *mocksvc.MockSearchService {
	svc := mocksvc.NewMockSearchService(ctrl)
	svc.EXPECT().DeleteByScope(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}
