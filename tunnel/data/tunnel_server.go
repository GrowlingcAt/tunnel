package data

import "database/sql"

type TunnelServer struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	Token        string `json:"token"`
	ServerConfig string `json:"server_config"`
	ClientConfig string `json:"client_config"`
	Deploy       int    `json:"deploy"`
	CreateAt     int64  `json:"create_at"`
	UpdateAt     int64  `json:"update_at"`
}

type ITunnelServerData interface {
	Add(e *TunnelServer) error
	GetByUserID(userID int64) (*TunnelServer, error)
	UpdateDeployStatus(e *TunnelServer) error
	UpdateConfigs(e *TunnelServer) error
}

type tunnelServerData struct {
	db *sql.DB
}

func (d *data) NewTunnelServerData() ITunnelServerData {
	return &tunnelServerData{
		db: d.db,
	}
}

func (d *tunnelServerData) Add(e *TunnelServer) error {
	sqlStr := "insert into tunnel_server(user_id,token,deploy,create_at,update_at)values(?,?,?,?,?)"
	res, err := d.db.Exec(sqlStr, e.UserID, e.Token, e.Deploy, e.CreateAt, e.UpdateAt)
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
func (d *tunnelServerData) GetByUserID(userID int64) (*TunnelServer, error) {
	sqlStr := "select id,token,server_config,client_config,deploy from tunnel_server where user_id = ?"
	row := d.db.QueryRow(sqlStr, userID)
	server := &TunnelServer{}
	serverConfig := sql.NullString{}
	clientConfig := sql.NullString{}
	err := row.Scan(&server.ID, &server.Token, &serverConfig, &clientConfig, &server.Deploy)
	if serverConfig.Valid {
		server.ServerConfig = serverConfig.String
	}
	if clientConfig.Valid {
		server.ClientConfig = clientConfig.String
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return server, err
}
func (d *tunnelServerData) UpdateDeployStatus(e *TunnelServer) error {
	sqlStr := "update tunnel_server set deploy = ?, update_at = ? where id = ?"
	_, err := d.db.Exec(sqlStr, e.Deploy, e.UpdateAt, e.ID)
	return err
}
func (d *tunnelServerData) UpdateConfigs(e *TunnelServer) error {
	sqlStr := "update tunnel_server set server_config = ?, client_config = ?,deploy=?, update_at = ? where id = ?"
	_, err := d.db.Exec(sqlStr, e.ServerConfig, e.ClientConfig, e.Deploy, e.UpdateAt, e.ID)
	return err
}
