package AVITO

type User struct {
	Id      int    `json:"id" binding:"required"`
	Token   string `json:"token" binding:"required"`
	Tag_id  int    `json:"tag_id" binding:"required"`
	IsAdmin bool   `json:"is_admin" binding:"required"`
}
