package main

import (
	AVITO "avito/elements"
	"avito/pkg/cache"
	"avito/pkg/handler"
	"avito/pkg/repository"
	"avito/pkg/service"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	//Загржаем данные из config.yml
	err := initConfig()
	if err != nil {
		logrus.Fatalf("initializing config error: %s", err.Error())
	}

	//Подключаемся к базе данных
	//Берем данные из cofig.yml и передаем в качестве
	//струкуры DBConfig в NewPostgresDB
	db, err := repository.NewPostgresDB(repository.DBConfig{
		Host:     viper.GetString("dbPostgres.host"),
		Port:     viper.GetString("dbPostgres.port"),
		UserName: viper.GetString("dbPostgres.username"),
		DBName:   viper.GetString("dbPostgres.dbname"),
		SSLMode:  viper.GetString("dbPostgres.sslmode"),
		Password: viper.GetString("dbPostgres.password"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	//Подключаемся к Redis
	//Берем данные из cofig.yml и передаем в качестве
	//струкуры RedisConfig в NewPostgresDB и передаем
	//5 минут (продолжительность "жизни" данных)
	redisCache := cache.NewRedisCache(cache.RedisConfig{
		Host: viper.GetString("redis.host"),
		DB:   viper.GetInt("redis.db"),
	})

	redis := cache.NewCache(redisCache, time.Minute*5)

	//Инициализируем repository (сервис через который
	//будут запускаться функции взаимодействующие с бд)
	repos := repository.NewRepository(db)

	//Инициализируем service (через него будут запускаться
	//Промежуточные функции и будет запускаться взаимодействие с БД
	//через repository)
	services := service.NewService(repos, redis)

	//Инициализируем hadlers (пути по которым можно отправлять запросы)
	handlers := handler.NewHandler(services)

	//Инициализируем сервер
	server := new(AVITO.Server)

	//Запускаем сервер
	err = server.Run(viper.GetString("port"), handlers.InitRoutes())
	if err != nil {
		logrus.Fatalf("error while running server: %s", err.Error())
	}
	fmt.Println("Site run")
}

// Функция для загрузки данных из config.yml
func initConfig() error {
	viper.AddConfigPath("../config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
