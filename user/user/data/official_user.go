package data

import (
	"database/sql"
	"fmt"
	"user/pkg/constants"
)

const TBL_OFFICIAL_USER = "official_user"

type OfficialUser struct {
	ID        int64
	UserID    int64
	OpenID    string
	NickName  string
	AvatarUrl string
	CreateAt  int64
}
type IOfficialUserData interface {
	AddUser(user *OfficialUser) error
	GetByOpenID(openID string) (*OfficialUser, error)
}

type officialUserData struct {
	table string
	db    *sql.DB
}

func (d *officialUserData) GetByOpenID(openID string) (*OfficialUser, error) {
	sqlStr := fmt.Sprintf("select id,user_id,open_id,`nickname`,avatar_url,create_at from %s where open_id = ?", d.table)
	row := d.db.QueryRow(sqlStr, openID)
	user := &OfficialUser{}
	err := row.Scan(&user.ID, &user.UserID, &user.OpenID, &user.NickName, &user.AvatarUrl, &user.CreateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *officialUserData) AddUser(user *OfficialUser) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	userSql := fmt.Sprintf("insert into %s (`name`,avatar_url,`state`,create_at)values(?,?,?,?)", TBL_USER)
	res, err := tx.Exec(userSql, user.NickName, user.AvatarUrl, constants.Active, user.CreateAt)
	if err != nil {
		tx.Rollback()
		return err
	}
	user.UserID, err = res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	officialUserSql := fmt.Sprintf("insert into %s (user_id,open_id,nickname,avatar_url,create_at)values(?,?,?,?,?)", d.table)
	res, err = tx.Exec(officialUserSql, user.UserID, user.OpenID, user.NickName, user.AvatarUrl, user.CreateAt)
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
