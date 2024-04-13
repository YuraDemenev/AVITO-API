package repository

import (
	AVITO "avito/elements"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

type BannersPostgres struct {
	db *sqlx.DB
}

func NewBannersPostgres(db *sqlx.DB) *BannersPostgres {
	return &BannersPostgres{db: db}
}

// Функция /user_banner
func (b *BannersPostgres) GetBanner(tag int, feature int, isAdmin bool) (banner AVITO.Banner, errStatus int, err error) {
	//Проверка на правильность данных
	if feature <= 0 || tag <= 0 {
		return banner, http.StatusBadRequest, errors.New("not correct data tag <= or feature <= 0")
	}

	query_ := fmt.Sprintf(`
	SELECT DISTINCT b.id, b.title, b.text,b.url, b.created_at, b.updated_at, b.is_active
	FROM %s AS b
	JOIN %s AS tags
	ON b.id = tags.banner_id 
	AND tags.feature_id = %d
	AND tags.tag_id = %d
	`, bannersTable, bannersTagsTable, feature, tag)

	//Если пользователь не админ нужно добавить условие, чтобы
	//Он получил banner с is_active = true
	if !isAdmin {
		query_ += "\n AND b.is_active = true"
	}

	row := b.db.QueryRow(query_)
	err = row.Scan(&banner.Id, &banner.Title, &banner.Text, &banner.Url, &banner.Created_at, &banner.Updated_at, &banner.Is_active)
	if err != nil {
		return banner, http.StatusNotFound, err
	}

	return banner, http.StatusOK, nil
}

//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------

// Функция на /banner
func (b *BannersPostgres) AdminGetBanners(tag int, feature int, limit int, offset int) (banners []AVITO.Banner, tags [][]int, errStatus int, err error) {
	//Проверка данных
	if limit <= 0 || offset < 0 || (feature <= 0 && tag <= 0) {
		//Здесь должна быть 400 ошибка но её нет в api
		return nil, nil, http.StatusInternalServerError, errors.New("not correct data  tag <= 0 or feature <= 0 or offset < 0 limit <= 0")
	}

	banners = make([]AVITO.Banner, 0)
	//Создание запроса
	//Основная часть запроса
	query_ := fmt.Sprintf(`
	SELECT DISTINCT b.id,b.title,b.text,b.url,b.created_at,b.updated_at,b.is_active
	FROM %s as b
	JOIN %s AS tags
	ON b.id = tags.banner_id
	`, bannersTable, bannersTagsTable)

	//Если есть feature то добавляем
	if feature != 0 {
		query_ += fmt.Sprintf("AND tags.feature_id = %d \n", feature)
	}
	//Если есть tag_id то добавляем
	if tag != 0 {
		query_ += fmt.Sprintf("AND tags.tag_id = %d \n", tag)
	}

	//Добавляем offset и limit
	query_ += fmt.Sprintf(`ORDER BY b.id ASC 
	LIMIT %d OFFSET %d`, limit, offset)

	rows, err := b.db.Query(query_)
	if err != nil {
		return banners, nil, http.StatusInternalServerError, err
	}

	//Проходимся по результату и получаем все баннеры
	for rows.Next() {
		//Get all banners
		var banner AVITO.Banner
		if err := rows.Scan(&banner.Id, &banner.Title, &banner.Text, &banner.Url,
			&banner.Created_at, &banner.Updated_at, &banner.Is_active); err != nil {
			return banners, nil, http.StatusInternalServerError, err
		}
		banners = append(banners, banner)
	}

	//Проверка если если offset > кол-ва полученных значений то нужно вернуться
	if len(banners) == 0 {
		return nil, nil, 0, nil
	}

	//Для того чтобы создать строку типа (1,2)
	var builder strings.Builder

	builder.WriteString("(")
	builder.WriteString(strconv.Itoa(banners[0].Id))
	for i := 1; i < len(banners); i++ {

		builder.WriteString(",")
		builder.WriteString(strconv.Itoa(banners[i].Id))
	}
	builder.WriteString(")")

	//Чтобы получить все тэги относящиеся к баннеру нужнен ещё один запрос
	query_ = fmt.Sprintf(`
	SELECT tags.banner_id,tags.tag_id 
	FROM %s as tags
	WHERE tags.banner_id in %s
	ORDER BY tags.banner_id ASC
	`, bannersTagsTable, builder.String())

	rows, err = b.db.Query(query_)
	if err != nil {
		return banners, nil, http.StatusInternalServerError, err
	}

	//Для отслеживания баннера
	lastId := banners[0].Id
	//Для создание массива, который хранит tags
	locSlice := make([]int, 0)
	//Для сохранения всех массивов с тэгами
	tagsSlice := make([][]int, len(banners))
	//У tagsSlice уже есть len поэтому нужно обращаться
	//К его индексам
	tagIndex := 0

	for rows.Next() {
		var tag int
		var bannerId int
		if err := rows.Scan(&bannerId, &tag); err != nil {
			return banners, nil, http.StatusInternalServerError, err
		}

		//Если нынешнее banner_id != lastId
		//Значит получен полный набор тегов для баннера
		if bannerId != lastId {
			lastId = bannerId
			//DeepCopy чтобы значения не сломались
			copySlice := make([]int, len(locSlice))
			copy(copySlice, locSlice)
			tagsSlice[tagIndex] = copySlice

			//Обнуляем locSlice
			locSlice = make([]int, 0)
			//Увеличиваем индекс
			tagIndex++

		}
		locSlice = append(locSlice, tag)

	}
	//Сохраняем последнее значение (последний набор tags не зашел в if bannerId != lastId )
	copySlice := make([]int, len(locSlice))
	copy(copySlice, locSlice)
	tagsSlice[tagIndex] = copySlice

	return banners, tagsSlice, 0, nil
}

//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------

func (b *BannersPostgres) CreateNewBanner(tags []int, feature int, content AVITO.Banner, is_active bool) (banner AVITO.Banner, errStatus int, err error) {
	//Проверка данных
	if len(tags) == 0 || feature <= 0 {
		return banner, http.StatusBadRequest, errors.New("массив tags пуст или feature <= 0")
	}

	//Для того чтобы создать строку типа (1,2)
	var builder strings.Builder

	builder.WriteString("(")
	builder.WriteString(strconv.Itoa(tags[0]))
	for i := 1; i < len(tags); i++ {
		if tags[i] <= 0 {
			return banner, http.StatusBadRequest, errors.New("в массиве tags есть значение <= 0")
		}
		builder.WriteString(",")
		builder.WriteString(strconv.Itoa(tags[i]))
	}
	builder.WriteString(")")

	//Баннер однозначно определяется тэгами и фичей поэтому нужно
	//проверить существует ли такой баннер
	query_ := fmt.Sprintf(`
	SELECT tags.banner_id
	FROM %s AS tags
	WHERE tags.feature_id = %d
    AND tags.tag_id in %s
    AND NOT EXISTS (
   		SELECT 1
    	FROM %s as tags1
    	WHERE tags.banner_id = tags1.banner_id
        AND tags1.tag_id not in %s
        AND tags1.feature_id = %d
    )
	GROUP BY tags.banner_id
	HAVING COUNT(DISTINCT tags.tag_id) = %d;
	`, bannersTagsTable, feature, builder.String(), bannersTagsTable, builder.String(), feature, len(tags))

	var banner_id int

	//В запросе ждем получить ошибку 'no rows', это значит такого
	//Банера нет и можно его создавать
	row := b.db.QueryRow(query_)
	err = row.Scan(&banner_id)
	if err != nil && err.Error() != "sql: no rows in result set" {
		//Если произошла непредвиденная ошибка
		return banner, http.StatusInternalServerError, err

	} else if err == nil {
		//Если нет ошибки, то такой баннер есть
		errorString := fmt.Sprintf("такой баннер уже есть. id: %d", banner_id)
		return banner, http.StatusBadRequest, errors.New(errorString)
	}

	//запрос на создание баннера
	query_ = fmt.Sprintf(`
	INSERT INTO %s (title,text,url,created_at,updated_at,is_active)
	VALUES ('%s','%s','%s',NOW(),NOW(),%t)
	RETURNING id, title, text, url, created_at, updated_at, is_active;
	`, bannersTable, content.Title, content.Text, content.Url, is_active)

	//Создаем баннер
	err = b.db.QueryRow(query_).Scan(&banner.Id, &banner.Title, &banner.Text,
		&banner.Url, &banner.Created_at, &banner.Updated_at, &banner.Is_active)
	if err != nil {
		return banner, http.StatusInternalServerError, err
	}

	//Создаём связь с тэгами, так как тэгов может быть несколько,
	//Нужно сделать добавление сразу всех
	builder.Reset()

	for i := 0; i < len(tags); i++ {
		builder.WriteString("(")
		//Добавляем banner_id
		builder.WriteString(strconv.Itoa(banner.Id) + ",")
		//Добавляем feature_id
		builder.WriteString(strconv.Itoa(feature) + ",")
		//Добавляем tag_id
		builder.WriteString(strconv.Itoa(tags[i]) + ")")
		//Последние скобки не должны иметь ','
		if i+1 < len(tags) {
			builder.WriteString(",")
		}

	}

	query_ = fmt.Sprintf(`
	INSERT INTO %s(banner_id, feature_id, tag_id)
	VALUES %s
	`, bannersTagsTable, builder.String())

	//Создаем связь
	err = b.db.QueryRow(query_).Err()
	if err != nil {
		return banner, http.StatusInternalServerError, err
	}

	return banner, 0, nil
}

//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------

func (b *BannersPostgres) ChangeBanner(banner_id int, tags []int, feature int, content AVITO.Banner) (banner AVITO.Banner, curTags []int, curFeaturId int, errStatus int, err error) {
	//Проверка данных если все данные нулевые тогда ошибка
	if len(tags) == 0 && feature == 0 && content.Text == "" && content.Title == "" && content.Url == "" {
		return banner, nil, 0, http.StatusBadRequest, errors.New("все данные пустые")
	}

	//Для сохранения тэгов
	getTagsSlice := make([]int, 0)

	//Для получения данных и работы с ними
	var getId int
	var getFeature_id int
	var getTitle string
	var getText string
	var getUrl string

	//Сначала получаем уже имеющуюся информацию об этом баннере
	//Получаем title, text, url, feature_id, tag_ids
	query_ := fmt.Sprintf(`
		SELECT b.title,b.text,b.url, tags.feature_id,tags.tag_id
		FROM %s AS b
		JOIN %s AS tags 
		ON b.id = tags.banner_id
		AND tags.banner_id = %d
		`, bannersTable, bannersTagsTable, banner_id)

	rows, err := b.db.Query(query_)
	if err != nil {
		return banner, nil, 0, http.StatusInternalServerError, err
	}

	for rows.Next() {
		//Для сохранения в массив
		var locTag int

		if err := rows.Scan(&getTitle, &getText, &getUrl, &getFeature_id, &locTag); err != nil {
			return banner, nil, 0, http.StatusInternalServerError, err
		}
		//Если у нас нет тэгов то сохраняем полученные
		if len(tags) == 0 {
			getTagsSlice = append(getTagsSlice, locTag)
		}
	}
	//чтобы не путаться при создании запросов на изменение
	//данные которые нужны для изменений будут в переменных
	//начинающихся с get
	if feature != 0 {
		getFeature_id = feature
	}
	//Также с остальными переменными
	if content.Text != "" {
		getText = content.Text
	}
	if content.Title != "" {
		getTitle = content.Title
	}
	if content.Url != "" {
		getUrl = content.Url
	}
	//Заполняем массив тэгов
	if len(tags) != 0 {
		getTagsSlice = append(getTagsSlice, tags...)

	}

	//Нужно проверить можем ли мы изменить данные в баннере
	//для этого нужно узнать нет ли уже такого баннера

	//Для того чтобы создать строку типа (1,2)
	var builder strings.Builder

	builder.WriteString("(")
	builder.WriteString(strconv.Itoa(getTagsSlice[0]))
	for i := 1; i < len(getTagsSlice); i++ {
		if getTagsSlice[i] <= 0 {
			return banner, nil, 0, http.StatusBadRequest, errors.New("в массиве tags есть значение <= 0")
		}
		builder.WriteString(",")
		builder.WriteString(strconv.Itoa(getTagsSlice[i]))
	}
	builder.WriteString(")")

	//Баннер однозначно определяется тэгами и фичей поэтому нужно
	//проверить существует ли такой баннер
	query_ = fmt.Sprintf(`
		SELECT tags.banner_id
		FROM %s AS tags
		WHERE tags.feature_id = %d
		AND tags.tag_id in %s
		AND NOT EXISTS (
			   SELECT 1
			FROM %s as tags1
			WHERE tags.banner_id = tags1.banner_id
			AND tags1.tag_id not in %s
			AND tags1.feature_id = %d
		)
		GROUP BY tags.banner_id
		HAVING COUNT(DISTINCT tags.tag_id) = %d;
		`, bannersTagsTable, getFeature_id, builder.String(), bannersTagsTable, builder.String(), getFeature_id, len(getTagsSlice))

	row := b.db.DB.QueryRow(query_)

	//пытаемся получить id
	//Здесь  нужна ошибка об отсутсвии строк, потому что это
	//Значит, что такого баннера нет
	err = row.Scan(&getId)
	if err != nil && err.Error() != "sql: no rows in result set" {
		//Если произошла непредвиденная ошибка
		return banner, nil, 0, http.StatusInternalServerError, err

	} else if err == nil {
		//Если нет ошибки, то такой баннер есть
		errorString := fmt.Sprintf("такой баннер уже есть. id: %d", banner_id)
		return banner, nil, 0, http.StatusBadRequest, errors.New(errorString)
	}

	//Нужно одновременно поменять таблицу banners и таблицу banners_tags. Но banners_tags меняется не всегда
	//Поэтому 2 query, которые потом при необходимости собируться в один.
	query1 := ""
	query2 := ""

	//Так как мы будем собирать 2 query эффективно это сделать с помощью многопоточности
	var wg sync.WaitGroup

	//Если мы меняем tags или/и feature то мы сначало удаляем старые, затем добавляем новые
	if len(tags) != 0 || feature != 0 {
		wg.Add(1)

		go func(query1 *string) {
			defer wg.Done()

			//Builder для создания INSERT VALUES
			builder.Reset()

			//Создаём связь с тэгами, так как тэгов может быть несколько,
			//Нужно сделать добавление сразу всех

			for i := 0; i < len(getTagsSlice); i++ {
				builder.WriteString("(")
				//Добавляем banner_id
				builder.WriteString(strconv.Itoa(banner_id) + ",")
				//Добавляем feature_id
				builder.WriteString(strconv.Itoa(getFeature_id) + ",")
				//Добавляем tag_id
				builder.WriteString(strconv.Itoa(getTagsSlice[i]) + ")")
				//Последние скобки не должны иметь ','
				if i+1 < len(getTagsSlice) {
					builder.WriteString(",")
				}

			}

			*query1 = fmt.Sprintf(`
			DELETE FROM %s AS tags
			WHERE tags.banner_id = %d;
			INSERT INTO %s 
			VALUES %s;
			`, bannersTagsTable, banner_id, bannersTagsTable, builder.String())

		}(&query1)
	}
	wg.Add(1)

	go func(query2 *string) {
		defer wg.Done()
		//Изменяем данные в таблице banners
		*query2 = fmt.Sprintf(`
		UPDATE %s 
		SET title = '%s', 
		text = '%s', 
		url = '%s',
		is_active = %t,
		updated_at = NOW() 
		WHERE id = %d 
		RETURNING id, title, text, url, created_at, updated_at, is_active
		`, bannersTable, getTitle, getText, getUrl, content.Is_active, banner_id)
	}(&query2)

	wg.Wait()

	query_ = ""
	//Если не надо обновлять banners_tags
	if query1 == "" {
		query_ = query2

	} else {
		//Если надо обновить обе таблицы
		query_ = query1 + query2
	}

	//Делаем изменение в banners_tags
	err = b.db.DB.QueryRow(query_).Scan(&banner.Id, &banner.Title, &banner.Text, &banner.Url,
		&banner.Created_at, &banner.Updated_at, &banner.Is_active)
	if err != nil {
		return banner, nil, 0, http.StatusInternalServerError, err
	}

	return banner, getTagsSlice, getFeature_id, 0, nil
}

//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------------------------------------------

func (b *BannersPostgres) DeleteBanner(banner_id int) (errStatus int, err error) {
	//Проверка данных. id не может быть <= 0
	if banner_id <= 0 {
		return http.StatusBadRequest, errors.New("banner id <= 0")
	}

	//В базе связь сделана с DELETE ON CASCADE поэтому
	//Просто удаляем баннер по id
	query_ := fmt.Sprintf(`
	WITH deleted AS 
	(DELETE FROM %s WHERE id = %d RETURNING *) 
	SELECT count(*) FROM deleted
	`, bannersTable, banner_id)

	//Для того чтобы проверить сколько было удаленно баннеров
	var count int

	err = b.db.DB.QueryRow(query_).Scan(&count)
	//Ошибка на стороне сервера
	if err != nil {
		return http.StatusInternalServerError, err

	} else if count == 0 {
		return http.StatusBadRequest, errors.New("0 баннеров было удаленно, данные неккоректны")
	}

	return 0, nil
}
