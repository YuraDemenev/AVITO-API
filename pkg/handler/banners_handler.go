package handler

import (
	AVITO "avito/elements"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Структура для получение данных,
// Когда пользователь запрашивает баннер
type userBannerRequest struct {
	Tag_id          int  `json:"tag_id"`
	Feature_id      int  `json:"feature_id"`
	UseLastRevision bool `json:"use_last_revision"`
	Limit           int  `json:"limit"`
	Offset          int  `json:"offset"`
}

// Структура для возвращения баннера
type bannerAnswer struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Url   string `json:"url"`
}

// Структура для запросов на создание
// И изменение баннера
type BannerCreate struct {
	Tag_ids    []int        `json:"tag_ids"`
	Feature_id int          `json:"feature_id"`
	Content    bannerAnswer `json:"content"`
	Is_active  bool         `json:"is_active"`
}

// Чтобы не делать большие конструкции с if или case
// Чтобы просто и быстро получать строку ответа
var DictionaryErrors = map[int]string{
	400: "{'error':' Неккорентные данные'}",
	401: "{'error':' Пользователь не авторизован'}",
	403: "{'error':' Пользователь не имеет доступа'}",
	404: "{'error':' Баннер для не найден'}",
	500: "{'error':' Внутренняя ошибка сервера'}",
}

// Функция для получения баннера пользователем
func (h *Handler) userGetBanner(c *gin.Context) {
	//Структура для получения данных
	var requestData userBannerRequest

	//Создаем декодер
	decoder := json.NewDecoder(c.Request.Body)
	//Запрещаем иметь в запросе поля которых нет в структуре
	decoder.DisallowUnknownFields()

	//Декодируем
	err := decoder.Decode(&requestData)
	if err != nil {
		logrus.Errorf("Ошибка при декодирования request body. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(400)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusBadRequest]))
		return
	}

	//Получаем токен
	userToken := c.Request.Header.Get("token")

	//Проверяем токен пользователя
	IsAdmin, errStatus, err := h.service.CheckUserToken(userToken)
	if err != nil {
		//если ошибка
		logrus.Errorf("Ошибка при проверки токена пользователя при получении банера. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return

	}

	//Идем в базу данных за баннером
	title, text, url, errStatus, err := h.service.GetBanner(requestData.Tag_id, requestData.Feature_id,
		requestData.UseLastRevision, IsAdmin)
	if err != nil {
		logrus.Errorf("Ошибка при запросе на получение баннера. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return

	}

	//Создаём баннер для возвращения
	var banner bannerAnswer
	banner.Title = title
	banner.Text = text
	banner.Url = url

	//Конвертируем в json
	result, err := json.Marshal(banner)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{"error":" Внутренняя ошибка сервера"}`))
	}

	c.Writer.Write(result)
}

//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------

func (h *Handler) adminGetBanners(c *gin.Context) {
	//Структура для получения данных
	var requestData userBannerRequest
	//Создаем декодер
	decoder := json.NewDecoder(c.Request.Body)
	//Запрещаем иметь в запросе поля которых нет в структуре
	decoder.DisallowUnknownFields()

	//Декодируем
	err := decoder.Decode(&requestData)
	if err != nil {
		logrus.Errorf("Ошибка при декодирования request body. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusBadRequest]))
		return
	}

	//Получаем токен
	userToken := c.Request.Header.Get("token")

	//Проверяем админ ли пользователь
	checkAdmin := h.CheckAdmin(c, userToken)
	if !checkAdmin {
		return
	}

	//Идем в базу данных за баннерами и тегами
	banners, tags, errStatus, err := h.service.AdminGetBanners(requestData.Tag_id, requestData.Feature_id,
		requestData.Limit, requestData.Offset)
	if err != nil {
		logrus.Errorf("Ошибка при запросе на получение баннера. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return
	}

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

	//Создаём массив для возвращения
	answerSlice := make([]localBannerAnswer, 0)

	//В цикле проходим по всем баннерам и заполняем структуру ответа
	for i, banner := range banners {
		var answer localBannerAnswer
		answer.Banner_id = banner.Id
		answer.Tag_ids = tags[i]
		answer.Feature_id = requestData.Feature_id
		//Структура баннера
		var bannerAnswer bannerAnswer
		bannerAnswer.Text = banner.Text
		bannerAnswer.Title = banner.Title
		bannerAnswer.Url = banner.Url
		//Добавляем баннер в ответ
		answer.Content = bannerAnswer
		answer.Is_active = banner.Is_active
		answer.Created_at = banner.Created_at
		answer.Updated_at = banner.Updated_at

		//Добавляем в массив
		answerSlice = append(answerSlice, answer)
	}

	//Конвертируем в json
	result, err := json.Marshal(answerSlice)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{"error":" Внутренняя ошибка сервера"}`))
	}

	c.Writer.Write(result)
}

//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------

