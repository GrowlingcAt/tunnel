package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	Delete(id int64) error
}
type tunnelAppData struct {
	db *sql.DB
}

func (d *data) NewTunnelAppData() ITunnelAppData {
	return &tunnelAppData{db: d.db}
}
func (d *tunnelAppData) Add(e *TunnelApp) error {

	// 默认 custom_domains 为空字符串
	customDomains := ""

	// 如果应用配置了自定义域名
	if len(e.CustomDomains) > 0 {

		// 将 []string 转换成 JSON 字符串
		// 例如：["a.com","b.com"]
		bytes, err := json.Marshal(e.CustomDomains)
		if err != nil {
			return err
		}

		// []byte 转换成 string，准备写入数据库
		customDomains = string(bytes)
	}

	// 插入一条应用记录
	sqlStr := "insert into tunnel_app(user_id,server_id,`name`,`type`,`proxy_type`,`local_ip`,local_port,custom_domains,create_at,update_at)values(?,?,?,?,?,?,?,?,?,?) "

	// 执行 INSERT
	res, err := d.db.Exec(
		sqlStr,
		e.UserID,
		e.ServerID,
		e.Name,
		e.Type,
		e.ProxyType,
		e.LocalIP,
		e.LocalPort,
		customDomains,
		e.CreateAt,
		e.UpdateAt,
	)
	if err != nil {
		return err
	}

	// 获取数据库自动生成的自增 ID
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// 将数据库生成的 ID 回写到当前对象
	e.ID = id

	return nil
}

