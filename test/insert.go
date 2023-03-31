package main

import (
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/milvus"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var collectionName string = "test256"
var dimension int64 = 256
var indexFileSize int64 = 1024
var metricType int32 = int32(milvus.L2)
var nq int64 = 100
var nprobe int64 = 64
var nb int64 = 10000
var topk int64 = 100
var nlist int64 = 16384

var delete_id_array = []int64{
	1677048628230439977,
	1677048628230439978,
	1677048628230439979,
	1677048628230439980,
	1677048628230439981,
	1677048628230439982,
	1677048628230439983,
	1677048628230439984,
	1677048628230439985,
	1677048628230439986,
}

func main() {
	address := "10.15.9.78"
	port := "19530"

	//getIndex(address, port)
	insert(address, port)
}

func dropAllCollection(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	collections, _, err := client.ListCollections(ctx)

	for _, collection := range collections {
		client.DropCollection(ctx, collection)
	}
}

func getIndex(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	info, _, err := client.GetIndexInfo(ctx, collectionName)
	fmt.Println(info)
}

func createIndex(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	_, err = client.CreateIndex(ctx, &milvus.IndexParam{
		CollectionName: collectionName,
		IndexType:      milvus.IVFFLAT,
	})
	if err != nil {
		return
	}

	fmt.Println("create Index success")
}

func rowCounts(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	entities, _, err := client.CountEntities(ctx, collectionName)
	if err != nil {
		return
	}

	fmt.Println(entities)
}

func dropSpecCollection(address string, port string, colName string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	_, err = client.DropCollection(ctx, colName)
	if err != nil {
		return
	}
}

func connect(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	//Client version
	println("Client version: " + client.GetClientVersion(ctx))

	if client.IsConnected(ctx) == false {
		println("client: not connected: ")
		return
	}
	println("Server status: connected")

	//Get server version
	var version string
	var status milvus.Status
	version, status, err = client.ServerVersion(ctx)
	if err != nil {
		println("Cmd rpc failed: " + err.Error())
	}
	if !status.Ok() {
		println("Get server version failed: " + status.GetMessage())
		return
	}
	println("Server version: " + version)
}

func insert(address string, port string) {
	var i, j int64
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	// 创建集合参数
	collectionParam := milvus.CollectionParam{collectionName, dimension, indexFileSize, metricType}
	hasCollection, status, err := client.HasCollection(ctx, collectionName)
	if err != nil {
		println("HasColletcion rpc failed")
	}
	if hasCollection == false {
		status, err = client.CreateCollection(ctx, collectionParam)
		if err != nil {
			println("CreateCollection rpc failed: " + err.Error())
			return
		}
		if !status.Ok() {
			println("Create collection failed: " + status.GetMessage())
			return
		}
		println("Create collection " + collectionName + " success")
	}

	//test insert vectors
	records := make([]milvus.Entity, nb)
	recordArray := make([][]float32, nb)
	for i = 0; i < nb; i++ {
		recordArray[i] = make([]float32, dimension)
		for j = 0; j < dimension; j++ {
			recordArray[i][j] = float32(i % (j + 1))
		}
		records[i].FloatData = recordArray[i]
	}

	// idArray
	idArray := make([]int64, nb)
	for i = 0; i < nb; i++ {
		idArray[i] = i + 1
	}
	println("Begin to insert data")
	insertParam := milvus.InsertParam{collectionName, "", records, idArray}
	id_array, status, err := client.Insert(ctx, &insertParam)
	if err != nil {
		println("Insert rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Insert vector failed: " + status.GetMessage())
		return
	}
	if len(id_array) != int(nb) {
		println("ERROR: return id array is null")
	}
	println("Insert vectors success!")

	time.Sleep(3 * time.Second)

	//test describe collection
	collectionParam, status, err = client.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		println("DescribeCollection rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("DescribeCollection rpc failed: " + status.GetMessage())
		return
	}
	println("CollectionName:" + collectionParam.CollectionName + "----Dimension:" + strconv.Itoa(int(collectionParam.Dimension)) +
		"----IndexFileSize:" + strconv.Itoa(int(collectionParam.IndexFileSize)))
	println("Id_Array:")
	for i := 0; i < len(id_array); i++ {
		fmt.Println(id_array[i])
	}

}

func delete(address string, port string) {
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	stats, err := client.DeleteEntityByID(ctx, collectionName, "", delete_id_array)
	if err != nil {
		println("DescribeCollection rpc failed: " + err.Error())
		return
	}
	if !stats.Ok() {
		println("Create index failed: " + stats.GetMessage())
		return
	}

	fmt.Println("Delete Success")
}

func example(address string, port string) {
	var i, j int64
	connectParam := milvus.ConnectParam{IPAddress: address, Port: port}
	ctx := context.TODO()
	client, err := milvus.NewMilvusClient(ctx, connectParam)
	if err != nil {
		log.Fatalf("Client connect failed: %v", err)
	}

	//Client version
	println("Client version: " + client.GetClientVersion(ctx))

	if client.IsConnected(ctx) == false {
		println("client: not connected: ")
		return
	}
	println("Server status: connected")

	//Get server version
	var version string
	var status milvus.Status
	version, status, err = client.ServerVersion(ctx)
	if err != nil {
		println("Cmd rpc failed: " + err.Error())
	}
	if !status.Ok() {
		println("Get server version failed: " + status.GetMessage())
		return
	}
	println("Server version: " + version)

	//test create collection
	collectionParam := milvus.CollectionParam{collectionName, dimension, indexFileSize, metricType}
	var hasCollection bool
	//hasCollection, status, err = client.HasCollection(collectionName)
	if err != nil {
		println("HasCollection rpc failed: " + err.Error())
	}
	if hasCollection == false {
		status, err = client.CreateCollection(ctx, collectionParam)
		if err != nil {
			println("CreateCollection rpc failed: " + err.Error())
			return
		}
		if !status.Ok() {
			println("Create collection failed: " + status.GetMessage())
			return
		}
		println("Create collection " + collectionName + " success")
	}

	hasCollection, status, err = client.HasCollection(ctx, collectionName)
	if err != nil {
		println("HasCollection rpc failed: " + err.Error())
		return
	}
	if hasCollection == false {
		println("Create collection failed: " + status.GetMessage())
		return
	}
	println("Collection: " + collectionName + " exist")

	println("**************************************************")

	//test show collections
	var collections []string
	collections, status, err = client.ListCollections(ctx)
	if err != nil {
		println("ShowCollections rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Show collections failed: " + status.GetMessage())
		return
	}
	println("ShowCollections: ")
	for i = 0; i < int64(len(collections)); i++ {
		println(" - " + collections[i])
	}

	//test insert vectors
	records := make([]milvus.Entity, nb)
	recordArray := make([][]float32, nb)
	for i = 0; i < nb; i++ {
		recordArray[i] = make([]float32, dimension)
		for j = 0; j < dimension; j++ {
			recordArray[i][j] = float32(i % (j + 1))
		}
		records[i].FloatData = recordArray[i]
	}
	insertParam := milvus.InsertParam{collectionName, "", records, nil}
	id_array, status, err := client.Insert(ctx, &insertParam)
	if err != nil {
		println("Insert rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Insert vector failed: " + status.GetMessage())
		return
	}
	if len(id_array) != int(nb) {
		println("ERROR: return id array is null")
	}
	println("Insert vectors success!")

	time.Sleep(3 * time.Second)

	//test describe collection
	collectionParam, status, err = client.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		println("DescribeCollection rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Create index failed: " + status.GetMessage())
		return
	}
	println("CollectionName:" + collectionParam.CollectionName + "----Dimension:" + strconv.Itoa(int(collectionParam.Dimension)) +
		"----IndexFileSize:" + strconv.Itoa(int(collectionParam.IndexFileSize)))

	//Construct query vectors
	queryRecords := make([]milvus.Entity, nq)
	queryVectors := make([][]float32, nq)
	for i = 0; i < nq; i++ {
		queryVectors[i] = make([]float32, dimension)
		for j = 0; j < dimension; j++ {
			queryVectors[i][j] = float32(rand.Float64())
		}
		queryRecords[i].FloatData = queryVectors[i]
	}

	println("**************************************************")

	//Search without create index
	var topkQueryResult milvus.TopkQueryResult
	extraParams := "{\"nprobe\" : 32}"
	searchParam := milvus.SearchParam{collectionName, queryRecords, topk, nil, extraParams}
	topkQueryResult, status, err = client.Search(ctx, searchParam)
	if err != nil {
		println("Search rpc failed: " + err.Error())
	}
	println("Search without index results: ")
	for i = 0; i < 10; i++ {
		print(topkQueryResult.QueryResultList[i].Ids[0])
		print("        ")
		println(topkQueryResult.QueryResultList[i].Distances[0])
	}

	println("**************************************************")

	//test CountCollection
	var collectionCount int64
	collectionCount, status, err = client.CountEntities(ctx, collectionName)
	if err != nil {
		println("CountCollection rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Get collection count failed: " + status.GetMessage())
		return
	}
	println("Collection count:" + strconv.Itoa(int(collectionCount)))

	//Create index
	println("Start create index...")
	extraParams = "{\"nlist\" : 16384}"
	indexParam := milvus.IndexParam{collectionName, milvus.IVFFLAT, extraParams}
	status, err = client.CreateIndex(ctx, &indexParam)
	if err != nil {
		println("CreateIndex rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Create index failed: " + status.GetMessage())
		return
	}
	println("Create index success!")

	//Describe index
	indexParam, status, err = client.GetIndexInfo(ctx, collectionName)
	if err != nil {
		println("DescribeIndex rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Describe index failed: " + status.GetMessage())
	}
	println(indexParam.CollectionName + "----index type:" + strconv.Itoa(int(indexParam.IndexType)))

	//Preload collection
	loadCollectionParam := milvus.LoadCollectionParam{collectionName, nil}
	status, err = client.LoadCollection(ctx, loadCollectionParam)
	if err != nil {
		println("PreloadCollection rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println(status.GetMessage())
	}
	println("Preload collection success")

	println("**************************************************")

	//Search with IVFSQ8 index
	extraParams = "{\"nprobe\" : 32}"
	searchParam = milvus.SearchParam{collectionName, queryRecords, topk, nil, extraParams}
	topkQueryResult, status, err = client.Search(ctx, searchParam)
	if err != nil {
		println("Search rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Search vectors failed: " + status.GetMessage())
	}
	println("Search with index results: ")
	for i = 0; i < 10; i++ {
		print(topkQueryResult.QueryResultList[i].Ids[0])
		print("        ")
		println(topkQueryResult.QueryResultList[i].Distances[0])
	}

	println("**************************************************")

	//Drop index
	status, err = client.DropIndex(ctx, collectionName)
	if err != nil {
		println("DropIndex rpc failed: " + err.Error())
		return
	}
	if !status.Ok() {
		println("Drop index failed: " + status.GetMessage())
	}

	//Drop collection
	status, err = client.DropCollection(ctx, collectionName)
	hasCollection, status1, err := client.HasCollection(ctx, collectionName)
	if !status.Ok() || !status1.Ok() || hasCollection == true {
		println("Drop collection failed: " + status.GetMessage())
		return
	}
	println("Drop collection " + collectionName + " success!")

	//GetConfig
	var configInfo string
	configInfo, status, err = client.GetConfig(ctx, "*")
	if !status.Ok() {
		println("Get config failed: " + status.GetMessage())
	}
	println("config: ")
	println(configInfo)

	//Disconnect
	err = client.Disconnect(ctx)
	if err != nil {
		println("Disconnect failed!")
		return
	}
	println("Client disconnect server success!")

	//Server status
	var serverStatus string
	serverStatus, status, err = client.ServerStatus(ctx)
	if !status.Ok() {
		println("Get server status failed: " + status.GetMessage())
	}
	println("Server status: " + serverStatus)

}