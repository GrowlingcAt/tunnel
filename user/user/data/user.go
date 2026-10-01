package data

import (
	"database/sql"
	"fmt"
	"strings"
	"user/pkg/constants"
)

type User struct {
	ID        int64
	Phone     string
	UserName  string
	Name      string
	AvatarUrl string
	Email     string
	Pwd       string
	State     constants.UserState
	CreateAt  int64
}
type IUserData interface {
	AddUser(user *User) error
	GetByID(id int64) (*User, error)
	GetByPhone(phone string) (*User, error)
	UpdateUserInfo(user *User) error
}

const TBL_USER = "user"

type userData struct {
	table string
	db    *sql.DB
}

func (d *userData) GetByID(id int64) (*User, error) {
	sqlStr := fmt.Sprintf("select id,username,phone,`name`,avatar_url,email,`state`,create_at from %s where id = ?", d.table)
	row := d.db.QueryRow(sqlStr, id)
	user := &User{}
	var phone, email, username sql.NullString
	err := row.Scan(&user.ID, &username, &phone, &user.Name, &user.AvatarUrl, &email, &user.State, &user.CreateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if email.Valid {
		user.Email = email.String
	}
	if username.Valid {
		user.UserName = username.String
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userData) GetByPhone(phone string) (*User, error) {
	sqlStr := fmt.Sprintf("select id,username,phone,`name`,avatar_url,email,`state`,create_at from %s where phone = ?", d.table)
	row := d.db.QueryRow(sqlStr, phone)
	user := &User{}
	var email, username sql.NullString
	err := row.Scan(&user.ID, &username, &user.Phone, &user.Name, &user.AvatarUrl, &email, &user.State, &user.CreateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if email.Valid {
		user.Email = email.String
	}
	if username.Valid {
		user.UserName = username.String
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *userData) AddUser(user *User) error {
	var res sql.Result
	var err error
	var fields = "`name`,avatar_url,pwd,`state`,create_at"
	var values = fmt.Sprintf("'%s','%s','%s',%d,%d", user.Name, user.AvatarUrl, user.Pwd, user.State, user.CreateAt)
	if user.UserName != "" {
		fields += ",username"
		values += fmt.Sprintf(",'%s'", user.UserName)
	}
	if user.Email != "" {
		fields += ",email"
		values += fmt.Sprintf(",'%s'", user.Email)
	}
	if user.Phone != "" {
		fields += ",phone"
		values += fmt.Sprintf(",'%s'", user.Phone)
	}
	res, err = d.db.Exec(fmt.Sprintf("insert into %s (%s)values(%s)", d.table, fields, values))
	if err != nil {
		return err
	}
	user.ID, _ = res.LastInsertId()
	return nil
}
func (d *userData) UpdateUserInfo(user *User) error {
	if user.ID == 0 {
		return nil
	}
	updateFields := ""
	args := make([]interface{}, 0)
	if user.Name != "" {
		updateFields += "name = ?,"
		args = append(args, user.Name)
	}
	if user.AvatarUrl != "" {
		updateFields += "avatar_url = ?,"
		args = append(args, user.AvatarUrl)
	}

	if len(args) == 0 {
		return nil
	}
	sqlStr := fmt.Sprintf("update %s set %s where id = ?", d.table, strings.Trim(updateFields, ","))
	args = append(args, user.ID)
	_, err := d.db.Exec(sqlStr, args...)
	return err
}
