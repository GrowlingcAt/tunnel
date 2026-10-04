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

// Add 向 tunnel_server 表中新增一个 Tunnel Server
func (d *tunnelServerData) Add(e *TunnelServer) error {

	// 插入 Tunnel Server 的基本信息
	sqlStr := "insert into tunnel_server(user_id,token,deploy,create_at,update_at)values(?,?,?,?,?)"
	// 执行 INSERT
	res, err := d.db.Exec(
		sqlStr,
		e.UserID,   // 所属用户
		e.Token,    // Server 的认证 Token
		e.Deploy,   // 当前部署状态
		e.CreateAt, // 创建时间
		e.UpdateAt, // 更新时间
	)
	if err != nil {
		return err
	}
	// 获取数据库自动生成的自增 ID
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	// 将数据库生成的 ID 回写到 TunnelServer 对象
	e.ID = id
	return nil
}

func (d *tunnelServerData) GetByUserID(userID int64) (*TunnelServer, error) {

	// 根据用户 ID 查询对应的 Tunnel Server
	sqlStr := "select id,token,server_config,client_config,deploy from tunnel_server where user_id = ?"

	// 查询单条记录
	row := d.db.QueryRow(sqlStr, userID)

	// 创建 TunnelServer 对象，用于接收查询结果
	server := &TunnelServer{}

	// server_config 和 client_config 可能为 NULL，
	// 所以使用 sql.NullString 接收
	serverConfig := sql.NullString{}
	clientConfig := sql.NullString{}

	// 将查询结果依次保存到对应变量中
	err := row.Scan(
		&server.ID,
		&server.Token,
		&serverConfig,
		&clientConfig,
		&server.Deploy,
	)

	// 如果 server_config 不为 NULL，
	// 将数据库中的值赋值给 ServerConfig
	if serverConfig.Valid {
		server.ServerConfig = serverConfig.String
	}

	// 如果 client_config 不为 NULL，
	// 将数据库中的值赋值给 ClientConfig
	if clientConfig.Valid {
		server.ClientConfig = clientConfig.String
	}

	// 如果没有找到该用户对应的 Tunnel Server
	if err == sql.ErrNoRows {
		return nil, nil
	}

	// 返回查询到的 Tunnel Server 和错误信息
	return server, err
}

func (d *tunnelServerData) UpdateDeployStatus(e *TunnelServer) error {

	// 更新指定 Tunnel Server 的部署状态和更新时间
	sqlStr := "update tunnel_server set deploy = ?, update_at = ? where id = ?"

	// 执行 UPDATE
	// 根据 Server ID 找到对应记录，
	// 将 deploy 和 update_at 更新为新的值
	_, err := d.db.Exec(
		sqlStr,
		e.Deploy,
		e.UpdateAt,
		e.ID,
	)

	// 返回更新过程中产生的错误
	return err
}

func (d *tunnelServerData) UpdateConfigs(e *TunnelServer) error {
	// 更新 Tunnel Server 的服务端配置、客户端配置、
	// 部署状态和更新时间
	sqlStr := "update tunnel_server set server_config = ?, client_config = ?,deploy=?, update_at = ? where id = ?"

	// 执行 UPDATE
	// 根据 Server ID 找到对应的 Tunnel Server，
	// 更新它的配置和部署状态
	_, err := d.db.Exec(
		sqlStr,
		e.ServerConfig, // 服务端配置
		e.ClientConfig, // 客户端配置
		e.Deploy,       // 部署状态
		e.UpdateAt,     // 更新时间
		e.ID,           // Server ID
	)
	// 返回更新过程中产生的错误
	return err
}
