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
	// 根据应用 ID 删除对应地址
	DeleteByAppID(appID int64) error
}

type tunnelAddrData struct {
	db *sql.DB
}

func (d *data) NewTunnelAddrData() ITunnelAddrData {
	return &tunnelAddrData{
		db: d.db,
	}
}

// AssignPorts 为一批 TunnelAddr 分配可用端口，并将地址信息写入数据库。
func (d *tunnelAddrData) AssignPorts(list []*TunnelAddr, ip string, minPort int, maxPort int) error {

	// 没有需要分配端口的地址，直接返回
	if len(list) == 0 {
		return nil
	}

	// 开启数据库事务
	// 后面的端口范围更新和 tunnel_addr 写入要么全部成功，要么全部回滚
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// 查询指定 IP 的端口范围，并加行锁
	// FOR UPDATE 可以防止多个请求同时分配相同端口
	sqlStr := "select id,ip,min_port,curr_port,max_port from tunnel_port_range where ip = ? for update"
	row := tx.QueryRow(sqlStr, ip)

	// 创建端口范围对象
	portRange := &PortRange{}

	// 将数据库中的端口范围信息读取到 portRange
	err = row.Scan(
		&portRange.ID,
		&portRange.IP,
		&portRange.MinPort,
		&portRange.CurrPort,
		&portRange.MaxPort,
	)

	// 查询出现其他错误，回滚事务
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	// 如果该 IP 还没有端口范围记录
	if err == sql.ErrNoRows {

		// 初始化端口范围
		portRange.IP = ip
		portRange.MinPort = minPort
		portRange.MaxPort = maxPort

		// 当前端口从最小端口的前一个开始，
		// 后面分配时先 +1，因此第一个分配到 minPort
		portRange.CurrPort = minPort - 1
	}

	// 当前需要分配的端口数量
	portNum := len(list)

	// 判断剩余端口是否足够
	if portRange.CurrPort+portNum > portRange.MaxPort {
		tx.Rollback()

		// 没有可用端口
		err = zerror.NewByMsg("无可用端口，请联系管理员")
		return err
	}

	// 给每个 TunnelAddr 分配端口
	for _, addr := range list {

		// 当前端口 +1
		portRange.CurrPort += 1

		// 将分配到的端口保存到 TunnelAddr
		addr.Port = portRange.CurrPort
	}

	// 如果之前不存在端口范围记录
	if portRange.ID == 0 {

		// 创建新的端口范围记录
		sqlStr = "insert into tunnel_port_range(ip,min_port,curr_port,max_port,create_at,update_at)values(?,?,?,?,?,?)"

		_, err = tx.Exec(
			sqlStr,
			portRange.IP,
			portRange.MinPort,
			portRange.CurrPort,
			portRange.MaxPort,
			time.Now().Unix(),
			time.Now().Unix(),
		)

		if err != nil {
			tx.Rollback()
			return err
		}

	} else {

		// 已经存在端口范围，只更新当前已经分配到哪个端口
		sqlStr = "update tunnel_port_range set curr_port = ?,update_at = ? where id = ?"

		_, err = tx.Exec(
			sqlStr,
			portRange.CurrPort,
			portRange.UpdateAt,
			portRange.ID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 将每个 TunnelAddr 写入 tunnel_addr 表
	for _, addr := range list {

		sqlStr = "insert into tunnel_addr(server_id,app_id,ip,port,port_type,create_at,update_at)values(?,?,?,?,?,?,?)"

		_, err = tx.Exec(
			sqlStr,
			addr.ServerID,
			addr.AppID,
			addr.IP,
			addr.Port,
			addr.PortType,
			time.Now().Unix(),
			time.Now().Unix(),
		)

		// 写入失败，回滚整个事务
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 所有操作成功，提交事务
	err = tx.Commit()

	return err
}

// GetAddrList 根据 Tunnel Server ID 查询该 Server 下的所有地址信息
func (d *tunnelAddrData) GetAddrList(serverID int64) ([]*TunnelAddr, error) {

	// 查询指定 Server 下的所有 tunnel_addr
	sqlStr := "select id,server_id,app_id,ip,port,port_type from tunnel_addr where server_id = ?"

	// 执行查询
	rows, err := d.db.Query(sqlStr, serverID)

	// 没有查询到数据
	if err == sql.ErrNoRows {
		return nil, nil
	}

	// 查询数据库出错
	if err != nil {
		return nil, err
	}

	// 函数结束时关闭结果集
	defer rows.Close()

	// 保存查询到的地址列表
	var list []*TunnelAddr

	// 遍历查询结果
	for rows.Next() {

		// 创建 TunnelAddr 对象
		addr := &TunnelAddr{}

		// 将数据库字段映射到 TunnelAddr
		err = rows.Scan(
			&addr.ID,
			&addr.ServerID,
			&addr.AppID,
			&addr.IP,
			&addr.Port,
			&addr.PortType,
		)

		// 数据读取失败
		if err != nil {
			return nil, err
		}

		// 将当前地址加入列表
		list = append(list, addr)
	}

	// 返回该 Server 下的所有地址
	return list, nil
}
func (d *tunnelAddrData) DeleteByAppID(appID int64) error {
	sqlStr := "delete from tunnel_addr where app_id = ?"
	_, err := d.db.Exec(sqlStr, appID)
	return err
}
