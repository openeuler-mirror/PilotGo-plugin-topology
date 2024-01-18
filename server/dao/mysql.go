package dao

import (
	"database/sql"
	"fmt"
	"strings"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/agentmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/conf"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Global_mysql *MysqlClient

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
		err = errors.Wrapf(err, "**fatal**2") // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
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
		err := errors.Errorf("mysql connect failed: %s(url: %s) **fatal**2", err.Error(), url) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	var db *sql.DB
	if db, err = m.db.DB(); err != nil {
		err = errors.Errorf("get mysql sql.db failed: %s **fatal**2", err.Error()) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	// mysql 模型迁移
	err = m.db.AutoMigrate(&meta.TopoConfiguration{})
	if err != nil {
		err = errors.Errorf("mysql automigrate failed: %s **fatal**2", err.Error()) // err top
		agentmanager.ErrorTransmit(agentmanager.Topo.Tctx, err, agentmanager.Topo.ErrCh, true)
	}

	return m
}

func ensureDatabase(conf *conf.MysqlConf) error {
	if conf == nil {
		err := errors.New("mysql config error **fatal**1")
		return err
	}

	if conf.Addr == "" || conf.Username == "" || conf.Password == "" || conf.DB == "" {
		err := errors.Errorf("mysql config error: addr(%s) username(%s) password(%s) db(%s) **fatal**1", conf.Addr, conf.Username, conf.Password, conf.DB)
		return err
	}

	addr_arr := strings.Split(conf.Addr, ":")
	if len(addr_arr) != 2 {
		err := errors.Errorf("mysql addr error: %s **fatal**2", conf.Addr)
		return err
	}

	url := fmt.Sprintf("%s:%s@(%s:%s)/?charset=utf8mb4&parseTime=true", conf.Username, conf.Password, addr_arr[0], addr_arr[1])

	db, err := gorm.Open(mysql.Open(url))
	if err != nil {
		err := errors.Errorf("mysql connect failed: %s **fatal**2", err.Error())
		return err
	}

	creatDataBase := "CREATE DATABASE IF NOT EXISTS " + conf.DB + " DEFAULT CHARSET utf8 COLLATE utf8_general_ci"
	db.Exec(creatDataBase)

	d, err := db.DB()
	if err != nil {
		err = errors.Errorf("get mysql sql.db failed: %s **fatal**2", err.Error())
		return err
	}
	if err = d.Close(); err != nil {
		err = errors.Errorf("close mysql sql.db failed: %s **fatal**2", err.Error())
		return err
	}
	return nil
}

func (m *MysqlClient) QueryTopoConfiguration(tcid uint) (*meta.TopoConfiguration, error) {
	var tc *meta.TopoConfiguration = new(meta.TopoConfiguration)
	err := m.db.Model(&meta.TopoConfiguration{}).Where("id=?", tcid).First(tc).Error
	if err != nil {
		err = errors.Errorf("query topo configuration failed: %s, %d", err.Error(), tcid)
		return nil, err
	}
	return tc, nil
}

func (m *MysqlClient) AddTopoConfiguration(tc *meta.TopoConfiguration) error {
	_tc := tc
	err := m.db.Save(_tc).Error
	if err != nil {
		err = errors.Errorf("add topo configuration failed: %s, %+v", err.Error(), tc)
		return err
	}
	return nil
}

func (m *MysqlClient) DeleteTopoConfiguration(tcid uint) error {
	err := m.db.Where("id = ?", tcid).Unscoped().Delete(meta.TopoConfiguration{}).Error
	if err != nil {
		err = errors.Errorf("delete topo configuration failed: %s, %d", err.Error(), tcid)
		return err
	}
	return nil
}
