package storage

import (
	"context"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
)

type StatStorage interface {
	CreateStat(ctx context.Context, user *Stat) error
	GetStat(ctx context.Context, userName string, chatID int64) (*Stat, error)
	StatIsExists(ctx context.Context, userName string, chatID int64) (bool, error)
	UpdateStat(ctx context.Context, user *Stat) error
	GetStats(ctx context.Context, chatID int64) ([]*Stat, error)
}

type DateStorage interface {
	CreateStartDate(ctx context.Context, user *Date) error
	GetDate(ctx context.Context, userName string, chatID int64) (*Date, error)
	DateIsExist(ctx context.Context, userName string, chatID int64) (bool, error)
	UpdateStopDate(ctx context.Context, user *Date) error
	DeleteDate(ctx context.Context, userName string, chatID int64) error
}

type Storage struct {
	db *bun.DB
}

func NewStorage(db *bun.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func PrepareStorage(ctx context.Context, db *bun.DB) error {

	_, err := db.NewCreateTable().
		Model((*Stat)(nil)).
		Table("stats").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		Model((*Date)(nil)).
		Table("dates").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) CreateStat(ctx context.Context, user *Stat) error {

	_, err := u.db.NewInsert().
		Model(user).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) GetStat(ctx context.Context, userName string, chatID int64) (*Stat, error) {

	stat := new(Stat)
	err := u.db.NewSelect().
		Model(stat).
		Where("username = ? AND chat_id = ?", userName, chatID).
		Scan(ctx)
	if err != nil {
		return &Stat{}, err
	}

	return stat, nil
}

func (u *Storage) StatIsExists(ctx context.Context, userName string, chatID int64) (bool, error) {

	exists, err := u.db.NewSelect().
		Model((*Stat)(nil)).
		Where("username = ? AND chat_id = ?", userName, chatID).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u *Storage) UpdateStat(ctx context.Context, user *Stat) error {

	_, err := u.db.NewUpdate().
		Model(user).
		Where("username = ? AND chat_id = ?", user.Name, user.ChatID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) GetStats(ctx context.Context, chatID int64) ([]*Stat, error) {

	var stats []*Stat
	err := u.db.NewSelect().
		Model(&stats).
		Scan(ctx)
	if err != nil {
		return []*Stat{}, err
	}

	return stats, nil
}

func (u *Storage) CreateStartDate(ctx context.Context, date *Date) error {

	_, err := u.db.NewInsert().
		Model(date).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) DateIsExist(ctx context.Context, userName string, chatID int64) (bool, error) {

	exists, err := u.db.NewSelect().
		Model((*Date)(nil)).
		Where("username = ? AND chat_id = ?", userName, chatID).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u *Storage) GetDate(ctx context.Context, userName string, chatID int64) (*Date, error) {

	date := new(Date)
	_, err := u.db.NewSelect().
		Model(date).
		Where("username = ? AND chat_id = ?", userName, chatID).
		ScanAndCount(ctx)
	if err != nil {
		return &Date{}, err
	}

	return date, nil
}

func (u *Storage) UpdateStopDate(ctx context.Context, user *Date) error {

	_, err := u.db.NewUpdate().
		Model(user).
		Where("username = ? AND chat_id = ?", user.Name, user.ChatID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) DeleteDate(ctx context.Context, userName string, chatID int64) error {

	_, err := u.db.NewDelete().
		Model((*Date)(nil)).
		Where("username = ? AND chat_id = ?", userName, chatID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
