package service

import (
	"go.uber.org/mock/gomock"
)

func mockSearchIndex(ctrl *gomock.Controller) *MockSearchService {
	svc := NewMockSearchService(ctrl)
	svc.EXPECT().EnqueueIndex(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}

func mockSearchDelete(ctrl *gomock.Controller) *MockSearchService {
	svc := NewMockSearchService(ctrl)
	svc.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}

func mockSearchDeleteByScope(ctrl *gomock.Controller) *MockSearchService {
	svc := NewMockSearchService(ctrl)
	svc.EXPECT().DeleteByScope(gomock.Any(), gomock.Any()).Return(nil)
	return svc
}
