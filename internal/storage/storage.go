package storage

import (
	"context"
	"os"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
)

type StatStorage interface {
	CreateStat(ctx context.Context, stat *Stat) error
	GetStat(ctx context.Context, userName string, chatID int64) (*Stat, error)
	StatIsExists(ctx context.Context, userName string, chatID int64) (bool, error)
	UpdateStat(ctx context.Context, stat *Stat) error
	GetStats(ctx context.Context, chatID int64) ([]*Stat, error)
}

type DateStorage interface {
	CreateStartDate(ctx context.Context, date *Date) error
	GetDate(ctx context.Context, userName string, chatID int64) (*Date, error)
	DateIsExist(ctx context.Context, userName string, chatID int64) (bool, error)
	UpdateStopDate(ctx context.Context, date *Date) error
	DeleteDate(ctx context.Context, userName string, chatID int64) error
}

type DirStorage interface {
	CreateDir(ctx context.Context, dir *Dir) error
	GetDirByPath(ctx context.Context, path string) (*Dir, error)
	GetDirByUUID(ctx context.Context, uuid string) (*Dir, error)
	DirIsExistsByPath(ctx context.Context, path string) (bool, error)
	DirIsExistsByUUID(ctx context.Context, uuid string) (bool, error)
}

type IStorage interface {
	PrepareTables(ctx context.Context) error
	PrepareArchiveTable(ctx context.Context) error
}

type Storage struct {
	db *bun.DB
}

func NewStorage(db *bun.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (u *Storage) PrepareTables(ctx context.Context) error {

	_, err := u.db.NewCreateTable().
		Model((*Stat)(nil)).
		Table("stats").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = u.db.NewCreateTable().
		Model((*Date)(nil)).
		Table("dates").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = u.db.NewCreateTable().
		Model((*Dir)(nil)).
		Table("dirs").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) PrepareArchiveTable(ctx context.Context) error {

	visitedDirs := make(map[string]bool)
	var visitDir func(path string) error
	visitDir = func(path string) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		files, err := file.ReadDir(-1)
		if err != nil {
			return err
		}

		for _, f := range files {
			currentPath := path + "/" + f.Name()
			if !visitedDirs[currentPath] {
				if f.IsDir() {
					visitedDirs[currentPath] = true
					exists, err := u.DirIsExistsByPath(ctx, currentPath)
					if err != nil {
						return err
					}
					if !exists {
						dir := new(Dir)
						dir.Path = currentPath
						dir.ParentPath = path
						if err := u.CreateDir(ctx, dir); err != nil {
							return err
						}
					}

					visitDir(currentPath)
				}
			}
		}

		return nil
	}

	exists, err := u.DirIsExistsByPath(ctx, "archive")
	if err != nil {
		return err
	}
	if !exists {
		dir := new(Dir)
		dir.Path = "archive"
		dir.ParentPath = "archive"
		if err := u.CreateDir(ctx, dir); err != nil {
			return err
		}
	}

	if err := visitDir("archive"); err != nil {
		return err

	}

	return nil
}

func (u *Storage) CreateStat(ctx context.Context, stat *Stat) error {

	_, err := u.db.NewInsert().
		Model(stat).
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
		return stat, err
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

func (u *Storage) UpdateStat(ctx context.Context, stat *Stat) error {

	_, err := u.db.NewUpdate().
		Model(stat).
		Where("username = ? AND chat_id = ?", stat.Name, stat.ChatID).
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
		return stats, err
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
		return date, err
	}

	return date, nil
}

func (u *Storage) UpdateStopDate(ctx context.Context, date *Date) error {

	_, err := u.db.NewUpdate().
		Model(date).
		Where("username = ? AND chat_id = ?", date.Name, date.ChatID).
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

func (u *Storage) CreateDir(ctx context.Context, dir *Dir) error {

	_, err := u.db.NewInsert().
		Model(dir).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *Storage) GetDirByPath(ctx context.Context, path string) (*Dir, error) {

	dir := new(Dir)
	err := u.db.NewSelect().
		Model(dir).
		Where("path = ?", path).
		Scan(ctx)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

func (u *Storage) GetDirByUUID(ctx context.Context, uuid string) (*Dir, error) {

	dir := new(Dir)
	err := u.db.NewSelect().
		Model(dir).
		Where("uuid = ?", uuid).
		Scan(ctx)
	if err != nil {
		return dir, err
	}

	return dir, nil
}

func (u *Storage) DirIsExistsByPath(ctx context.Context, path string) (bool, error) {

	exists, err := u.db.NewSelect().
		Model((*Dir)(nil)).
		Where("path = ?", path).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (u *Storage) DirIsExistsByUUID(ctx context.Context, uuid string) (bool, error) {

	exists, err := u.db.NewSelect().
		Model((*Dir)(nil)).
		Where("uuid = ?", uuid).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}
