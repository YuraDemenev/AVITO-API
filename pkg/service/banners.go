package service

import (
	AVITO "avito/elements"
	"avito/pkg/cache"
	"avito/pkg/repository"

	"github.com/sirupsen/logrus"
)

type BannersService struct {
	repo        repository.Banners
	cacheClient cache.CacheImages
}

func NewBannersService(repo repository.Banners, cache cache.CacheImages) *BannersService {
	return &BannersService{repo: repo, cacheClient: cache}
}

// Получение баннера
func (b *BannersService) GetBanner(tag int, feature int, useLastRevision bool, isAdmin bool) (title string, text string, url string, errStatus int, err error) {
	//Сначало проверяем наличие банера в redis
	banner, err := b.cacheClient.GetBanner(feature, tag, useLastRevision, isAdmin)
	//Если баннера нет в redis  берём его из Postgres
	if err != nil {
		//Если это не ошибка об отсутствии баннера, log error
		if err.Error() != "redis: nil" {
			logrus.Errorf("Ошибка при попытки получить данные из redis. Ошибка :%s", err.Error())
		}

		banner, errStatus, err = b.repo.GetBanner(tag, feature, isAdmin)
		//Если произошла ошибка при получении баннера из Postgres
		//Выходим из функции с ошибкой
		if err != nil {
			return "", "", "", errStatus, err
		}

		//Сохраняем баннер в redis
		tags := []int{tag}
		_, err = b.cacheClient.SetBanner(banner, tags, feature)

		//Если не получилось сохранить баннер в redis логируем ошибку
		if err != nil {
			logrus.Errorf("баннер не сохранился в redis. Ошибка :%s", err.Error())
		}

		return banner.Title, banner.Text, banner.Url, errStatus, err
	}
	//Возвращаем баннер
	return banner.Title, banner.Text, banner.Url, 0, nil

}

// Получение баннеров по тэгу и/или фичи
func (b *BannersService) AdminGetBanners(tag int, feature int, limit int, offset int) (banners []AVITO.Banner, tags [][]int, errStatus int, err error) {
	//Так как здесь нет промежуточных функций, сразу идём в repo
	return b.repo.AdminGetBanners(tag, feature, limit, offset)
}

// Создание нового баннера
func (b *BannersService) CreateNewBanner(tags []int, feature int, content AVITO.Banner, is_active bool) (banner_id, errStatus int, err error) {
	banner, errStatus, err := b.repo.CreateNewBanner(tags, feature, content, is_active)
	//Если нет ошибки, то обновляем данные в redis
	if err == nil {
		_, errRedis := b.cacheClient.SetBanner(banner, tags, feature)
		//Если произошла ошибка при сохранении в redis, логируем её
		if errRedis != nil {
			logrus.Errorf("не удалось сохранить баннер в redis. Ошибка: %s", errRedis.Error())
		}
	}
	return banner.Id, errStatus, err
}

// Изменение существующего банннера
func (b *BannersService) ChangeBanner(banner_id int, tags []int, feature int, content AVITO.Banner) (errStatus int, err error) {
	banner, curTags, curFeatureId, errStatus, err := b.repo.ChangeBanner(banner_id, tags, feature, content)
	//Если нет ошибки, то обновляем данные в redis
	if err == nil {
		_, errRedis := b.cacheClient.SetBanner(banner, curTags, curFeatureId)
		//Если произошла ошибка при сохранении в redis, логируем её
		if errRedis != nil {
			logrus.Errorf("не удалось сохранить баннер в redis. Ошибка: %s", errRedis.Error())
		}

	}
	return errStatus, err

}

// Удаление баннера
func (b *BannersService) DeleteBanner(banner_id int) (errStatus int, err error) {
	//Удаляем баннер из БД
	errStatus, err = b.repo.DeleteBanner(banner_id)
	//Если нет ошибки то удаляем и из redis
	if err == nil {
		errRedis := b.cacheClient.DeleteBanner(banner_id)
		if errRedis != nil {
			logrus.Errorf("баннер не удалился из redis. Ошибка:%s", errRedis.Error())
		}
	}
	return errStatus, err

}
