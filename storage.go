package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

type User struct {
	Name string `json:"username"`
	Group int64 `json:"group"`
	Start string `json:"start_date"`
	Stop string `json:"stop_date"`
	Counter int `json:"counter"`
	AverageTimeSleep float64 `json:"averagetimesleep"`
}

type Storage interface {
	CreateStat(*User) error
	GetStat(userName string, group int64) (*User, error)
	UpdateStat(*User) error
	GetStats(group int64) ([]*User, error)
	CreateStartDate(user *User) error
	GetDate(userName string) (*User, error)
	UpdateStopDate(user *User) error
	DeleteDate(userName string) error
}

type UserStorage struct {}

func NewUserStorage() *UserStorage {
	return &UserStorage{}
}

func CreateConnection() (*sql.DB, error) {

	DSN := fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/storage?sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func PrepareStorage(db *sql.DB) error {

	qs := []string{
		`DROP TABLE IF EXISTS stats;`,
		`CREATE TABLE stats(username VARCHAR(20), group_id INTEGER, counter INTEGER, averagetimesleep NUMERIC(30, 2));`,
		`DROP TABLE IF EXISTS dates;`,
		`CREATE TABLE dates(username VARCHAR(20), start_date VARCHAR(100), stop_date VARCHAR(100));`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UserStorage) CreateStat(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO stats(username, group_id, counter, averagetimesleep) VALUES($1, $2, $3, $4)",
		user.Name, user.Group, user.Counter, user.AverageTimeSleep,
		)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) GetStat(userName string, group int64) (*User, error) {

	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user User
	user.Name = userName
	user.Group = group

	row := db.QueryRow(
		"SELECT counter, averagetimesleep FROM stats WHERE username=$1 AND group_id=$2",
		userName, group,
		)
	row.Scan(&user.Counter, &user.AverageTimeSleep)

	return &user, nil
}

func (u *UserStorage) UpdateStat(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"UPDATE stats SET counter=$1, averagetimesleep=$2 WHERE username=$3 AND group_id=$4",
		user.Counter, user.AverageTimeSleep, user.Name, user.Group,
		)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) GetStats(group int64) ([]*User, error) {

	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var users []*User

	rows, err := db.Query(
		"SELECT username, counter, averagetimesleep FROM stats WHERE group_id=$1",
		group,
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

func (u *UserStorage) CreateStartDate(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO dates(username, start_date) VALUES($1, $2)",
		user.Name, user.Start,
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) GetDate(userName string) (*User, error) {

	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user User
	user.Name = userName

	row := db.QueryRow(
		"SELECT start_date, stop_date FROM dates WHERE username=$1",
		userName,
	)
	row.Scan(&user.Start, &user.Stop)

	return &user, nil
}

func (u *UserStorage) UpdateStopDate(user *User) error {

	db, err := CreateConnection()
	if err != nil {
		return nil
	}
	defer db.Close()

	_, err = db.Exec(
		"UPDATE dates SET stop_date=$1 WHERE username=$2",
		user.Stop, user.Name,
		)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStorage) DeleteDate(userName string) error {

	db, err := CreateConnection()
	if err != nil {
		return nil
	}
	defer db.Close()

	_, err = db.Exec(
		"DELETE FROM dates WHERE username=$1",
		userName,
		)
	if err != nil {
		return err
	}

	return nil
}
