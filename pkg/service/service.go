package service

import (
	AVITO "avito/elements"
	"avito/pkg/cache"
	"avito/pkg/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Users interface {
	CheckUserToken(token string) (bool, int, error)
}

type Banners interface {
	GetBanner(tag int, feature int, useLastRevision bool, isAdmin bool) (title string, text string, url string, errStatus int, err error)
	AdminGetBanners(tag int, feature int, limit int, offset int) (banners []AVITO.Banner, tags [][]int, errStatus int, err error)
	CreateNewBanner(tags []int, feature int, content AVITO.Banner, is_active bool) (banner_id, errStatus int, err error)
	ChangeBanner(banner_id int, tags []int, feature int, content AVITO.Banner) (errStatus int, err error)
	DeleteBanner(banner_id int) (errStatus int, err error)
}

type Service struct {
	Users
	Banners
}

func NewService(repos *repository.Repository, cache *cache.Cache) *Service {
	return &Service{
		Users:   NewUsersService(repos.Users, cache.CacheImages),
		Banners: NewBannersService(repos.Banners, cache.CacheImages),
	}
}
