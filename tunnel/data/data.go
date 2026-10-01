package data

import "database/sql"

type IData interface {
	NewTunnelServerData() ITunnelServerData
	NewTunnelAppData() ITunnelAppData
	NewTunnelAddrData() ITunnelAddrData
}
type data struct {
	db *sql.DB
}

func NewData(db *sql.DB) IData {
	return &data{
		db: db,
	}
}
