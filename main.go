package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*** Structs ***/
type Article struct {
	Id      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type Name struct {
	TokenId	int			`json:"tokenId"`
	Value 	string 	`json:"value"`
	World 	string 	`json:"world"`
}

type NameID struct {
	ID		primitive.ObjectID `bson:"_id"`
	Id		int
	Value string
	World string
}

type World struct {
	Desc		string `json:"desc"` 
	Image		string `json:"image"`
	Owner 	string `json:"owner"`
	Title		string `json:"title"`
	Type 		string `json:"type"`
	Status 	string `json:"status"`
}

type WorldID struct {
	// World
	Desc		string 	`json:"desc"` 
	Image		string 	`json:"image"`
	Owner 	string 	`json:"owner"`
	Title		string 	`json:"title"`
	Type 		string 	`json:"type"`
	Status 	string 	`json:"status"`
	ID	primitive.ObjectID `bson:"_id"`
}

// Global Variables
var Articles []Article
var CharacterLimit			int
var DatabaseName 				string
var DescriptionLimit		int
var CollectionNames 		string 
var CollectionWorlds 		string
var MongoDbUser 				string
var MongoDbPass 				string
var MongoDbHost 				string
var MongoDbPort 				int
var ServerPort 					string

/*--- MongoDB Helper Funcs ---*/
// Get MongoDB Client
func getMongoDBClient() (*mongo.Client, error) {
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	var mongoDbUri = fmt.Sprintf("mongodb://%s:%s@%s:%d", MongoDbUser, MongoDbPass, MongoDbHost, MongoDbPort)
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDbUri))
	if err != nil {
			log.Fatal(err)
	}
  return client, nil
}

/*--- REST Calls ---*/
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

// POST "/name"
func createNewName(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createNewName")
	/* Umarshal POST body into Name struct */
	reqBody, _ := ioutil.ReadAll(r.Body)
	var name Name 
	json.Unmarshal(reqBody, &name)
	CharacterLimit := viper.GetInt("CHARACTER_LIMIT")  // first 20 characters 
	if len(name.Value) > 20 { name.Value = name.Value[: + CharacterLimit] }
	if len(name.World) > 20 { name.World = name.World[: + CharacterLimit] }

	/* Connect to database */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	fmt.Println(DatabaseName)
	database := client.Database(DatabaseName)
	collection := database.Collection(CollectionNames)

	// Get Token Id
	count, err := collection.CountDocuments(context.TODO(), bson.M{}) 
	if err != nil { 
		 // handle error 
	} 
	fmt.Println("New TokenId: ")
	fmt.Println(count)

	// Insert New Doc
	name.TokenId = int(count)
	resultInsert, err := collection.InsertOne(ctx, name)
	if err != nil {
		log.Fatal(err)
	}

	// Print Return Statement
	fmt.Println(name)
	fmt.Println(resultInsert)
	json.NewEncoder(w).Encode(name)
}

// POST "/world"
func createWorld(w http.ResponseWriter, r *http.Request) {
	// Enter Func
	fmt.Println("Endpoint Hit: createWorld")
	
	/* Umarshal POST body into struct */
	reqBody, _ := ioutil.ReadAll(r.Body)
	var world World 
	json.Unmarshal(reqBody, &world)
	if len(world.Desc) > DescriptionLimit { world.Desc = world.Desc[: + DescriptionLimit] }
	if len(world.Title) > CharacterLimit { world.Title = world.Title[: + CharacterLimit] }

	/* Connect to database */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database(DatabaseName).Collection(CollectionWorlds)

	// Insert New Doc
	result, err := collection.InsertOne(ctx, world)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	// Print Return Statement
	var worldID WorldID
	copier.Copy(&worldID, &world)
	id := result.InsertedID.(primitive.ObjectID) //.Hex()
	worldID.ID = id
	fmt.Println(worldID)
	json.NewEncoder(w).Encode(worldID)
}

