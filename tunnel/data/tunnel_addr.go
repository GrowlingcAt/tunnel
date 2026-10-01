package data

import (
	"database/sql"
	"time"
	"tunnel/pkg/zerror"
)

type TunnelAddr struct {
	ID       int64  `json:"id"`
	ServerID int64  `json:"server_id"`
	AppID    int64  `json:"app_id"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	PortType string `json:"port_type"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
}

type PortRange struct {
	ID       int64  `json:"id"`
	IP       string `json:"ip"`
	MinPort  int    `json:"min_port"`
	CurrPort int    `json:"curr_port"`
	MaxPort  int    `json:"max_port"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
}

type ITunnelAddrData interface {
	// 分配端口
	AssignPorts(list []*TunnelAddr, ip string, minPort int, maxPort int) error
	// 获取addr
	GetAddrList(serverID int64) ([]*TunnelAddr, error)
}

type tunnelAddrData struct {
	db *sql.DB
}

func (d *data) NewTunnelAddrData() ITunnelAddrData {
	return &tunnelAddrData{
		db: d.db,
	}
}
func (d *tunnelAddrData) AssignPorts(list []*TunnelAddr, ip string, minPort int, maxPort int) error {
	if len(list) == 0 {
		return nil
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	sqlStr := "select id,ip,min_port,curr_port,max_port from tunnel_port_range where ip = ? for update"
	row := tx.QueryRow(sqlStr, ip)
	portRange := &PortRange{}
	err = row.Scan(&portRange.ID, &portRange.IP, &portRange.MinPort, &portRange.CurrPort, &portRange.MaxPort)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}
	if err == sql.ErrNoRows {
		portRange.IP = ip
		portRange.MinPort = minPort
		portRange.MaxPort = maxPort
		portRange.CurrPort = minPort - 1
	}
	portNum := len(list)
	if portRange.CurrPort+portNum > portRange.MaxPort {
		tx.Rollback()
		err = zerror.NewByMsg("无可用端口，请联系管理员")
		return err
	}
	for _, addr := range list {
		portRange.CurrPort += 1
		addr.Port = portRange.CurrPort
	}
	if portRange.ID == 0 {
		sqlStr = "insert into tunnel_port_range(ip,min_port,curr_port,max_port,create_at,update_at)values(?,?,?,?,?,?)"
		_, err = tx.Exec(sqlStr, portRange.IP, portRange.MinPort, portRange.CurrPort, portRange.MaxPort, time.Now().Unix(), time.Now().Unix())
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		sqlStr = "update tunnel_port_range set curr_port = ?,update_at = ? where id = ?"
		_, err = tx.Exec(sqlStr, portRange.CurrPort, portRange.UpdateAt, portRange.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, addr := range list {
		sqlStr = "insert into tunnel_addr(server_id,app_id,ip,port,port_type,create_at,update_at)values(?,?,?,?,?,?,?)"
		_, err = tx.Exec(sqlStr, addr.ServerID, addr.AppID, addr.IP, addr.Port, addr.PortType, time.Now().Unix(), time.Now().Unix())
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	return err
}
func (d *tunnelAddrData) GetAddrList(serverID int64) ([]*TunnelAddr, error) {
	sqlStr := "select id,server_id,app_id,ip,port,port_type from tunnel_addr where server_id = ?"
	rows, err := d.db.Query(sqlStr, serverID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*TunnelAddr
	for rows.Next() {
		addr := &TunnelAddr{}
		err = rows.Scan(&addr.ID, &addr.ServerID, &addr.AppID, &addr.IP, &addr.Port, &addr.PortType)
		if err != nil {
			return nil, err
		}
		list = append(list, addr)
	}
	return list, nil
}
