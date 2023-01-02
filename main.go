package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type Article struct {
	Id      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type Command struct {
	Name 	string
	Value string
}

type CommandDB struct {
	ID		primitive.ObjectID `bson:"_id"`
	Name 	string
	Value string
}

type Name struct {
	Id		int
	Value string
	World string
}

type NameID struct {
	ID		primitive.ObjectID `bson:"_id"`
	Id		int
	Value string
	World string
}

// let's declare a global Articles array
// that we can then populate in our main function
// to simulate a database
var Articles []Article
var MongoDbUser string
var MongoDbPass string
var MongoDbHost string
var MongoDbPort int
var ServerPort string

/* Get MongoDB Client */
func getMongoDBClient() (*mongo.Client, error) {
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	var mongoDbUri = fmt.Sprintf("mongodb://%s:%s@%s:%d", MongoDbUser, MongoDbPass, MongoDbHost, MongoDbPort)
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoDbUri))
	if err != nil {
			log.Fatal(err)
	}
  return client, nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func createNewName(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createNewName")
	// get the body of our POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array.    
	reqBody, _ := ioutil.ReadAll(r.Body)
	var name Name 
	json.Unmarshal(reqBody, &name)

	/* Insert new document into database */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	database := client.Database("immria")
	collection := database.Collection("names")
	result, err := collection.InsertOne(ctx, name)
	if err != nil {
		log.Fatal(err)
	}

	// Print Return Statement
	fmt.Println(name)
	fmt.Println(result)
	json.NewEncoder(w).Encode(name)
}

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
	database := viper.GetString("DATABASE_NAME")
	collection := viper.GetString("COLLECTION_NAMES")
	gallery := client.Database(database).Collection(collection)
	filterFind := bson.D{{"value", value}}
	var result Name
	err = gallery.FindOne(context.TODO(), filterFind).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Sprintf("Name Id: %d",   result.Id)
	fmt.Println("Name Value: " + result.Value)
	fmt.Println("Name World: " + result.World)
	json.NewEncoder(w).Encode(result)
	fmt.Println("Successfully returned name!")
}

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
	database := viper.GetString("DATABASE_NAME")
	collection := viper.GetString("COLLECTION_NAMES")
	gallery := client.Database(database).Collection(collection)
	filterFind := bson.D{{"id", name.Id}}
	var result NameID
	err = gallery.FindOne(context.TODO(), filterFind).Decode(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.ID.Hex())
	fmt.Println(result.Value)

	/* Update Document by Id */
	id, _ := primitive.ObjectIDFromHex(result.ID.Hex())
	filterUpdate := bson.D{{"_id", id}}
	valueUpdate := bson.D{{"$set", bson.D{{"value", name.Value}}}}
	resultUpdate, err := gallery.UpdateOne(context.TODO(), filterUpdate, valueUpdate)
	if err != nil {
		panic(err)
	}
	fmt.Println(resultUpdate)

	// Return Statement
	returnStatement := fmt.Sprintf("Successfully updated name id: %d with value: %s",name.Id , name.Value )
	fmt.Fprintf(w, returnStatement)
	fmt.Println(returnStatement)
}

func returnAllNames(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllNames")

	/* Connect to our DB */
	client, err := getMongoDBClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get All Documents in Names Collection */
	database := viper.GetString("DATABASE_NAME")
	collection := viper.GetString("COLLECTION_NAMES")
	gallery := client.Database(database).Collection(collection)
	cursor, err := gallery.Find(ctx, bson.M{}, nil)
	if err != nil {
		log.Fatal(err)
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
    log.Fatal(err)
	}
	fmt.Println(results)
	json.NewEncoder(w).Encode(results)
	fmt.Println("Successfully returned name!")
}


// REST Endpoints
func handleRequests() {
	// Creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	/* Health Check */
	myRouter.HandleFunc("/", homePage)

	/* Names */
	myRouter.HandleFunc("/name", createNewName).Methods("POST")
	// myRouter.HandleFunc("/article/{id}", deleteName).Methods("DELETE")
	myRouter.HandleFunc("/name/{value}", getNameByValue)
	myRouter.HandleFunc("/name", updateExistingName).Methods("PATCH")
	myRouter.HandleFunc("/names", returnAllNames)
	log.Fatal(http.ListenAndServe(ServerPort, myRouter))
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
	// fmt.Println("EXAMPLE_PATH is\t", viper.GetString("EXAMPLE_PATH"))
	// fmt.Println("EXAMPLE_VAR is\t", viper.GetString("EXAMPLE_VAR"))

	// Start up server
	fmt.Println("Starting server...")
	handleRequests()
}
