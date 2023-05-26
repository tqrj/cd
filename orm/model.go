package orm

import (
	"database/sql/driver"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// Model is the interface for all models.
// It only requires an Identity() method to return the primary key field
// name and value.
type Model interface {
	// Identity returns the primary key field of the model.
	// A very common case is that the primary key field is ID.
	Identity() (fieldName string, value any)
}

// BasicModel implements Model interface with an auto increment primary key ID.
//
// BasicModel is actually the gorm.Model struct which contains the following
// fields:
//
//	ID, CreatedAt, UpdatedAt, DeletedAt
//
// It is a good idea to embed this struct as the base struct for all models:
//
//	type User struct {
//	  orm.BasicModel
//	}
type BasicModel struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt *LocalTime
	UpdatedAt *LocalTime
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m BasicModel) Identity() (fieldName string, value any) {
	return "ID", m.ID
}

type LocalTime time.Time

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02 15:04:05"))), nil
}

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
