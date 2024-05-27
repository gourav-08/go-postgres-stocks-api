package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gourav-08/go-postgres-stocks-api/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type response struct {
	Id      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("There was an error in loading .env")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to postgres!")
	return db
}

func CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock models.Stock

	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable to decode request body. %v", err)
	}

	insertId := insertStock(stock)

	res := response{
		Id:      insertId,
		Message: "Stock created succesfully",
	}

	json.NewEncoder(w).Encode(res)
}

func GetStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to decode request body. %v", err)
	}

	stock, err := getStock(int64(id))
	if err != nil {
		log.Fatalf("Unable to get stock. %v", err)
	}

	json.NewEncoder(w).Encode(stock)
}

func GetAllStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := getAllStocks()
	if err != nil {
		log.Fatalf("Unable to get all stocks. %v", err)
	}

	json.NewEncoder(w).Encode(stocks)
}

func UpdateStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to decode request body. %v", err)
	}

	var stock models.Stock
	err = json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("Unable decode request body. %v", err)
	}

	updateRows := updateStock(int64(id), stock)
	msg := fmt.Sprintf("Stock updated succesfully. Total records affected: %v", updateRows)
	res := response{
		Id:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func DeleteStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("Unable to parse Int. %v", err)
	}
	deletedRows := deleteStock(int64(id))
	msg := fmt.Sprintf("Succesfully deleted the stock. Total records affected: %v", deletedRows)
	res := response{
		Id:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func insertStock(stock models.Stock) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `INSERT INTO stocks(name, price, company) VALUES ($1, $2, $3) RETURNING stockid`
	var id int64

	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute query. %v", err)
	}

	fmt.Printf("Inserted a single record. %v", id)
	return id
}

func getStock(id int64) (models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stock models.Stock
	sqlStatement := `SELECT * FROM stocks WHERE stockid = $1`

	err := db.QueryRow(sqlStatement, id).Scan(&stock.StockId, &stock.Name, &stock.Price, &stock.Company)
	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return stock, nil
	case nil:
		return stock, nil
	default:
		log.Fatalf("Unable to get stock. %v", err)
	}

	return stock, nil
}

func getAllStocks() ([]models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stocks []models.Stock
	sqlStatement := `SELECT * FROM stocks`
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatalf("Unable to execute query. %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var stock models.Stock

		err = rows.Scan(&stock.StockId, &stock.Name, &stock.Price, &stock.Company)
		if err != nil {
			log.Fatalf("Unable to parse row %v", err)
		}
		stocks = append(stocks, stock)
	}

	return stocks, err
}

func updateStock(id int64, stock models.Stock) int64 {
	db := createConnection()
	defer db.Close()

	sqlStatement := `UPDATE stocks SET name = $2, price = $3, company = $4 WHERE stockid = $1`
	result, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)
	if err != nil {
		log.Fatalf("Unable to update row %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking affected rows %v", err)
	}
	return rowsAffected
}

func deleteStock(id int64) int64 {
	db := createConnection()
	defer db.Close()

	sqlStatement := `DELETE from stocks WHERE stockid = $1`
	result, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatalf("Unable to delete row %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking affected rows %v", err)
	}
	return rowsAffected
}
