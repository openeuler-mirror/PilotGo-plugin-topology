package mysqlmanager

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/errormanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/pluginclient"
	"gitee.com/openeuler/PilotGo/sdk/response"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Global_Mysql *MysqlClient

type MysqlClient struct {
	ip       string
	port     string
	username string
	password string
	dbname   string
	db       *gorm.DB
}

func MysqldbInit(conf *conf.MysqlConf) *MysqlClient {
	err := ensureDatabase(conf)
	if err != nil {
		err = errors.Wrapf(err, "**errstackfatal**2") // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
	}

	m := &MysqlClient{
		ip:       strings.Split(conf.Addr, ":")[0],
		port:     strings.Split(conf.Addr, ":")[1],
		username: conf.Username,
		password: conf.Password,
		dbname:   conf.DB,
	}

	url := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=true", m.username, m.password, m.ip, m.port, m.dbname)

	m.db, err = gorm.Open(mysql.Open(url), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		err := errors.Errorf("mysql connect failed: %s(url: %s) **errstackfatal**2", err.Error(), url) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
	}

	var db *sql.DB
	if db, err = m.db.DB(); err != nil {
		err = errors.Errorf("get mysql sql.db failed: %s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	// mysql 模型迁移
	err = m.db.AutoMigrate(&Topo_configuration_DB{})
	if err != nil {
		err = errors.Errorf("mysql automigrate failed: %s **errstackfatal**2", err.Error()) // err top
		errormanager.ErrorTransmit(pluginclient.GlobalContext, err, true)
	}

	return m
}

func ensureDatabase(conf *conf.MysqlConf) error {
	if conf == nil {
		err := errors.New("mysql config error **errstackfatal**1")
		return err
	}

	if conf.Addr == "" || conf.Username == "" || conf.Password == "" || conf.DB == "" {
		err := errors.Errorf("mysql config error: addr(%s) username(%s) password(%s) db(%s) **errstackfatal**1", conf.Addr, conf.Username, conf.Password, conf.DB)
		return err
	}

	addr_arr := strings.Split(conf.Addr, ":")
	if len(addr_arr) != 2 {
		err := errors.Errorf("mysql addr error: %s **errstackfatal**2", conf.Addr)
		return err
	}

	url := fmt.Sprintf("%s:%s@(%s:%s)/?charset=utf8mb4&parseTime=true", conf.Username, conf.Password, addr_arr[0], addr_arr[1])

	db, err := gorm.Open(mysql.Open(url))
	if err != nil {
		err := errors.Errorf("mysql connect failed: %s **errstackfatal**2", err.Error())
		return err
	}

	creatDataBase := "CREATE DATABASE IF NOT EXISTS " + conf.DB + " DEFAULT CHARSET utf8 COLLATE utf8_general_ci"
	db.Exec(creatDataBase)

	d, err := db.DB()
	if err != nil {
		err = errors.Errorf("get mysql sql.db failed: %s **errstackfatal**2", err.Error())
		return err
	}
	if err = d.Close(); err != nil {
		err = errors.Errorf("close mysql sql.db failed: %s **errstackfatal**2", err.Error())
		return err
	}
	return nil
}

func (m *MysqlClient) QuerySingleTopoConfiguration(tcid uint) (*Topo_configuration_DB, error) {
	var tcdb *Topo_configuration_DB = new(Topo_configuration_DB)

	if err := m.db.Model(&Topo_configuration_DB{}).Where("id=?", tcid).First(tcdb).Error; err != nil {
		err = errors.Errorf("query topo configuration failed: %s, %d **errstack**0", err.Error(), tcid)
		return nil, err
	}

	return tcdb, nil
}

func (m *MysqlClient) QueryTopoConfigurationList(query *response.PaginationQ) ([]*Topo_configuration_DB, int, error) {
	tcdbs := make([]*Topo_configuration_DB, 0)
	if err := m.db.Order("id desc").Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize).Find(&tcdbs).Error; err != nil {
		return nil, 0, errors.Errorf("query topo configuration list failed: %s **errstack**0", err.Error())
	}

	var total int64
	if err := m.db.Model(&Topo_configuration_DB{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return tcdbs, int(total), nil
}

func (m *MysqlClient) AddTopoConfiguration(tc *Topo_configuration_DB) (int, error) {
	_tc := tc
	if err := m.db.Save(_tc).Error; err != nil {
		err = errors.Errorf("add topo configuration failed: %s, %+v **errstack**0", err.Error(), tc)
		return -1, err
	}

	return int(_tc.ID), nil
}

func (m *MysqlClient) DeleteTopoConfiguration(tcid uint) error {
	if err := m.db.Where("id = ?", tcid).Unscoped().Delete(Topo_configuration_DB{}).Error; err != nil {
		return errors.Errorf("delete topo configuration failed: %s, %d **errstack**0", err.Error(), tcid)
	}

	return nil
}

func (m *MysqlClient) TopoConfigurationToDB(tc *Topo_configuration) (*Topo_configuration_DB, error) {
	var tcdb *Topo_configuration_DB = new(Topo_configuration_DB)

	noderules_bytes, noderules_err := json.Marshal(tc.NodeRules)
	tagrules_bytes, tagrules_err := json.Marshal(tc.TagRules)
	if noderules_err != nil || tagrules_err != nil {
		err := errors.Errorf("json marshal error: noderules(%s) tagrules)%s **errstack**4", noderules_err, tagrules_err)
		return nil, err
	}

	tcdb.ID = tc.ID
	tcdb.Name = tc.Name
	tcdb.Description = tc.Description
	tcdb.CreatedAt = tc.CreatedAt
	tcdb.UpdatedAt = tc.UpdatedAt
	tcdb.Version = tc.Version
	tcdb.Preserve = tc.Preserve
	tcdb.BatchId = tc.BatchId
	tcdb.NodeRules = string(noderules_bytes)
	tcdb.TagRules = string(tagrules_bytes)

	return tcdb, nil
}

func (m *MysqlClient) DBToTopoConfiguration(tcdb *Topo_configuration_DB) (*Topo_configuration, error) {
	var tc *Topo_configuration = new(Topo_configuration)

	noderules_err := json.Unmarshal([]byte(tcdb.NodeRules), &tc.NodeRules)
	tagrules_err := json.Unmarshal([]byte(tcdb.TagRules), &tc.TagRules)
	if noderules_err != nil || tagrules_err != nil {
		err := errors.Errorf("json unmarshal error: noderules(%s) tagrules)%s **errstack**4", noderules_err, tagrules_err)
		return nil, err
	}

	tc.ID = tcdb.ID
	tc.Name = tcdb.Name
	tc.Description = tcdb.Description
	tc.CreatedAt = tcdb.CreatedAt
	tc.UpdatedAt = tcdb.UpdatedAt
	tc.Version = tcdb.Version
	tc.Preserve = tcdb.Preserve
	tc.BatchId = tcdb.BatchId

	return tc, nil
}
