// Package connector ดูแลการเชื่อมต่อระบบภายนอก (ที่นี่คือ MongoDB)
package connector

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoConnectTimeout         = 10 * time.Second
	mongoServerSelectionTimeout = 10 * time.Second // ไม่รอ 30 วินาทีตามค่า default ตอน DB ล่ม
	mongoPingTimeout            = 10 * time.Second
	mongoDisconnectTimeout      = 10 * time.Second
)

type MongoDB struct {
	Client *mongo.Client
}

// NewMongoDBClient คือ Constructor ของ connector
// คืน error แทน client ที่ยังสร้างไม่เสร็จ → main.go หยุดได้ตั้งแต่ตอนเริ่ม
func NewMongoDBClient(mongodbURI string, maxPoolSize uint64) (*MongoDB, error) {
	option := options.Client().
		ApplyURI(mongodbURI).
		SetMaxPoolSize(maxPoolSize).
		SetConnectTimeout(mongoConnectTimeout).
		SetServerSelectionTimeout(mongoServerSelectionTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), mongoConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		return nil, err
	}
	return &MongoDB{Client: client}, nil
}

// Ping ตรวจว่าต่อ DB ได้จริง (mongo.Connect ยังไม่ได้ต่อจริงจนกว่าจะมีการใช้งาน)
func (conn *MongoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoPingTimeout)
	defer cancel()
	return conn.Client.Ping(ctx, nil)
}

// SelectDB คืน database ให้ service/storage ใช้
func (conn *MongoDB) SelectDB(dbName string) *mongo.Database {
	return conn.Client.Database(dbName)
}

func (conn *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoDisconnectTimeout)
	defer cancel()
	return conn.Client.Disconnect(ctx)
}
