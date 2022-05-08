package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"

	_ "github.com/lib/pq"
)

type Person struct {
	UUID     string `json:uuid`
	Name     string `json:name`
	NickName string `json:nickname`
}

type Response struct {
	Status     string   `json:"status"`
	StatucCode int      `json:"statusCode"`
	Data       []Person `json:"data"`
}

type ErrorResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func openConnection() *sql.DB {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	newPort, err := strconv.Atoi(dbPort)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, newPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func GETPeople(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println("Sejatinya masuk ga GETPeople")

	db := openConnection()

	rows, err := db.Query("SELECT * FROM Person")

	var people []Person

	for rows.Next() {
		var person Person
		rows.Scan(&person.Name, &person.NickName, &person.UUID)

		people = append(people, person)
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Fatal(err)
	} else {
		if len(people) > 0 {
			json.NewEncoder(w).Encode(&Response{"success", 200, people})
		} else {
			json.NewEncoder(w).Encode(&ErrorResponse{"error", 404, "Data Not Found"})
		}
	}

	defer rows.Close()
	defer db.Close()
}

func POSTPerson(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	db := openConnection()

	var p Person
	err := json.NewDecoder(r.Body).Decode(&p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var newUUID string

	newUUID = uuid.New().String()

	fmt.Println(newUUID)

	sqlStatement := `INSERT INTO person (name, nickname, uuid) VALUES ($1, $2, $3)`
	_, err = db.Exec(sqlStatement, p.Name, p.NickName, newUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)

	allDatas, _ := db.Query("SELECT * FROM Person")

	var people []Person

	for allDatas.Next() {
		var person Person

		allDatas.Scan(&person.Name, &person.NickName, &person.UUID)
		people = append(people, person)
	}

	json.NewEncoder(w).Encode(&Response{"success", 201, people})

	defer db.Close()
}

func PersonByUUID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uuid := ps.ByName("id")

	db := openConnection()

	rows, err := db.Query("SELECT * FROM Person WHERE uuid=$1", uuid)

	if err != nil {
		log.Fatal(err)
	}

	var people []Person

	fmt.Println(rows)

	for rows.Next() {
		var person Person
		rows.Scan(&person.Name, &person.NickName, &person.UUID)
		people = append(people, person)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(&Response{"success", 200, people})

	defer rows.Close()
	defer db.Close()
}

func DeleteByUUID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uuid := ps.ByName("id")

	db := openConnection()

	rows, err := db.Query("DELETE FROM Person WHERE uuid=$1", uuid)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)

	allDatas, _ := db.Query("SELECT * FROM Person")

	var people []Person

	for allDatas.Next() {
		var person Person

		allDatas.Scan(&person.Name, &person.NickName, &person.UUID)
		people = append(people, person)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(&Response{"success", 200, people})

	defer rows.Close()
	defer db.Close()

}

func UpdatePersonbyUUID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uuid := ps.ByName("id")
	db := openConnection()

	var p Person
	err := json.NewDecoder(r.Body).Decode(&p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	sqlStatement := `UPDATE Person SET name=$1, nickname=$2, uuid=$3 WHERE uuid=$4`
	_, err = db.Exec(sqlStatement, p.Name, p.NickName, uuid, uuid)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)

	allDatas, _ := db.Query("SELECT * FROM Person")

	var people []Person

	for allDatas.Next() {
		var person Person

		allDatas.Scan(&person.Name, &person.NickName, &person.UUID)
		people = append(people, person)
	}

	json.NewEncoder(w).Encode(&Response{"success", 200, people})

	defer db.Close()

}

func main() {

	router := httprouter.New()

	router.GET("/", GETPeople)
	router.POST("/", POSTPerson)
	router.GET("/person/:id", PersonByUUID)
	router.DELETE("/person/:id", DeleteByUUID)
	router.PUT("/person/:id", UpdatePersonbyUUID)

	err := http.ListenAndServe(":8080", router)
	log.Fatal(err)
}
