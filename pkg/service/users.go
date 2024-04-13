package service

import (
	"avito/pkg/cache"
	"avito/pkg/repository"

	"github.com/sirupsen/logrus"
)

type UsersService struct {
	repo        repository.Users
	cacheClient cache.CacheImages
}

func NewUsersService(repo repository.Users, cache cache.CacheImages) *UsersService {
	return &UsersService{repo: repo, cacheClient: cache}
}

// Проверка токена пользователя
func (u *UsersService) CheckUserToken(token string) (bool, int, error) {
	var errStatus int

	//Сначала проверяем есть ли пользователь в redis
	user, err := u.cacheClient.GetUser(token)

	//Если не получилось взять пользователя из redis
	if err != nil {
		//Если это не ошибка об отсутствии баннера, log error
		if err.Error() != "redis: nil" {
			logrus.Errorf("Ошибка при попытки получить данные из redis. Ошибка :%s", err.Error())
		}

		user, errStatus, err = u.repo.CheckUserToken(token)
		//Сохраняем в redis пользователя
		if err == nil {
			errRedis := u.cacheClient.SetUser(user)
			if errRedis != nil {
				logrus.Errorf("не удалось сохранить пользователя в redis. Ошибка :%s", errRedis.Error())
			}
		}
	}

	return user.IsAdmin, errStatus, err
}
