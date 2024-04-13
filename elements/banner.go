package AVITO

type Banner struct {
	Id         int    `json:"id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Text       string `json:"text" binding:"required"`
	Url        string `json:"url" binding:"required"`
	Created_at string `json:"created_at" binding:"required"`
	Updated_at string `json:"updated_at" binding:"required"`
	Is_active  bool   `json:"is_active" binding:"required"`
}
