package data

import (
	"database/sql"
	"fmt"
	"user/pkg/constants"
)

const TBL_GITLAB_USER = "gitlab_user"

type GitlabUser struct {
	ID        int64
	UserID    int64
	GitlabID  int64
	UserName  string
	Name      string
	AvatarUrl string
	Email     string
	CreateAt  int64
}
type IGitlabUserData interface {
	AddUser(user *GitlabUser) error
	GetByGitlabID(gitlabID int64) (*GitlabUser, error)
	GetByUserID(userID int64) (*GitlabUser, error)
}

type gitlabUserData struct {
	table string
	db    *sql.DB
}

func (d *gitlabUserData) GetByGitlabID(gitlabID int64) (*GitlabUser, error) {
	sqlStr := fmt.Sprintf("select id,user_id,gitlab_id,username,`name`,avatar_url,email,create_at from %s where gitlab_id = ?", d.table)
	row := d.db.QueryRow(sqlStr, gitlabID)
	user := &GitlabUser{}
	err := row.Scan(&user.ID, &user.UserID, &user.GitlabID, &user.UserName, &user.Name, &user.AvatarUrl, &user.Email, &user.CreateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *gitlabUserData) GetByUserID(userID int64) (*GitlabUser, error) {
	sqlStr := fmt.Sprintf("select id,user_id,gitlab_id,username,`name`,avatar_url,email,create_at from %s where user_id = ?", d.table)
	row := d.db.QueryRow(sqlStr, userID)
	user := &GitlabUser{}
	err := row.Scan(&user.ID, &user.UserID, &user.GitlabID, &user.UserName, &user.Name, &user.AvatarUrl, &user.Email, &user.CreateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (d *gitlabUserData) AddUser(user *GitlabUser) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	userSql := fmt.Sprintf("insert into %s (`name`,avatar_url,`state`,create_at)values(?,?,?,?)", TBL_USER)
	res, err := tx.Exec(userSql, user.Name, user.AvatarUrl, constants.Active, user.CreateAt)
	if err != nil {
		tx.Rollback()
		return err
	}
	user.UserID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	gitlabUserSql := fmt.Sprintf("insert into %s (user_id,gitlab_id,username,`name`,avatar_url,email,create_at)values(?,?,?,?,?,?,?)", d.table)
	res, err = tx.Exec(gitlabUserSql, user.UserID, user.GitlabID, user.UserName, user.Name, user.AvatarUrl, user.Email, user.CreateAt)
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
