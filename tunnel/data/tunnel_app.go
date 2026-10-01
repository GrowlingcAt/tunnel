package data

import (
	"database/sql"
	"encoding/json"
)

type TunnelApp struct {
	ID            int64    `json:"id"`
	UserID        int64    `json:"user_id"`
	ServerID      int64    `json:"server_id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	ProxyType     string   `json:"proxy_type"`
	LocalIP       string   `json:"local_ip"`
	LocalPort     int      `json:"local_port"`
	CustomDomains []string `json:"custom_domains"`
	EntryDomain   string   `json:"entry_domain"`
	EntryPort     int      `json:"entry_port"`
	Proxy         string   `json:"proxy"`
	ProxyPort     int      `json:"proxy_port"`
	CreateAt      int64    `json:"create_at"`
	UpdateAt      int64    `json:"update_at"`
}
type ITunnelAppData interface {
	Add(e *TunnelApp) error
	GetByID(id int64) (*TunnelApp, error)
	GetByServerID(serverID int64) ([]*TunnelApp, error)
	GetByUserID(userID int64) ([]*TunnelApp, error)
	Update(e *TunnelApp) error
	GetByType(pageIndex, pageSize int, typ string) ([]*TunnelApp, error)
}
type tunnelAppData struct {
	db *sql.DB
}

func (d *data) NewTunnelAppData() ITunnelAppData {
	return &tunnelAppData{db: d.db}
}
func (d *tunnelAppData) Add(e *TunnelApp) error {
	customDomains := ""
	if len(e.CustomDomains) > 0 {
		bytes, err := json.Marshal(e.CustomDomains)
		if err != nil {
			return err
		}
		customDomains = string(bytes)
	}
	sqlStr := "insert into tunnel_app(user_id,server_id,`name`,`type`,`proxy_type`,`local_ip`,local_port,custom_domains,create_at,update_at)values(?,?,?,?,?,?,?,?,?,?) "
	res, err := d.db.Exec(sqlStr, e.UserID, e.ServerID, e.Name, e.Type, e.ProxyType, e.LocalIP, e.LocalPort, customDomains, e.CreateAt, e.UpdateAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}
func (d *tunnelAppData) GetByID(id int64) (*TunnelApp, error) {
	sqlStr := "select id,user_id,server_id,name,type,proxy_type,local_ip,local_port,custom_domains,entry_domain,entry_port,proxy,proxy_port,create_at,update_at from tunnel_app where id = ?"
	row := d.db.QueryRow(sqlStr, id)
	app := &TunnelApp{}
	var customDomains string
	err := row.Scan(&app.ID, &app.UserID, &app.ServerID, &app.Name, &app.Type, &app.ProxyType,
		&app.LocalIP, &app.LocalPort, &customDomains,
		&app.EntryDomain, &app.EntryPort,
		&app.Proxy, &app.ProxyPort,
		&app.CreateAt, &app.UpdateAt)
	if err != nil {
		return nil, err
	}
	if customDomains != "" {
		err = json.Unmarshal([]byte(customDomains), &app.CustomDomains)
		if err != nil {
			return nil, err
		}
	}
	return app, nil
}
func (d *tunnelAppData) GetByUserID(userID int64) ([]*TunnelApp, error) {
	sqlStr := "select id,user_id,server_id,name,type,proxy_type,local_ip,local_port,custom_domains,entry_domain,entry_port,proxy,proxy_port,create_at,update_at from tunnel_app where user_id = ?"
	rows, err := d.db.Query(sqlStr, userID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []*TunnelApp
	for rows.Next() {
		app := &TunnelApp{}
		var customDomains string
		err = rows.Scan(&app.ID, &app.UserID, &app.ServerID, &app.Name, &app.Type,
			&app.ProxyType, &app.LocalIP, &app.LocalPort,
			&customDomains,
			&app.EntryDomain, &app.EntryPort,
			&app.Proxy, &app.ProxyPort,
			&app.CreateAt, &app.UpdateAt)
		if err != nil {
			return nil, err
		}
		if customDomains != "" {
			err = json.Unmarshal([]byte(customDomains), &app.CustomDomains)
			if err != nil {
				return nil, err
			}
		}
		apps = append(apps, app)
	}
	return apps, nil
}
func (d *tunnelAppData) GetByServerID(serverID int64) ([]*TunnelApp, error) {
	sqlStr := "select id,user_id,server_id,name,type,proxy_type,local_ip,local_port,custom_domains,entry_domain,entry_port,proxy,proxy_port,create_at,update_at from tunnel_app where server_id = ?"
	rows, err := d.db.Query(sqlStr, serverID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []*TunnelApp
	for rows.Next() {
		app := &TunnelApp{}
		var customDomains string
		err = rows.Scan(&app.ID, &app.UserID, &app.ServerID, &app.Name, &app.Type,
			&app.ProxyType, &app.LocalIP, &app.LocalPort,
			&customDomains,
			&app.EntryDomain, &app.EntryPort,
			&app.Proxy, &app.ProxyPort,
			&app.CreateAt, &app.UpdateAt)
		if err != nil {
			return nil, err
		}
		if customDomains != "" {
			err = json.Unmarshal([]byte(customDomains), &app.CustomDomains)
			if err != nil {
				return nil, err
			}
		}
		apps = append(apps, app)
	}
	return apps, nil
}
func (d *tunnelAppData) Update(e *TunnelApp) error {
	customDomains := ""
	if len(e.CustomDomains) > 0 {
		bytes, err := json.Marshal(e.CustomDomains)
		if err != nil {
			return err
		}
		customDomains = string(bytes)
	}
	sqlStr := "update tunnel_app set `name`=?,`type`=?,proxy_type=?,local_ip=?,local_port=?,custom_domains=?,entry_domain=?,entry_port=?,proxy=?,proxy_port=?,update_at=? where id=?"
	_, err := d.db.Exec(sqlStr, e.Name, e.Type, e.ProxyType, e.LocalIP, e.LocalPort,
		customDomains, e.EntryDomain, e.EntryPort, e.Proxy, e.ProxyPort, e.UpdateAt, e.ID)
	return err
}
func (d *tunnelAppData) GetByType(pageIndex, pageSize int, typ string) ([]*TunnelApp, error) {
	if pageIndex < 1 {
		pageIndex = 1
	}
	if pageSize < 1 {
		pageSize = 1000
	}
	offset := (pageIndex - 1) * pageSize
	sqlStr := "select id,entry_domain,entry_port from tunnel_app where type = ? limit ?,?"
	rows, err := d.db.Query(sqlStr, typ, offset, pageSize)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TunnelApp, 0)
	for rows.Next() {
		item := &TunnelApp{}
		err = rows.Scan(&item.ID, &item.EntryDomain, &item.EntryPort)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}
