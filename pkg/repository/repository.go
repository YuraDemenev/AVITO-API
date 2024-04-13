package repository

import (
	AVITO "avito/elements"

	"github.com/jmoiron/sqlx"
)

type Users interface {
	CheckUserToken(token string) (user AVITO.User, errStatus int, err error)
}

type Banners interface {
	GetBanner(tag int, feature int, isAdmin bool) (banner AVITO.Banner, errStatus int, err error)
	AdminGetBanners(tag int, feature int, limit int, offset int) (banners []AVITO.Banner, tags [][]int, errStatus int, err error)
	CreateNewBanner(tags []int, feature int, content AVITO.Banner, is_active bool) (banner AVITO.Banner, errStatus int, err error)
	ChangeBanner(banner_id int, tags []int, feature int, content AVITO.Banner) (banner AVITO.Banner, curTags []int, curFeaturId int, errStatus int, err error)
	DeleteBanner(banner_id int) (errStatus int, err error)
}

type Repository struct {
	Users
	Banners
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Users:   NewUsersPostgres(db),
		Banners: NewBannersPostgres(db),
	}

}
