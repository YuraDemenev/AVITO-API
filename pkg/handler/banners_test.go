package handler

import (
	"avito/pkg/cache"
	"avito/pkg/repository"
	"avito/pkg/service"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Загрузка данных из config.yml
func initConfig() error {
	viper.AddConfigPath("../../config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

// Что бы тесты работали нужно инициализировать все сервисы как в main
func initServices() *Handler {
	//Загржаем данные из config.yml
	err := initConfig()
	if err != nil {
		logrus.Fatalf("initializing config error: %s", err.Error())
	}

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

	redisCache := cache.NewRedisCache(cache.RedisConfig{
		Host: viper.GetString("redis.host"),
		DB:   viper.GetInt("redis.db"),
	})

	redis := cache.NewCache(redisCache, time.Minute*5)
	repos := repository.NewRepository(db)
	services := service.NewService(repos, redis)
	handlers := NewHandler(services)
	return handlers
}

// Тесты проверяющие /user_banner
func TestHandler_userBanner(t *testing.T) {
	//Таблица со всеми тестам
	testTable := []struct {
		name              string
		requestBody       string
		requestHeaderName string
		requestHeaderData string
		expectedBody      string
		expectedCode      int
	}{
		//Тест #1
		{
			name: "Test #1. OK. корректный стандартный запрос",
			requestBody: `{
				"tag_id":1 ,
				"feature_id":2,
				"use_last_revision":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{"title":"second_title","text":"second_text","url":"second_url"}`,
			expectedCode:      http.StatusOK,
		},
		//Тест #2
		{
			name: "Test #2. NOT OK. запрос в котором неверные данные tag_id = 0",
			requestBody: `{
						"tag_id":0 ,
						"feature_id":2,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
		//Тест #3
		{
			name: "Test #3.OK. запрос в котором token принадлежит админу",
			requestBody: `{
						"tag_id":1 ,
						"feature_id":1,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{"title":"first_title","text":"first_text","url":"first_url"}`,
			expectedCode:      http.StatusOK,
		},
		//Тест #4
		{
			name: "Test #4.NOT OK. запрос в котором tag_id, которого нет в БД",
			requestBody: `{
						"tag_id":999999 ,
						"feature_id":1,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Баннер для не найден'}`,
			expectedCode:      http.StatusNotFound,
		},
		//Тест #5
		{
			name: "Test #5.NOT OK. запрос в котором feature_id, которого нет в БД",
			requestBody: `{
							"tag_id":1 ,
							"feature_id":999999,
							"use_last_revision":true
						}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Баннер для не найден'}`,
			expectedCode:      http.StatusNotFound,
		},
		//Тест #6
		{
			name: "Test #6.NOT OK. запрос в котором токен, которого нет в бд",
			requestBody: `{
									"tag_id":1 ,
									"feature_id":2,
									"use_last_revision":true
								}`,
			requestHeaderName: "token",
			requestHeaderData: "something",
			expectedBody:      `{'error':' Пользователь не авторизован'}`,
			expectedCode:      http.StatusUnauthorized,
		},
	}

	//Поднимаем все сервисы
	handlers := initServices()
	router := handlers.InitRoutes()

	//В цикле проходимся по каждому тесту и запускаем его
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/user_banner", strings.NewReader(test.requestBody))

			req.Header.Add(test.requestHeaderName, test.requestHeaderData)
			w := httptest.NewRecorder()
			//Выполняется запрос
			router.ServeHTTP(w, req)
			//Получаем ответ
			responseData, _ := io.ReadAll(w.Body)

			//Проверяем ответ
			assert.Equal(t, test.expectedBody, string(responseData))
			assert.Equal(t, test.expectedCode, w.Code)
		})

	}

}

// Тесты проверяющие /banner
func TestHandler_BannerGet(t *testing.T) {
	//Таблица со всеми тестам
	testTable := []struct {
		name              string
		requestBody       string
		requestHeaderName string
		requestHeaderData string
		expectedBody      string
		expectedCode      int
		time              string
	}{
		//Test #1
		{
			name: "Test #1. OK. корректный стандартный запрос",
			requestBody: `{
						"tag_id": 1,
						"feature_id": 2,
						"limit": 2,
						"offset": 0
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `[{"banner_id":2,"tag_ids":[1,2],"feature_id":2,"content":{"title":"second_title","text":"second_text","url":"second_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"},{"banner_id":4,"tag_ids":[1,2,3],"feature_id":2,"content":{"title":"fourth_title","text":"fourth_text","url":"fourth_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"}]`,
			expectedCode:      http.StatusOK,
			time:              "0000-01-01T14:22:34.336011Z",
		},
		//Тест #2
		{
			name: "Test #2.NOT OK. запрос в котором токен, которого нет в бд",
			requestBody: `{
						"tag_id":1 ,
						"feature_id":2,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "something",
			expectedBody:      `{'error':' Пользователь не авторизован'}`,
			expectedCode:      http.StatusUnauthorized,
		},
		//Тест #3
		{
			name: "Test #3.NOT OK. запрос в котором tag_id, которого нет в БД",
			requestBody: `{
						"tag_id":999999 ,
						"feature_id":1,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Тест #4
		{
			name: "Test #4.NOT OK. запрос в котором feature_id, которого нет в БД",
			requestBody: `{
						"tag_id":1 ,
						"feature_id":999999,
						"use_last_revision":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #5
		{
			name: "Test #5.NOT OK. запрос от обычного пользователя",
			requestBody: `{
						"tag_id": 1,
						"feature_id": 2,
						"limit": 2,
						"offset": 0
					}`,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{'error':' Пользователь не имеет доступа'}`,
			expectedCode:      http.StatusForbidden,
		},
		//Test #6
		{
			name: "Test #6. OK. запрос где нет tag_id",
			requestBody: `{
						"tag_id": 0,
						"feature_id": 2,
						"limit": 2,
						"offset": 0
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `[{"banner_id":2,"tag_ids":[1,2],"feature_id":2,"content":{"title":"second_title","text":"second_text","url":"second_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"},{"banner_id":4,"tag_ids":[1,2,3],"feature_id":2,"content":{"title":"fourth_title","text":"fourth_text","url":"fourth_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"}]`,
			expectedCode:      http.StatusOK,
			time:              "0000-01-01T14:22:34.336011Z",
		},
		//Test #7
		{
			name: "Test #7.NOT OK. запрос где нет feature_id",
			requestBody: `{
						"tag_id": 1,
						"feature_id": 0,
						"limit": 4,
						"offset": 0
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `[{"banner_id":1,"tag_ids":[1],"feature_id":0,"content":{"title":"first_title","text":"first_text","url":"first_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"},{"banner_id":2,"tag_ids":[1,2],"feature_id":0,"content":{"title":"second_title","text":"second_text","url":"second_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"},{"banner_id":4,"tag_ids":[1,2,3],"feature_id":0,"content":{"title":"fourth_title","text":"fourth_text","url":"fourth_url"},"is_active":true,"created_at":"0000-01-01T14:22:34.336011Z","updated_at":"0000-01-01T14:22:34.336011Z"}]`,
			expectedCode:      http.StatusOK,
			time:              "0000-01-01T14:22:34.336011Z",
		},
		//Test #8
		{
			name: "Test #8. OK. запрос где offset > кол-ва баннеров",
			requestBody: `{
				"tag_id": 1,
				"feature_id": 2,
				"limit": 2,
				"offset": 2
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `[]`,
			expectedCode:      http.StatusOK,
		},
		//Test #9
		{
			name: "Test #9. OK. запрос где limit = 0",
			requestBody: `{
				"tag_id": 1,
				"feature_id": 2,
				"limit": 0,
				"offset": 0
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #10
		{
			name: "Test #10. OK. запрос где limit = 1",
			requestBody: `{
				"tag_id": 0,
				"feature_id": 2,
				"limit": 1,
				"offset": 0
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `[{"banner_id":2,"tag_ids":[1,2],"feature_id":2,"content":{"title":"second_title","text":"second_text","url":"second_url"},"is_active":true,"created_at":"0000-01-01T16:11:31.529628Z","updated_at":"0000-01-01T16:11:31.529628Z"}]`,
			expectedCode:      http.StatusOK,
			time:              "0000-01-01T16:11:31.529628Z",
		},
	}

	//Поднимаем все сервисы
	handlers := initServices()
	router := handlers.InitRoutes()

	//В цикле проходимся по каждому тесту и запускаем его
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/banner", strings.NewReader(test.requestBody))

			req.Header.Add(test.requestHeaderName, test.requestHeaderData)
			w := httptest.NewRecorder()
			//Выполняется запрос
			router.ServeHTTP(w, req)
			//Получаем ответ
			responseData, _ := io.ReadAll(w.Body)

			//Структура для возвращения ответа
			type localBannerAnswer struct {
				Banner_id  int          `json:"banner_id"`
				Tag_ids    []int        `json:"tag_ids"`
				Feature_id int          `json:"feature_id"`
				Content    bannerAnswer `json:"content"`
				Is_active  bool         `json:"is_active"`
				Created_at string       `json:"created_at"`
				Updated_at string       `json:"updated_at"`
			}
			bannersAnswer := make([]localBannerAnswer, 0)

			err := json.Unmarshal(responseData, &bannersAnswer)
			//Так как поля created_at, updated_at могут меняться походу разработки,
			//их сравнивать не надо
			if err == nil {
				for i := range bannersAnswer {
					bannersAnswer[i].Created_at = test.time
					bannersAnswer[i].Updated_at = test.time
				}

				responseData, _ = json.Marshal(bannersAnswer)
				//Проверяем ответ
				assert.Equal(t, test.expectedBody, string(responseData))
				assert.Equal(t, test.expectedCode, w.Code)
			} else {

				//Проверяем ответ
				assert.Equal(t, test.expectedBody, string(responseData))
				assert.Equal(t, test.expectedCode, w.Code)
			}
		})

	}

}

// Тест рассчитан что в базу не внесли дополнительных изменений. Тест на создание баннера ждет id 5
func TestHandler_BannerPost(t *testing.T) {
	//Таблица со всеми тестам
	testTable := []struct {
		name              string
		requestBody       string
		requestHeaderName string
		requestHeaderData string
		expectedBody      string
		expectedCode      int
	}{
		//Тест #1
		{
			name: "Test #1. OK. корректный стандартный запрос",
			requestBody: `{
				"tag_ids": [1,2],
				"feature_id":1,
				"content": {
						"title": "test_title",
						"text": "test_text",
						"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{"banner_id":5}`,
			expectedCode:      http.StatusCreated,
		},
		//Тест #2
		{
			name: "Test #2.NOT OK. запрос в котором токен, которого нет в бд",
			requestBody: `{
				"tag_ids": [1,2],
				"feature_id":1,
				"content": {
						"title": "test_title",
						"text": "test_text",
						"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "something",
			expectedBody:      `{'error':' Пользователь не авторизован'}`,
			expectedCode:      http.StatusUnauthorized,
		},
		//Тест #3
		{
			name: "Test #3.NOT OK. запрос в котором tag_id, которого нет в БД",
			requestBody: `{
				"tag_ids": [9999999999999,2],
				"feature_id":1,
				"content": {
						"title": "test_title",
						"text": "test_text",
						"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Тест #4
		{
			name: "Test #4.NOT OK. запрос в котором feature_id, которого нет в БД",
			requestBody: `{
				"tag_ids": [1,2],
				"feature_id":9999999999,
				"content": {
						"title": "test_title",
						"text": "test_text",
						"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #5
		{
			name: "Test #5.NOT OK. запрос от обычного пользователя",
			requestBody: `{
				"tag_ids": [1,2],
				"feature_id":1,
				"content": {
						"title": "test_title",
						"text": "test_text",
						"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{'error':' Пользователь не имеет доступа'}`,
			expectedCode:      http.StatusForbidden,
		},
		//Test #6
		{
			name: "Test #6.NOT OK. запрос c полями которых не должно быть",
			requestBody: `{
						"something":"anything",
						"tag_ids": [1,2],
						"feature_id":1,
						"content": {
								"title": "test_title",
								"text": "test_text",
								"url": "test_url"
							},
						"is_active":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #7
		{
			name: "Test #7.NOT OK. запрос на создание баннера, который уже есть",
			requestBody: `{
				"tag_ids": [1,2],
				"feature_id":1,
				"content": {
					"title": "test_title",
					"text": "test_text",
					"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
		//Test #8
		{
			name: "Test #8.NOT OK. запрос на создание баннера, без тэгов",
			requestBody: `{
				"tag_ids": [],
				"feature_id":1,
				"content": {
					"title": "test_title",
					"text": "test_text",
					"url": "test_url"
					},
				"is_active":true
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
		//Test #9
		{
			name: "Test #9.NOT OK. запрос на создание баннера, где feature_id = 0 ",
			requestBody: `{
						"tag_ids": [1,2],
						"feature_id":0,
						"content": {
							"title": "test_title",
							"text": "test_text",
							"url": "test_url"
							},
						"is_active":true
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
	}

	//Поднимаем все сервисы
	handlers := initServices()
	router := handlers.InitRoutes()

	//В цикле проходимся по каждому тесту и запускаем его
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/banner", strings.NewReader(test.requestBody))

			req.Header.Add(test.requestHeaderName, test.requestHeaderData)
			w := httptest.NewRecorder()
			//Выполняется запрос
			router.ServeHTTP(w, req)
			//Получаем ответ
			responseData, _ := io.ReadAll(w.Body)

			//Проверяем ответ
			assert.Equal(t, test.expectedBody, string(responseData))
			assert.Equal(t, test.expectedCode, w.Code)
		})

	}

}

// Тесты проводяться на баннер с id 1
func TestHandler_BannerUpdate(t *testing.T) {
	//Таблица со всеми тестам
	testTable := []struct {
		name              string
		requestBody       string
		requestHeaderName string
		requestHeaderData string
		expectedBody      string
		expectedCode      int
	}{
		//Test #1
		{
			name: "Test #1. OK. корректный стандартный запрос",
			requestBody: `{
				"tag_ids": [],
				"feature_id":3,
				"content": {
						"title": "testChange_title",
						"text": "testChange_text",
						"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      ``,
			expectedCode:      http.StatusOK,
		},
		//Test #2
		{
			name: "Test #2.NOT OK. запрос где несуществующий tag_id",
			requestBody: `{
						"tag_ids": [99999999],
						"feature_id":3,
						"content": {
								"title": "testChange_title",
								"text": "testChange_text",
								"url": "testChange_url"
							},
						"is_active":false
					}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #3
		{
			name: "Test #3.NOT OK. запрос где несуществующий feature_id",
			requestBody: `{
							"tag_ids": [1],
							"feature_id":9999999,
							"content": {
									"title": "testChange_title",
									"text": "testChange_text",
									"url": "testChange_url"
								},
							"is_active":false
						}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #4
		{
			name: "Test #4.NOT OK. запрос где неверный tag_id ",
			requestBody: `{
				"tag_ids": [-1],
				"feature_id":1,
				"content": {
					"title": "testChange_title",
					"text": "testChange_text",
					"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Внутренняя ошибка сервера'}`,
			expectedCode:      http.StatusInternalServerError,
		},
		//Test #5
		{
			name: "Test #5.NOT OK. запрос где неверный feature_id ",
			requestBody: `{
				"tag_ids": [1],
				"feature_id":0,
				"content": {
					"title": "testChange_title",
					"text": "testChange_text",
					"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
		//Test #6
		{
			name: "Test #6.NOT OK. запрос от обычного пользователя ",
			requestBody: `{
				"tag_ids": [1],
				"feature_id":3,
				"content": {
					"title": "testChange_title",
					"text": "testChange_text",
					"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{'error':' Пользователь не имеет доступа'}`,
			expectedCode:      http.StatusForbidden,
		},
		//Test #7
		{
			name: "Test #7.NOT OK. запрос где несуществующий токен ",
			requestBody: `{
				"tag_ids": [1],
				"feature_id":0,
				"content": {
					"title": "testChange_title",
					"text": "testChange_text",
					"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "something",
			expectedBody:      `{'error':' Пользователь не авторизован'}`,
			expectedCode:      http.StatusUnauthorized,
		},
		//Test #8
		{
			name: "Test #8.NOT OK. запрос который меняет баннер на уже существующий ",
			requestBody: `{
				"tag_ids": [1,2,3],
				"feature_id":2,
				"content": {
					"title": "testChange_title",
					"text": "testChange_text",
					"url": "testChange_url"
					},
				"is_active":false
			}`,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
	}

	//Поднимаем все сервисы
	handlers := initServices()
	router := handlers.InitRoutes()

	//В цикле проходимся по каждому тесту и запускаем его
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest("PATCH", "/banner/{1}", strings.NewReader(test.requestBody))

			req.Header.Add(test.requestHeaderName, test.requestHeaderData)
			w := httptest.NewRecorder()
			//Выполняется запрос
			router.ServeHTTP(w, req)
			//Получаем ответ
			responseData, _ := io.ReadAll(w.Body)

			//Проверяем ответ
			assert.Equal(t, test.expectedBody, string(responseData))
			assert.Equal(t, test.expectedCode, w.Code)
		})

	}

}

// Изменения проводятся с баннер id 1
func TestHandler_BannerDelete(t *testing.T) {
	//Таблица со всеми тестам
	testTable := []struct {
		name              string
		requestBody       string
		requestHeaderName string
		requestHeaderData string
		expectedBody      string
		expectedCode      int
	}{
		//Test #1
		{
			name:              "Test #1. OK. корректный стандартный запрос",
			requestBody:       ``,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      ``,
			expectedCode:      http.StatusNoContent,
		},
		//Test #2
		{
			name:              "Test #2.NOT OK. запрос от обычного пользователя ",
			requestBody:       ``,
			requestHeaderName: "token",
			requestHeaderData: "user_UserFirst",
			expectedBody:      `{'error':' Пользователь не имеет доступа'}`,
			expectedCode:      http.StatusForbidden,
		},
		//Test #3
		{
			name:              "Test #3.NOT OK. запрос где несуществующий токен ",
			requestBody:       ``,
			requestHeaderName: "token",
			requestHeaderData: "something",
			expectedBody:      `{'error':' Пользователь не авторизован'}`,
			expectedCode:      http.StatusUnauthorized,
		},
		//Test #4
		{
			name:              "Test #4.NOT OK. запрос удаляет баннер которого нет",
			requestBody:       ``,
			requestHeaderName: "token",
			requestHeaderData: "admin_AdminFirst",
			expectedBody:      `{'error':' Неккорентные данные'}`,
			expectedCode:      http.StatusBadRequest,
		},
	}

	//Поднимаем все сервисы
	handlers := initServices()
	router := handlers.InitRoutes()

	//В цикле проходимся по каждому тесту и запускаем его
	for i, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			var req *http.Request
			//для 4 теста на несуществующий баннер
			if i == 3 {
				req, _ = http.NewRequest("DELETE", "/banner/{999999}", strings.NewReader(test.requestBody))
			} else {
				req, _ = http.NewRequest("DELETE", "/banner/{1}", strings.NewReader(test.requestBody))
			}

			req.Header.Add(test.requestHeaderName, test.requestHeaderData)
			w := httptest.NewRecorder()
			//Выполняется запрос
			router.ServeHTTP(w, req)
			//Получаем ответ
			responseData, _ := io.ReadAll(w.Body)

			//Проверяем ответ
			assert.Equal(t, test.expectedBody, string(responseData))
			assert.Equal(t, test.expectedCode, w.Code)
		})

	}

}

//Все тесты проходят правильно, если произошла ошибка
//Скорее всего нужно пересоздать DB с помощью init_down,init_up
