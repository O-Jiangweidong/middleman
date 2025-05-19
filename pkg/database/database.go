package database

import (
    "database/sql"
    "fmt"
    "log"
    "sync"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    
    "middleman/pkg/config"
    "middleman/pkg/database/models"
)

const (
    DefaultDBName = "middleman"
)

type Manager struct {
    dbs map[string]*gorm.DB
    dsn string
    mu  sync.RWMutex
}

func NewDatabaseManager() *Manager {
    conf := config.GetConf()
    dsn := fmt.Sprintf(
        "host=%s port=%v user=%s password=%s sslmode=disable",
        conf.DBHost, conf.DBPort, conf.DBUser, conf.DBPwd,
    )
    dm := &Manager{
        dbs: make(map[string]*gorm.DB), dsn: dsn,
    }
    if err := dm.InitDatabaseManager(); err != nil {
        log.Fatalf("init database manager failed: %v", err)
    }
    return dm
}

func (dm *Manager) InitDatabaseManager() error {
    if err := dm.createDatabase(DefaultDBName); err != nil {
        return err
    }
    db, err := dm.connectDB(DefaultDBName, func(db *gorm.DB) error {
        return db.AutoMigrate(&models.JumpServer{})
    })
    if err != nil {
        return err
    }
    dm.dbs[DefaultDBName] = db
    return nil
}

func (dm *Manager) GetDB(name string) (*gorm.DB, error) {
    dm.mu.RLock()
    db, exists := dm.dbs[name]
    dm.mu.RUnlock()
    
    if exists {
        return db, nil
    }
    
    dm.mu.Lock()
    defer dm.mu.Unlock()
    
    if db, exists = dm.dbs[name]; exists {
        return db, nil
    }
    
    err := dm.createDatabase(name)
    if err != nil {
        return nil, fmt.Errorf("failed to create database %s: %v", name, err)
    }
    
    db, err = dm.connectDB(name, func(db *gorm.DB) error {
        return db.AutoMigrate()
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to new database %s: %v", name, err)
    }
    
    dm.dbs[name] = db
    log.Printf("Database %s created and initialized successfully", name)
    return db, nil
}

func (dm *Manager) createDatabase(dbName string) error {
    sqlDB, err := sql.Open("postgres", dm.dsn)
    if err != nil {
        return err
    }
    defer sqlDB.Close()
    
    var exists bool
    err = sqlDB.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)
	`, dbName).Scan(&exists)
    
    if err != nil {
        return fmt.Errorf("检查数据库是否存在失败: %v", err)
    }
    
    if !exists {
        _, err = sqlDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
        if err != nil {
            return fmt.Errorf("创建数据库失败: %v", err)
        }
        log.Printf("数据库 %s 创建成功", dbName)
    } else {
        log.Printf("数据库 %s 已存在", dbName)
    }
    return nil
}

type DBConnectCallback func(db *gorm.DB) error

func (dm *Manager) connectDB(dbName string, callbacks ...DBConnectCallback) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s dbname=%s", dm.dsn, dbName)
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    for _, callback := range callbacks {
        if callback != nil {
            if err = callback(db); err != nil {
                return nil, err
            }
        }
    }
    return db, nil
}

var DBManager = NewDatabaseManager()
