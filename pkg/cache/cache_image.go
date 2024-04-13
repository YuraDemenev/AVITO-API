package cache

import (
	AVITO "avito/elements"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheImages interface {
	SetBanner(banner AVITO.Banner, tags []int, feature_id int) (errStatus int, err error)
	GetBanner(featureId int, tagId int, useLastRevision bool, isAdmin bool) (banner AVITO.Banner, err error)
	DeleteBanner(bannerId int) error
	GetUser(userToken string) (user AVITO.User, err error)
	SetUser(user AVITO.User) (err error)
}

type Cache struct {
	CacheImages
}

func NewCache(cache *redis.Client, duration time.Duration) *Cache {
	return &Cache{
		CacheImages: NewCacheImages(cache, duration),
	}

}
