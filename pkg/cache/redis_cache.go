package cache

import (
	AVITO "avito/elements"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	cache    *redis.Client
	duration time.Duration
}

func NewCacheImages(cache *redis.Client, duration time.Duration) *RedisCache {
	return &RedisCache{cache: cache, duration: duration}
}

//-------------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------------

// Сохранение баннера в redis
func (myCache *RedisCache) SetBanner(banner AVITO.Banner, tags []int, feature_id int) (errStatus int, err error) {
	ctx := context.Background()

	//Чтобы получить []byte баннера
	chBanner := make(chan []byte)
	//Чтобы получить строки с тэгами
	chTags := make(chan string, len(tags))

	//Для сохранения баннера преобразуем его  в []byte
	go func(banner AVITO.Banner) {
		bytes_, err := json.Marshal(banner)
		//Если ошибка то возвращаем пустой массив
		if err != nil {
			bytes_ = make([]byte, 0)
			chBanner <- bytes_
		}
		chBanner <- bytes_

	}(banner)

	//Для схранения строк тэгов
	tagsStrings := make([]string, len(tags))

	//Для получения строк тэгов
	go func(tags []int, feature_id int) {
		//В цикле проходимся по всем тэгам и создаём строку
		//Для сохранения в Redis.строка : 'featureId_tag_id'
		for _, tag := range tags {
			str := strconv.Itoa(feature_id) + "_" + strconv.Itoa(tag)
			chTags <- str
		}
	}(tags, feature_id)

	bannerBytes := <-chBanner
	//Проверяем получилось ли преобразовать баннер в []byte
	if len(bannerBytes) == 0 {
		return http.StatusInternalServerError, errors.New("при сохранении в redis не смог конвертировать баннер в []byte")
	}

	//Получаем строки
	for i := range tags {
		tagsStrings[i] = string(<-chTags)
	}

	//Сохраняем в Redis баннер
	myCache.cache.Set(ctx, strconv.Itoa(banner.Id), bannerBytes, myCache.duration)

	//Сохраняем тэги
	for _, v := range tagsStrings {
		myCache.cache.Set(ctx, v, strconv.Itoa(banner.Id), myCache.duration)
	}

	return 0, nil
}

//-------------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------------

func (myCache *RedisCache) GetBanner(featureId int, tagId int, useLastRevision bool, isAdmin bool) (banner AVITO.Banner, err error) {
	//Если нужно использовать последнюю версию
	//Не берём данные из Redis, а сразу идем в DB
	if useLastRevision {
		return banner, errors.New("redis: nil")
	}

	ctx := context.Background()

	var mutex = &sync.Mutex{}

	//Создаём строку 'featurId_tagId' для получения баннера id баннера
	str := strconv.Itoa(featureId) + "_" + strconv.Itoa(tagId)
	mutex.Lock()
	id, err := myCache.cache.Get(ctx, str).Result()
	mutex.Unlock()
	if err != nil {
		return banner, err
	}

	mutex.Lock()
	//По id пытаемся получить []byte баннера
	bytesSlice, err := myCache.cache.Get(ctx, id).Result()
	mutex.Unlock()
	if err != nil {
		return banner, err
	}

	//Преобразуем строку байт в banner
	err = json.Unmarshal([]byte(bytesSlice), &banner)
	if err != nil {
		return banner, err
	}

	//Проверка, если пользователь не админ он не может получить banner
	//У которого is_active = false
	if !banner.Is_active && !isAdmin {
		return banner, errors.New("redis: nil")
	}

	return banner, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------

// Удаляем баннер
func (myCache *RedisCache) DeleteBanner(bannerId int) error {
	ctx := context.Background()

	//Удаляем баннер с указаным id
	//Result  возвращает кол-во удаленных записей, нам это ненужно
	_, err := myCache.cache.Del(ctx, strconv.Itoa(bannerId)).Result()
	if err != nil {
		return err
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------

// Получаем пользователя по токену
func (myCache *RedisCache) GetUser(userToken string) (user AVITO.User, err error) {
	ctx := context.Background()

	var mutex = &sync.Mutex{}

	mutex.Lock()
	bytesSlice, err := myCache.cache.Get(ctx, userToken).Result()
	mutex.Unlock()
	if err != nil {
		return user, err
	}

	//Преобразуем строку байт в user
	err = json.Unmarshal([]byte(bytesSlice), &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------------------------------------------------

// Сохраняем пользователя
func (myCache *RedisCache) SetUser(user AVITO.User) (err error) {
	ctx := context.Background()

	bytesSlice, err := json.Marshal(user)
	if err != nil {
		return err
	}
	//Сохраняем в Redis user на 30 минут
	err = myCache.cache.Set(ctx, user.Token, bytesSlice, time.Minute*30).Err()
	if err != nil {
		return err
	}

	return nil
}