// Создание баннера
func (h *Handler) adminCreateBanner(c *gin.Context) {
	var requestData BannerCreate
	//Создаем декодер
	decoder := json.NewDecoder(c.Request.Body)
	//Запрещаем иметь в запросе поля которых нет в структуре
	decoder.DisallowUnknownFields()

	//Декодируем
	err := decoder.Decode(&requestData)
	if err != nil {
		logrus.Errorf("Ошибка при декодирования request body. Ошибка: %s", err.Error())
		//Здесь должен быть 400 но его нет в api
		c.Writer.WriteHeader(500)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusInternalServerError]))
		return
	}

	//Получаем токен
	userToken := c.Request.Header.Get("token")

	//Проверяем админ ли пользователь
	checkAdmin := h.CheckAdmin(c, userToken)
	if !checkAdmin {
		return
	}

	//Заполняем данными баннер для его передачи
	var banner AVITO.Banner
	banner.Text = requestData.Content.Text
	banner.Title = requestData.Content.Title
	banner.Url = requestData.Content.Url

	banner_id, errStatus, err := h.service.CreateNewBanner(requestData.Tag_ids, requestData.Feature_id,
		banner, requestData.Is_active)
	if err != nil {
		logrus.Errorf("Ошибка при создании банера. Ошибка: %s", err)
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return
	}

	//Структура для ответа
	type localAnswer struct {
		Banner_id int `json:"banner_id"`
	}

	//Передаем значение banner_id
	var answer localAnswer
	answer.Banner_id = banner_id

	//Конвертируем в json
	result, err := json.Marshal(answer)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		c.Writer.Write([]byte(`{"error":" Внутренняя ошибка сервера"}`))
	}

	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write(result)
}

//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------

func (h *Handler) adminUpdateBanner(c *gin.Context) {
	//Получение id из path
	idString := c.Param("{id}")
	//idString = {1} нужно преобразовать в int
	idString = idString[1 : len(idString)-1]
	id, err := strconv.Atoi(idString)
	if err != nil {
		logrus.Errorf("Ошибка при получении id. Ошибка: %s", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusBadRequest]))
		return
	}

	//Структура для получения данных
	var requestData BannerCreate

	//Создаем декодер
	decoder := json.NewDecoder(c.Request.Body)
	//Запрещаем иметь в запросе поля которых нет в структуре
	decoder.DisallowUnknownFields()

	//Декодируем
	err = decoder.Decode(&requestData)
	if err != nil {
		logrus.Errorf("Ошибка при декодирования request body. Ошибка: %s", err.Error())
		//Здесь должен быть 400 но его нет в api
		c.Writer.WriteHeader(500)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusInternalServerError]))
		return
	}

	//Получаем токен
	userToken := c.Request.Header.Get("token")

	//Проверяем админ ли пользователь
	checkAdmin := h.CheckAdmin(c, userToken)
	if !checkAdmin {
		return
	}

	//Для передачи в функцию ChangeBanner
	var banner AVITO.Banner
	banner.Text = requestData.Content.Text
	banner.Title = requestData.Content.Title
	banner.Url = requestData.Content.Url

	errStatus, err := h.service.ChangeBanner(id, requestData.Tag_ids, requestData.Feature_id, banner)
	if err != nil {
		logrus.Errorf("Ошибка при изменении баннера. Ошибка: %s", err)
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return
	}

}

//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------

func (h *Handler) adminDeleteBanner(c *gin.Context) {
	//Получение id из path
	idString := c.Param("{id}")
	//idString = {1} нужно преобразовать в int
	idString = idString[1 : len(idString)-1]
	id, err := strconv.Atoi(idString)
	if err != nil {
		logrus.Errorf("Ошибка при получении id. Ошибка: %s", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusBadRequest]))
		return
	}

	//Получаем токен
	userToken := c.Request.Header.Get("token")

	//Проверяем админ ли пользователь
	checkAdmin := h.CheckAdmin(c, userToken)
	if !checkAdmin {
		return
	}

	errStatus, err := h.service.DeleteBanner(id)
	if err != nil {

		logrus.Errorf("Ошибка при проверки токена пользователя при получении банера. Ошибка: %s", errors.New("is not admin"))
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)

}

func (h *Handler) CheckAdmin(c *gin.Context, userToken string) bool {
	//Проверяем токен пользователя
	IsAdmin, errStatus, err := h.service.CheckUserToken(userToken)
	if err != nil {
		//если ошибка
		logrus.Errorf("Ошибка при проверки токена пользователя при получении банера. Ошибка: %s", err.Error())
		c.Writer.WriteHeader(errStatus)
		c.Writer.Write([]byte(DictionaryErrors[errStatus]))
		return false

	} else if !IsAdmin {
		//Если пользователь не админ
		logrus.Errorf("Ошибка при проверки токена пользователя при получении банера. Ошибка: %s", errors.New("is not admin"))
		c.Writer.WriteHeader(http.StatusForbidden)
		c.Writer.Write([]byte(DictionaryErrors[http.StatusForbidden]))
		return false
	}
	return true
}
