package repository

import (
	AVITO "avito/elements"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type UsersPostgres struct {
	db *sqlx.DB
}

func NewUsersPostgres(db *sqlx.DB) *UsersPostgres {
	return &UsersPostgres{db: db}
}

func (u *UsersPostgres) CheckUserToken(token string) (user AVITO.User, errStatus int, err error) {

	//Запрос на получение пользователя с данным токеном
	query_ := fmt.Sprintf(`
		SELECT * FROM %s 
		WHERE token = '%s'`, usersTable, token)

	row := u.db.QueryRow(query_)
	err = row.Scan(&user.Id, &user.Token, &user.Tag_id, &user.IsAdmin)
	if err != nil {
		//Проверяем в чем именно ошибка и возвращаем
		//Соответствующий статус
		if err.Error() == "sql: no rows in result set" {
			return user, http.StatusUnauthorized, err

		} else {
			return user, http.StatusInternalServerError, err
		}

	}

	return user, 0, nil
}
