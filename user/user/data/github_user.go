package data

import (
	"database/sql"
	"fmt"
	"user/pkg/constants"
)

const TBL_GITHUB_USER = "github_user"

type GithubUser struct {
	ID        int64
	UserID    int64
	GithubID  int64
	UserName  string
	Name      string
	AvatarUrl string
	Email     string
	CreateAt  int64
}

type IGithubUserData interface {
	AddUser(user *GithubUser) error
	GetByGithubID(githubID int64) (*GithubUser, error)
	GetByUserID(userID int64) (*GithubUser, error)
}

type githubUserData struct {
	table string
	db    *sql.DB
}

func (d *githubUserData) GetByGithubID(githubID int64) (*GithubUser, error) {
	sqlStr := fmt.Sprintf("select id,user_id,github_id,username,`name`,avatar_url,email,create_at from %s where github_id = ?", d.table)
	row := d.db.QueryRow(sqlStr, githubID)

	user := &GithubUser{}
	err := row.Scan(
		&user.ID,
		&user.UserID,
		&user.GithubID,
		&user.UserName,
		&user.Name,
		&user.AvatarUrl,
		&user.Email,
		&user.CreateAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *githubUserData) GetByUserID(userID int64) (*GithubUser, error) {
	sqlStr := fmt.Sprintf("select id,user_id,github_id,username,`name`,avatar_url,email,create_at from %s where user_id = ?", d.table)
	row := d.db.QueryRow(sqlStr, userID)

	user := &GithubUser{}
	err := row.Scan(
		&user.ID,
		&user.UserID,
		&user.GithubID,
		&user.UserName,
		&user.Name,
		&user.AvatarUrl,
		&user.Email,
		&user.CreateAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *githubUserData) AddUser(user *GithubUser) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	userSql := fmt.Sprintf(
		"insert into %s (`name`,avatar_url,`state`,create_at)values(?,?,?,?)",
		TBL_USER,
	)

	res, err := tx.Exec(
		userSql,
		user.Name,
		user.AvatarUrl,
		constants.Active,
		user.CreateAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	user.UserID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	githubUserSql := fmt.Sprintf(
		"insert into %s (user_id,github_id,username,`name`,avatar_url,email,create_at)values(?,?,?,?,?,?,?)",
		d.table,
	)

	res, err = tx.Exec(
		githubUserSql,
		user.UserID,
		user.GithubID,
		user.UserName,
		user.Name,
		user.AvatarUrl,
		user.Email,
		user.CreateAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	user.ID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
