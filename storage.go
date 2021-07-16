package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var (
	DSN = "host=localhost port=5432 user=postgres dbname=storage sslmode=disable"
)

type User struct {
	Name string `json:"username"`
	Counter int `json:"counter"`
	AverageTimeSleep float64 `json:"averagetimesleep"`
}

type Storage interface {
	Create(*User) error
	Get(userName string) (*User, error)
	Update(*User) error
	GetAll() ([]*User, error)
}

type UserStorage struct {}

func NewUserStorage() *UserStorage {
	return &UserStorage{}
}

func CreateConnection() (*sql.DB, error) {

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PrepareStorage(db *sql.DB) error {

	qs := []string{
		`DROP TABLE IF EXISTS time_sleep;`,
		`CREATE TABLE time_sleep(username VARCHAR(20), counter INTEGER, averagetimesleep NUMERIC(30, 2));`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UserStorage) Create(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO time_sleep(username, counter, averagetimesleep) VALUES($1, $2, $3)",
		user.Name, user.Counter, user.AverageTimeSleep,
		)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) Get(userName string) (*User, error) {

	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user User
	user.Name = userName

	row := db.QueryRow(
		"SELECT counter, averagetimesleep FROM time_sleep WHERE username=$1",
		userName,
		)
	row.Scan(&user.Counter, &user.AverageTimeSleep)

	return &user, nil
}

func (u *UserStorage) Update(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"UPDATE time_sleep SET counter=$1, averagetimesleep=$2 WHERE username=$3",
		user.Counter, user.AverageTimeSleep, user.Name,
		)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) GetAll() ([]*User, error) {

	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var users []*User

	rows, err := db.Query(
		"SELECT username, counter, averagetimesleep FROM time_sleep",
		)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		user := User{}

		err := rows.Scan(&user.Name, &user.Counter, &user.AverageTimeSleep)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}
	defer rows.Close()

	return users, nil
}