// GetByID 根据应用 ID 查询单个应用的完整信息
func (d *tunnelAppData) GetByID(id int64) (*TunnelApp, error) {

	// 根据应用 ID 查询应用的所有相关字段
	sqlStr := "select id,user_id,server_id,name,type,proxy_type,local_ip,local_port,custom_domains,entry_domain,entry_port,proxy,proxy_port,create_at,update_at from tunnel_app where id = ?"

	// QueryRow 用于查询单条记录
	row := d.db.QueryRow(sqlStr, id)

	fmt.Printf("GetByID id=%d\n", id)

	// 创建 TunnelApp 对象，用于接收查询结果
	app := &TunnelApp{}

	// custom_domains 在数据库中以 JSON 字符串形式存储
	// 先使用 string 接收，之后再转换为 []string
	var customDomains string

	// 将数据库查询到的字段映射到 TunnelApp 对象
	err := row.Scan(
		&app.ID,
		&app.UserID,
		&app.ServerID,
		&app.Name,
		&app.Type,
		&app.ProxyType,
		&app.LocalIP,
		&app.LocalPort,
		&customDomains,
		&app.EntryDomain,
		&app.EntryPort,
		&app.Proxy,
		&app.ProxyPort,
		&app.CreateAt,
		&app.UpdateAt,
	)

	// 查询或读取数据失败，返回错误
	if err != nil {
		fmt.Printf("GetByID err=%v\n", err)
		return nil, err
	}

	// 如果 custom_domains 不为空
	if customDomains != "" {

		// 将 JSON 字符串反序列化为 []string
		err = json.Unmarshal([]byte(customDomains), &app.CustomDomains)

		// JSON 解析失败，返回错误
		if err != nil {
			return nil, err
		}
	}

	// 返回查询到的应用对象
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

// GetByServerID 根据 Tunnel Server 的 ID，查询该 Server 下的所有应用。
func (d *tunnelAppData) GetByServerID(serverID int64) ([]*TunnelApp, error) {

	// 根据 server_id 查询应用，并获取应用的完整信息
	sqlStr := "select id,user_id,server_id,name,type,proxy_type,local_ip,local_port,custom_domains,entry_domain,entry_port,proxy,proxy_port,create_at,update_at from tunnel_app where server_id = ?"

	// 执行 SQL 查询
	rows, err := d.db.Query(sqlStr, serverID)

	// 没有查询到数据
	if err == sql.ErrNoRows {
		return nil, nil
	}

	// 数据库查询出错
	if err != nil {
		return nil, err
	}

	// 函数结束时关闭查询结果
	defer rows.Close()

	// 保存查询到的所有应用
	var apps []*TunnelApp

	// 遍历查询结果
	for rows.Next() {

		// 创建一个 TunnelApp 对象
		app := &TunnelApp{}

		// custom_domains 在数据库中保存的是 JSON 字符串，
		// 所以先使用 string 接收
		var customDomains string

		// 将数据库查询结果映射到 TunnelApp 对象
		err = rows.Scan(
			&app.ID,
			&app.UserID,
			&app.ServerID,
			&app.Name,
			&app.Type,
			&app.ProxyType,
			&app.LocalIP,
			&app.LocalPort,
			&customDomains,
			&app.EntryDomain,
			&app.EntryPort,
			&app.Proxy,
			&app.ProxyPort,
			&app.CreateAt,
			&app.UpdateAt,
		)

		// 数据读取失败
		if err != nil {
			return nil, err
		}

		// 如果数据库中的 custom_domains 不为空
		if customDomains != "" {

			// 将 JSON 字符串转换成 []string
			err = json.Unmarshal([]byte(customDomains), &app.CustomDomains)

			// JSON 解析失败
			if err != nil {
				return nil, err
			}
		}

		// 将当前应用加入应用列表
		apps = append(apps, app)
	}

	// 返回该 Server 下的所有应用
	return apps, nil
}
func (d *tunnelAppData) Update(e *TunnelApp) error {

	// CustomDomains 是 []string
	// 数据库里不能直接存切片，所以先转换成 JSON 字符串
	customDomains := ""

	if len(e.CustomDomains) > 0 {
		bytes, err := json.Marshal(e.CustomDomains)
		if err != nil {
			return err
		}

		customDomains = string(bytes)
	}

	// 更新 tunnel_app 表
	// 根据 App.ID 找到对应的应用
	sqlStr := `
        update tunnel_app
        set
            name=?,
            type=?,
            proxy_type=?,
            local_ip=?,
            local_port=?,
            custom_domains=?,
            entry_domain=?,
            entry_port=?,
            proxy=?,
            proxy_port=?,
            update_at=?
        where id=?
    `

	// 执行SQL
	_, err := d.db.Exec(
		sqlStr,

		e.Name,        // 应用名称
		e.Type,        // 应用类型
		e.ProxyType,   // 代理类型
		e.LocalIP,     // 内网IP
		e.LocalPort,   // 内网端口
		customDomains, // 自定义域名
		e.EntryDomain, // 对外域名
		e.EntryPort,   // 对外端口
		e.Proxy,       // 代理
		e.ProxyPort,   // 代理端口
		e.UpdateAt,    // 更新时间

		e.ID, // 根据ID定位应用
	)

	return err
}

// GetByType 根据应用类型查询应用列表，并支持分页。
// 主要返回应用 ID、入口域名和入口端口。
func (d *tunnelAppData) GetByType(pageIndex, pageSize int, typ string) ([]*TunnelApp, error) {

	// 页码不能小于 1，默认从第 1 页开始
	if pageIndex < 1 {
		pageIndex = 1
	}

	// 每页数量不能小于 1，默认最多查询 1000 条
	if pageSize < 1 {
		pageSize = 1000
	}

	// 计算分页查询的起始位置
	// 例如 pageIndex=2、pageSize=100：
	// offset = (2-1)*100 = 100
	offset := (pageIndex - 1) * pageSize

	// 根据应用类型查询应用
	// LIMIT offset, pageSize 用于分页
	sqlStr := "select id,entry_domain,entry_port from tunnel_app where type = ? limit ?,?"

	// 执行 SQL 查询
	rows, err := d.db.Query(sqlStr, typ, offset, pageSize)

	// 没有查询到数据时，返回空结果
	if err == sql.ErrNoRows {
		return nil, nil
	}

	// 查询数据库出错
	if err != nil {
		return nil, err
	}

	// 函数结束时关闭查询结果集
	defer rows.Close()

	// 保存查询到的应用
	list := make([]*TunnelApp, 0)

	// 遍历查询结果
	for rows.Next() {
		// 创建一个 TunnelApp 对象
		item := &TunnelApp{}

		// 将数据库中的字段扫描到 TunnelApp 对象
		err = rows.Scan(
			&item.ID,
			&item.EntryDomain,
			&item.EntryPort,
		)
		if err != nil {
			return nil, err
		}

		// 将当前应用加入结果列表
		list = append(list, item)
	}

	// 返回查询到的应用列表
	return list, nil
}

func (d *tunnelAppData) Delete(appID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// 删除应用对应的地址
	_, err = tx.Exec(
		"delete from tunnel_addr where app_id = ?",
		appID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除应用
	_, err = tx.Exec(
		"delete from tunnel_app where id = ?",
		appID,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
