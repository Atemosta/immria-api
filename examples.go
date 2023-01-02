func updateExistingCommand(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: updateExistingCommand")
	// get the body of our POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array.    
	reqBody, _ := ioutil.ReadAll(r.Body)
	var command Command 
	json.Unmarshal(reqBody, &command)

	// Add command to our database
	// mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:pass12345@localhost"))
	if err != nil {
			log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
			log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	/* Get Document by Filter */
	// collection := client.Database("feh").Collection("robin")
	// filter := bson.D{{"name", "calendar"}}
	// var result Command
	// err = collection.FindOne(context.TODO(), filter).Decode(&result)
	// fmt.Println(result)
	// fmt.Println(result.name)
	// fmt.Println(result.value)

	// robinResult, err := robinCollection.InsertOne(ctx, command)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Print Return Statement
	// son.NewEncoder(w).Encode(command)
	fmt.Println("Endpoint Hit: updateExistingCommand")
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createNewArticle")
	// get the body of our POST request
	// unmarshal this into a new Article struct
	// append this to our Articles array.    
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article 
	json.Unmarshal(reqBody, &article)
	// update our global Articles array to include
	// our new Article
	Articles = append(Articles, article)
	json.NewEncoder(w).Encode(article)
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: deleteArticle")
	// once again, we will need to parse the path parameters
	vars := mux.Vars(r)
	// we will need to extract the `id` of the article we
	// wish to delete
	id := vars["id"]

	// we then need to loop through all our articles
	for index, article := range Articles {
			// if our id path parameter matches one of our
			// articles
			if article.Id == id {
					// updates our Articles array to remove the 
					// article
					Articles = append(Articles[:index], Articles[index+1:]...)
			}
	}
	fmt.Fprintf(w, "Successfully delete Article Id: " + id)
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}

func main() {
	Articles = []Article{
		Article{Id: "1", Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		Article{Id: "2", Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}
}