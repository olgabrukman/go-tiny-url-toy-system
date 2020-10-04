package dbconnection

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

const testTableName = "urlhash"

type DBItem struct {
	URL      string
	Encoding string
}

type MyCollection struct {
	collection *mongo.Collection
}

func NewDBCollectionConnection(configName string) (*MyCollection, error) {
	dbHost, dbPort, dbName, tableName, err := readDbConfig(configName)
	if err != nil {
		log.Fatalf("Failed to read configuration file; %s", err)
	}

	return connectToDB(dbHost, dbPort, dbName, tableName)
}

func NewTestDBCollectionConnection(t *testing.T) (*MyCollection, testcontainers.Container) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
	}

	mdb, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}

	port, _ := mdb.MappedPort(ctx, "27017")
	host, err := mdb.Host(ctx)

	if err != nil {
		t.Fatal("Container failed to start as expected", host, port)
	}

	t.Logf("MongoDB test container started on %s:%v", host, port)

	collection, err := connectToDB(host, port.Int(), "local", testTableName)
	if err != nil {
		t.Fatal("Failed to connect to DB and create new table", err)
	}

	return collection, mdb
}

func connectToDB(dbHost string, dbPort int, dbName string, tableName string) (*MyCollection, error) {
	dbURI := fmt.Sprintf("mongodb://%s:%d", dbHost, dbPort)
	//nolint: govet
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	ctx.Done()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatalf("Failed connect to db; %s", err)
	}

	collection := client.Database(dbName).Collection(tableName)

	return &MyCollection{collection}, nil
}

func readDbConfig(configName string) (dbHost string, dbPort int, dbName string, tableName string, err error) {
	viper.AddConfigPath("db/config")
	viper.SetConfigName(configName)
	viper.SetConfigType("properties")

	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	log.Printf("Using config: %s\n", viper.ConfigFileUsed())

	dbHost = viper.GetString("db.Host")
	dbPort = viper.GetInt("db.Port")
	dbName = viper.GetString("db.Name")
	tableName = viper.GetString("db.TableName")

	return
}

func (coll *MyCollection) InsertList(l *list.List) int {
	count := 0

	for e := l.Front(); e != nil; e = e.Next() {
		if item, ok := e.Value.(DBItem); ok {
			_, err := coll.collection.InsertOne(nil, bson.M{"url": item.URL, "hash": item.Encoding})
			if err == nil {
				count++
			}
		}
	}

	return count
}

func (coll *MyCollection) Insert(url string, hash string) (interface{}, error) {
	res, err := coll.collection.InsertOne(nil, bson.M{"url": url, "hash": hash})
	if res == nil {
		return nil, fmt.Errorf("failed to insert [%s, %s] to db", url, hash)
	}

	return res.InsertedID, err
}

func (coll *MyCollection) Delete(url string) (bool, error) {
	res, err := coll.collection.DeleteOne(nil, bson.M{"url": url})
	if res == nil || res.DeletedCount == 0 {
		return false, err
	}
	return true, err
}

type Result struct {
	ID   [12]byte
	URL  string
	Hash string
}

func (coll *MyCollection) GetHash(url string) (interface{}, error) {
	result := Result{}
	documents, _ := coll.collection.CountDocuments(nil, bson.M{"url": url})
	log.Println("Number of docs is", documents)

	err := coll.collection.FindOne(nil, bson.M{"url": url}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Hash, nil
}

func (coll *MyCollection) GetURL(hash string) (string, error) {
	result := Result{}

	err := coll.collection.FindOne(nil, bson.M{"hash": hash}).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
