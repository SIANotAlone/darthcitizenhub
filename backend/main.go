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
type Origin struct {
	Origin string `json:"origin"`
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

type Episode struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
}
type Episode_response struct {
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
	r.HandleFunc("/news/games/by_origin/{origin}", get_gaming_news_by_origin)
	r.HandleFunc("/news/games/by_origin/{origin}/{id}", get_gaming_news_byorigin_page_number)
	r.HandleFunc("/news/games/origins/", get_games_origins)
	r.HandleFunc("/news/games/favorite/", get_favorite_games)
	r.HandleFunc("/news/games/cancel_favorite/", cancel_favorit).Methods(http.MethodPost)
	r.HandleFunc("/news/games/changefavorite/", change_games_favorite).Methods(http.MethodPost)
	r.HandleFunc("/news/games/deleteallfavorite/", delete_all_favorite).Methods(http.MethodPost)
	r.HandleFunc("/news/serials", serials_news_page)
	r.HandleFunc("/news/films", films_news_page)

	r.HandleFunc("/episode/new/", new_episode).Methods(http.MethodPost)
	corsHandler := handlers.CORS(headers, methods, origins)(r)

	http.ListenAndServe(":80", corsHandler)
}

func new_episode(w http.ResponseWriter, r *http.Request) {

	var episode Episode
	json.NewDecoder(r.Body).Decode(&episode)

	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	req := "INSERT INTO allnews.episode (name, number, date, released) VALUES ($1, $2, CURRENT_DATE, false);"
	_, err = db.Exec(req, episode.Name, episode.Number)
	if err != nil {
		panic(err)
	}

	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Новий епізод створено"
	json.NewEncoder(w).Encode(resp)
	logger.Info("New episode has been created")

}

func get_gaming_news_byorigin_page_number(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str_id := vars["id"]
	origin := vars["origin"]
	id, err := strconv.Atoi(str_id)
	if err != nil {
		log.Fatal(err)
	}
	var gaming_news []Games_News
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	offset := (id * 20) - 20
	var query string = "SELECT * FROM allnews.games_news WHERE origin = '" + origin + "' ORDER BY id DESC LIMIT 20  OFFSET " + strconv.FormatInt(int64(offset), 10)
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
	json.NewEncoder(w).Encode(gaming_news)
	logger.Info("Games news by origin " + origin + " and page " + str_id)

}

func delete_all_favorite(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	req := `UPDATE allnews.games_news SET favorit = false WHERE favorit = true;`
	_, err = db.Exec(req)
	if err != nil {
		panic(err)
	}
	db.Close()
	var resp Favorit_response
	resp.Status = "200 OK"
	resp.Message = "Успішно видалено із закладок"
	json.NewEncoder(w).Encode(resp)
	logger.Info("Delete all favorite news")

}

func get_favorite_games(w http.ResponseWriter, r *http.Request) {
	var gaming_news []Games_News
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	query := "SELECT * FROM allnews.games_news WHERE favorit = True ORDER BY id DESC"
	// fmt.Println(query)
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
	json.NewEncoder(w).Encode(gaming_news)
	logger.Info("Get favorite gaming news")

}

func get_gaming_news_by_origin(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	origin := vars["origin"]

	var gaming_news []Games_News
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	query := "SELECT * FROM allnews.games_news WHERE origin = '" + origin + "' ORDER BY id DESC LIMIT 20"
	// fmt.Println(query)
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
	json.NewEncoder(w).Encode(gaming_news)
	logger.Info("Get last news by origin page 1")
}

func get_games_origins(w http.ResponseWriter, r *http.Request) {
	var origins []Origin
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("SELECT DISTINCT origin FROM allnews.games_news")
	for rows.Next() {
		f := Origin{}
		err = rows.Scan(&f.Origin)
		if err != nil {
			log.Fatal(err)
		}
		origins = append(origins, f)

	}
	db.Close()
	json.NewEncoder(w).Encode(origins)
	logger.Info("Get all origins")
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
