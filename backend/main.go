package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	logger "github.com/sirupsen/logrus"
)

type Games_News struct {
	Id      int     `json:"id"`
	Title   string  `json:"title"`
	Short   string  `json:"short"`
	Origin  string  `json:"origin"`
	Url     string  `json:"url"`
	Preview string  `json:"preview"`
	Time    float64 `json:"time"`
	Favorit bool    `json:"favorit"`
}
type Favorit struct {
	Id      int  `json:"id"`
	Checked bool `json:"checked"`
}
type Favorit_response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Cancel_favorit_response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func init() {
	// Log as JSON instead of the default ASCII formatter.

	logger.SetFormatter(&logger.TextFormatter{
		DisableColors:   false,
		ForceColors:     true,
		TimestampFormat: "02-01-2006 15:04:05",
		FullTimestamp:   true})

}

func main() {

	fmt.Println("Server is starting...")
	r := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"Content-Type"})
	methods := handlers.AllowedMethods([]string{"POST"})
	origins := handlers.AllowedOrigins([]string{"*"})
	r.HandleFunc("/news/games", gaming_news_page)
	r.HandleFunc("/news/games/{id}", gaming_news_page_number)
	r.HandleFunc("/news/games/cancel_favorite/", cancel_favorit).Methods(http.MethodPost)
	r.HandleFunc("/news/games/changefavorite/", change_games_favorite).Methods(http.MethodPost)
	r.HandleFunc("/news/serials", serials_news_page)
	r.HandleFunc("/news/films", films_news_page)
	corsHandler := handlers.CORS(headers, methods, origins)(r)
	http.ListenAndServe(":80", corsHandler)
}

func cancel_favorit(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	req := `UPDATE allnews.games_news SET favorit = false WHERE favorit = true`
	_, err = db.Exec(req)
	if err != nil {
		panic(err)
	}
	db.Close()
	var resp Cancel_favorit_response
	resp.Status = "200 OK"
	resp.Message = "Успішно змінено в базі даних"
	json.NewEncoder(w).Encode(resp)
	logger.Info("Cancel all favorit games news.")
}

func change_games_favorite(w http.ResponseWriter, r *http.Request) {

	var news_id Favorit
	json.NewDecoder(r.Body).Decode(&news_id)
	// fmt.Println("id: ", news_id.Id)
	// fmt.Println("checked:", news_id.Checked)

	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	req := `UPDATE allnews.games_news SET favorit = $1 WHERE id = $2;`
	_, err = db.Exec(req, news_id.Checked, news_id.Id)
	if err != nil {
		panic(err)
	}
	db.Close()
	var resp Favorit_response
	resp.Status = "200 OK"
	resp.Message = "Успішно записано в базу даних"
	json.NewEncoder(w).Encode(resp)
	logger.Info("Change favorite ", news_id.Id, " to ", news_id.Checked, " status ", resp.Status)

}
func gaming_news_page_number(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	str_id := vars["id"]
	id, err := strconv.Atoi(str_id)

	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var gaming_news_list = get_games_news_by_page(id)
	json.NewEncoder(w).Encode(gaming_news_list)
	logger.Info("Games news by page " + str_id + " handler")
}

func serials_news_page(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("serial news page")
}
func films_news_page(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("films news page")
}

func gaming_news_page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var gaming_news_list = get_games_news()
	logger.Info("latest games news handler")
	json.NewEncoder(w).Encode(gaming_news_list)
}

func get_games_news() []Games_News {
	var gaming_news []Games_News
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM allnews.games_news ORDER BY id DESC LIMIT 20")
	for rows.Next() {
		f := Games_News{}
		err = rows.Scan(&f.Id, &f.Title, &f.Short, &f.Origin, &f.Url, &f.Preview, &f.Time, &f.Favorit)
		if err != nil {
			log.Fatal(err)
		}
		gaming_news = append(gaming_news, f)

	}
	db.Close()
	return gaming_news
}
func get_games_news_by_page(id int) []Games_News {
	var gaming_news []Games_News
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	offset := (id * 20) - 20
	var query string = "SELECT * FROM allnews.games_news ORDER BY id DESC LIMIT 20  OFFSET " + strconv.FormatInt(int64(offset), 10)
	rows, err := db.Query(query)
	for rows.Next() {
		f := Games_News{}
		err = rows.Scan(&f.Id, &f.Title, &f.Short, &f.Origin, &f.Url, &f.Preview, &f.Time, &f.Favorit)
		if err != nil {
			log.Fatal(err)
		}
		gaming_news = append(gaming_news, f)

	}
	db.Close()

	return gaming_news
}