/* DELETE /name/{tokenid} */
func deleteNamebyTokenId(w http.ResponseWriter, r *http.Request){
	fmt.Println("Endpoint Hit: deleteNamebyTokenId")
	vars := mux.Vars(r)
	tokenIdStr := vars["tokenid"]
	tokenid, err := strconv.Atoi(tokenIdStr)
	fmt.Println(tokenIdStr)

	/* Connect to DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	collection := client.Database(DatabaseName).Collection(CollectionNames)
	filter := bson.D{{"tokenid", tokenid}} 
	deleteResult, err := collection.DeleteMany(context.TODO(), filter) 
	if err != nil { log.Fatal(err) } 
	msg := fmt.Sprintf("Deleted %v documents in the people collection", deleteResult.DeletedCount)
	fmt.Fprintf(w, msg)
	fmt.Println(msg)
}

// DELETE "/world/{id}"
func deleteWorld(w http.ResponseWriter, r *http.Request){
	// Start Func
	fmt.Println("Endpoint Hit: deleteWorld")

	// Get Path Params
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"]) // ObjectID

	/* Connect to DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	collection := client.Database(DatabaseName).Collection(CollectionWorlds)
	filter := bson.D{{"_id", id}} 
	result, err := collection.DeleteMany(context.TODO(), filter) 
	if err != nil { log.Fatal(err) } 

	// Return Result
	msg := fmt.Sprintf("Deleted %v documents in the worlds collection", result.DeletedCount)
	fmt.Println(msg)
	fmt.Fprintf(w, msg)
}

/* GET /name/tokenid/{tokenid} */
func getNameByTokenId(w http.ResponseWriter, r *http.Request){
	fmt.Println("Endpoint Hit: getNameById")
	vars := mux.Vars(r)
	tokenIdStr := vars["tokenid"]
	tokenid, err := strconv.Atoi(tokenIdStr)
	fmt.Println(tokenid)

	/* Connect to our DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	collection := client.Database(DatabaseName).Collection(CollectionNames)
	filterFind := bson.D{{"tokenid", tokenid}}
	var result Name
	err = collection.FindOne(context.TODO(), filterFind).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Name Id: %d",   result.TokenId)
	fmt.Println("Name Value: " + result.Value)
	fmt.Println("Name World: " + result.World)
	json.NewEncoder(w).Encode(result)
	fmt.Println("Successfully returned name!")
}

/* GET /name/value/{value} */
func getNameByValue(w http.ResponseWriter, r *http.Request){
	fmt.Println("Endpoint Hit: getNameByValue")
	vars := mux.Vars(r)
	value := vars["value"]

	/* Connect to our DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	collection := client.Database(DatabaseName).Collection(CollectionNames)
	filterFind := bson.D{{"value", value}}
	var result Name
	err = collection.FindOne(context.TODO(), filterFind).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Name Id: %d",   result.TokenId)
	fmt.Println("Name Value: " + result.Value)
	fmt.Println("Name World: " + result.World)
	json.NewEncoder(w).Encode(result)
	fmt.Println("Successfully returned name!")
}

/* GET /world/{id} */
func getWorld(w http.ResponseWriter, r *http.Request){
	// Start Func
	fmt.Println("Endpoint Hit: getWorld")

	// Get Path Params
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"]) // ObjectID
	
	// Connect to our DB
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	collection := client.Database(DatabaseName).Collection(CollectionWorlds)
	filter := bson.M{"_id": id}
	var result WorldID
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)

	// Return Results
	json.NewEncoder(w).Encode(result)
	fmt.Println("Successfully returned world!")
}

func getWorlds(w http.ResponseWriter, r *http.Request) {
	// Start Func
	fmt.Println("Endpoint Hit: getWorlds")

	// Get Query Params
	params := r.URL.Query()
	fmt.Println(params)
	var status string
	if _, ok := params["status"]; ok { 
		status = params["status"][0]
	} else {
		status = "public" 
	}

	/* Connect to our DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get All Documents in Collection By Filter */
	collection := client.Database(DatabaseName).Collection(CollectionWorlds)
	filter := bson.D{{"status", status}}
	cursor, err := collection.Find(ctx, filter, nil)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	fmt.Println(results)

	// Return Results
	json.NewEncoder(w).Encode(results)
	fmt.Println("Successfully returned worlds from status %s!", status)
}

func returnAllNamesByWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllNamesByWorld")
	vars := mux.Vars(r)
	world := vars["world"]

	/* Connect to our DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get All Documents in Names Collection */
	collection := client.Database(DatabaseName).Collection(CollectionNames)
	// sort := bson.M{"document.value": 1} // Sort by document value, 1 is ascending and -1 is descending
	filterFind := bson.D{{"world", world}}
	cursor, err := collection.Find(ctx, filterFind, nil)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
    log.Fatal(err)
	}
	fmt.Println(results)
	json.NewEncoder(w).Encode(results)
	fmt.Println("Successfully returned names from world %s!", world)
}

