package database

import (
	"github.com/sergeychur/give_it_away/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
)

const (
	GetUserById = "SELECT * FROM users WHERE vk_id = $1"
	CreateUser  = "INSERT INTO users (vk_id, name, surname, photo_url) VALUES ($1, $2, $3, $4)"
)

func (db *DB) GetUser(userId int) (models.User, int) {
	row := db.db.QueryRow(GetUserById, userId)
	user := models.User{}
	err := row.Scan(&user.VkId, &user.Carma, &user.Name, &user.Surname, &user.PhotoUrl)
	if err == pgx.ErrNoRows {
		return user, EMPTY_RESULT
	}
	if err != nil {
		log.Println(err.Error())
		return user, DB_ERROR
	}
	return user, FOUND
}

func (db *DB) CreateUser(userId int, name string, surname string, photoURL string) int {
	_, err := db.db.Exec(CreateUser, userId, name, surname, photoURL)
	if err != nil {
		return DB_ERROR
	}
	return CREATED
}