// PUT /name
func updateExistingName(w http.ResponseWriter, r *http.Request){
	fmt.Println("Endpoint Hit: updateExistingName")
	// Unmarshal post request
	reqBody, _ := ioutil.ReadAll(r.Body)
	var name Name 
	json.Unmarshal(reqBody, &name)

	/* Get Document by Filter */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database(DatabaseName).Collection(CollectionNames)
	filterFind := bson.D{{"tokenid", name.TokenId}}
	var result NameID
	err = collection.FindOne(context.TODO(), filterFind).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.ID.Hex())
	fmt.Println(result.Value)

	/* Update Document by Id */
	id, _ := primitive.ObjectIDFromHex(result.ID.Hex())
	filterUpdate := bson.D{{"_id", id}}
	valueUpdate := bson.D{{"$set", bson.D{{"value", name.Value}}}}
	resultUpdate, err := collection.UpdateOne(context.TODO(), filterUpdate, valueUpdate)
	if err != nil {
		panic(err)
	}
	fmt.Println(resultUpdate)

	// Return Statement
	filterFind = bson.D{{"_id", id}}
	var updatedName Name
	err = collection.FindOne(context.TODO(), filterFind).Decode(&updatedName)
	if err != nil {
		panic(err)
	}
	json.NewEncoder(w).Encode(updatedName)
	fmt.Println(updatedName)
}

// PUT "/world"
func updateWorld(w http.ResponseWriter, r *http.Request){
	// Start Func
	fmt.Println("Endpoint Hit: updateWorld")

	// Unmarshal Request Body
	reqBody, _ := ioutil.ReadAll(r.Body)
	var world WorldID 
	json.Unmarshal(reqBody, &world)
	fmt.Println("---world")
	fmt.Println(world)

	// Connect to DB
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Create Updated Document
	update := bson.D{
		{"$set", 
			bson.D{
				{"desc", 		world.Desc},
				{"image", 	world.Image},
				{"owner", 	world.Owner},
				{"title", 	world.Title},
				{"type", 		world.Type},
				{"status", 	world.Status},
			},
		},
	}
	fmt.Println("---update")
	fmt.Println(update)

	// Update Document by Id
	collection := client.Database(DatabaseName).Collection(CollectionWorlds)
	filter := bson.M{"_id": world.ID}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		panic(err)
	}
	fmt.Println("---result")
	fmt.Println(result)

	// Return Statement
	if result.ModifiedCount > 0 {
		json.NewEncoder(w).Encode(world)
		fmt.Println(world)
	} else {
		fmt.Fprintf(w, "Unable to update record")
		fmt.Println("Unable to update record")
	}
}

// REST Endpoints
func handleRequests() {
	// Creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	/* Health Check */
	myRouter.HandleFunc("/", homePage)

	/* Names */
	myRouter.HandleFunc("/names/{world}", returnAllNamesByWorld)
	myRouter.HandleFunc("/name/tokenid/{tokenid}", getNameByTokenId)
	myRouter.HandleFunc("/name/value/{value}", getNameByValue)
	myRouter.HandleFunc("/name/{tokenid}", deleteNamebyTokenId).Methods("DELETE")
	myRouter.HandleFunc("/name", createNewName).Methods("POST")
	myRouter.HandleFunc("/name", updateExistingName).Methods("PUT")

	/* Worlds */
	myRouter.HandleFunc("/worlds", getWorlds).Methods("GET") // By Status
	myRouter.HandleFunc("/world", createWorld).Methods("POST")
	myRouter.HandleFunc("/world", updateWorld).Methods("PUT")
	myRouter.HandleFunc("/world/{id}", getWorld).Methods("GET") // By Object ID
	myRouter.HandleFunc("/world/{id}", deleteWorld).Methods("DELETE") // By Object ID


	/* Set up CORS */
	// c := cors.New(cors.Options{
	// 	AllowedOrigins: []string{"*"},
	// 	AllowedMethods: []string{ http.MethodGet, http.MethodPost, http.MethodPut},
	// 	AllowedHeaders:   []string{"*"},
	// 	AllowCredentials: false,
	// })
	// handler := c.Handler(myRouter)
	handler := cors.AllowAll().Handler(myRouter)
	log.Fatal(http.ListenAndServe(ServerPort, handler))
}

// Main Function
func main() {
	// Set up config
	fmt.Println("Setting up config...")
	viper.SetConfigName("config") // Set the file name of the configurations file
	viper.AddConfigPath(".") // Set the path to look for the configurations file
	viper.AutomaticEnv() // Enable VIPER to read Environment Variables
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Reading variables without using the model
	fmt.Println("Reading variables without using the model..")
	ServerPort = viper.GetString("server.port")
	MongoDbUser = viper.GetString("mongodb.user")
	MongoDbPass = viper.GetString("mongodb.pass")
	MongoDbHost = viper.GetString("mongodb.host")
	MongoDbPort = viper.GetInt("mongodb.port")
	DescriptionLimit = viper.GetInt("DESCRIPTION_LIMIT")
	CharacterLimit = viper.GetInt("CHARACTER_LIMIT")
	CollectionNames = viper.GetString("COLLECTION_NAMES") 
	CollectionWorlds = viper.GetString("COLLECTION_WORLDS")
	DatabaseName = viper.GetString("DATABASE_NAME") 

	// Start up server
	fmt.Println("Starting server...")
	handleRequests()
}
