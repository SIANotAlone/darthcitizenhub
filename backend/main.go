package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/jung-kurt/gofpdf"
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

type Delete_episode struct {
	ID int `json:"id"`
}
type Episode_response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type Add_news struct {
	Episode_ID int `json:"episode_id"`
}
type Episode_db struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Number      int       `json:"number"`
	Date        time.Time `json:"date"`
	Released    bool      `json:"released"`
	Intro       *string   `json:"intro"`
	Ending      *string   `json:"ending"`
	Description *string   `json:"description"`
}

type Episode_notation struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Short    string `json:"short"`
	Origin   string `json:"origin"`
	Preview  string `json:"preview"`
	Notation string `json:"notation"`
}
type Notation struct {
	ID       int    `json:"id"`
	Notation string `json:"notation"`
}
type Notation_del struct {
	ID int `json:"id"`
}
type Episode_notation_responce struct {
	Episode  Episode_db         `json:"episode"`
	Notation []Episode_notation `json:"notation"`
}
type Statistics_main struct {
	Games_news            int64 `json:"games_news"`
	Last_30day            int64 `json:"last_30day"`
	This_month_from_first int64 `json:"this_month_from_first"`
	Last_month            int64 `json:"last_month"`
	Origins               int64 `json:"origins"`
	Episodes              int64 `json:"episodes"`
	Episodes_released     int64 `json:"episodes_released"`
	Notations             int64 `json:"notations"`
	Deleted_notations     int64 `json:"deleted_notations"`
}
type Statistics_by_month_db struct {
	Month_name                    string `json:"month_name"`
	Year                          int    `json:"year"`
	Start_of_month_unix_timestamp int64  `json:"start_of_month_unix_timestamp"`
	End_of_month_unix_timestamp   int64  `json:"end_of_month_unix_timestamp"`
	Count_news                    int64  `json:"count_news"`
}
type Statistics struct {
	Main                  Statistics_main                     `json:"main"`
	By_month              []Statistics_by_month_db            `json:"by_month"`
	By_origin             Statistics_by_origin                `json:"by_origin"`
	By_origin_this_month  Statistics_by_origin                `json:"by_origin_this_month"`
	By_origin_in_released News_by_origin_in_released_episodes `json:"by_origin_in_released"`
}
type News_by_origin_in_released_episodes struct {
	Origin []string `json:"origin"`
	Count  []int64  `json:"count"`
}
type Statistics_by_origin struct {
	Origins []string `json:"origins"`
	Count   []int64  `json:"count"`
}

type Scenario_create struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Scenario_delete struct {
	ID int `json:"id"`
}
type Scenarios struct {
	ID            int       `json:"id"`
	Number        int       `json:"number"`
	Title         string    `json:"title"`
	Date          time.Time `json:"date"`
	Date_released time.Time `json:"date_released"`
	Released      bool      `json:"released"`
}
type Scenario struct {
	ID            int       `json:"id"`
	Number        int       `json:"number"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	Date          time.Time `json:"date"`
	Date_released time.Time `json:"date_released"`
	Released      bool      `json:"released"`
}
type Scenario_update struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}
type ChannelStatistics struct {
	ViewCount       string `json:"viewCount"`
	SubscriberCount string `json:"subscriberCount"`
}

type Data_from_youtube struct {
	Items []struct {
		Statistics ChannelStatistics `json:"statistics"`
	} `json:"items"`
}
type Timeline struct {
	Number   int64
	Title    string
	Timecode string
}
type Episode_meta_information struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
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
	r.HandleFunc("/episode/delete/", delete_episode).Methods(http.MethodPost)
	r.HandleFunc("/episode/add_news/", add_news_from_favorit).Methods(http.MethodPost)
	r.HandleFunc("/episode/get_all", get_all_episodes)
	r.HandleFunc("/episode/get/{id}", get_episode_by_id)
	r.HandleFunc("/episode/get/{id}/pdf", generate_pdf_for_episode).Methods(http.MethodGet)
	r.HandleFunc("/episode/contents/", contents).Methods(http.MethodPost)
	r.HandleFunc("/episode/update_contents/", update_contents).Methods(http.MethodPost)
	r.HandleFunc("/episode/update_intro/", update_intro).Methods(http.MethodPost)
	r.HandleFunc("/episode/update_ending/", update_ending).Methods(http.MethodPost)
	r.HandleFunc("/episode/notation/update/", update_notation).Methods(http.MethodPost)
	r.HandleFunc("/episode/notation/delete/", delete_notation).Methods(http.MethodPost)
	r.HandleFunc("/episode/release/", release_episode).Methods(http.MethodPost)

	r.HandleFunc("/statistics", get_statistics)
	r.HandleFunc("/statistics/youtube", get_youtube_statistics)

	r.HandleFunc("/scenario/add", add_scenario).Methods(http.MethodPost)
	r.HandleFunc("/scenario/delete", delete_scenario).Methods(http.MethodPost)
	r.HandleFunc("/scenario/get_all", get_scenarios)
	r.HandleFunc("/scenario/{id}", get_scenarios_by_id)
	r.HandleFunc("/scenario/update/", update_scenario).Methods(http.MethodPost)
	r.HandleFunc("/scenario/release/", release_scenario).Methods(http.MethodPost)

	corsHandler := handlers.CORS(headers, methods, origins)(r)

	http.ListenAndServe(":80", corsHandler)
}
func update_contents(w http.ResponseWriter, r *http.Request) {
	var meta Episode_meta_information
	err := json.NewDecoder(r.Body).Decode(&meta)
	if err != nil {
		logger.Warn(err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "update allnews.episode set description = $1 where id=$2"

	_, err = db.Exec(req, meta.Content, meta.Id)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	logger.Info(fmt.Sprintf("Contents for episode %d updated", meta.Id))
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = fmt.Sprintf("Опис для епізоду %d успішно оновлено", meta.Id)
	json.NewEncoder(w).Encode(resp)

}
func update_intro(w http.ResponseWriter, r *http.Request) {
	var meta Episode_meta_information
	json.NewDecoder(r.Body).Decode(&meta)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "update allnews.episode set intro = $1 where id=$2"

	_, err = db.Exec(req, meta.Content, meta.Id)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	logger.Info(fmt.Sprintf("Intro for episode %d updated", meta.Id))
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = fmt.Sprintf("Вступ для епізоду %d успішно оновлено", meta.Id)
	json.NewEncoder(w).Encode(resp)
}

func update_ending(w http.ResponseWriter, r *http.Request) {
	var meta Episode_meta_information
	json.NewDecoder(r.Body).Decode(&meta)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "update allnews.episode set ending = $1 where id=$2"

	_, err = db.Exec(req, meta.Content, meta.Id)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	logger.Info(fmt.Sprintf("Ending for episode %d updated", meta.Id))
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = fmt.Sprintf("Закінчення для епізоду %d успішно оновлено", meta.Id)
	json.NewEncoder(w).Encode(resp)
}
func contents(w http.ResponseWriter, r *http.Request) {
	// Парсимо multipart form, який містить файл
	err := r.ParseMultipartForm(10 << 20) // 10MB - максимальний розмір файлу
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	timeline := get_contents_from_edl(file)
	contents := []string{}

	// Додавання заголовків
	contents = append(contents, "Сюди вставити опис для відео")
	contents = append(contents, "---------------------------------------------------------------------------------------------------------------")
	contents = append(contents, "Мій телеграм - https://t.me/darthcitizen")
	contents = append(contents, "Мій інстаграм - https://www.instagram.com/sianotalone")
	contents = append(contents, "---------------------------------------------------------------------------------------------------------------")

	// Додавання елементів таймлайну
	for _, item := range timeline {
		contents = append(contents, item.Timecode+" — "+item.Title)
	}
	resp := map[string]interface{}{
		"contents": contents,
	}
	logger.Info("Get contents from EDL file")
	// Відправляємо результат``
	w.Header().Set("Content-Type", "text/plain")
	json.NewEncoder(w).Encode(resp)
}

func get_contents_from_edl(file io.Reader) []Timeline {
	var timeline []Timeline
	scanner := bufio.NewScanner(file)
	var i = 0
	var item = Timeline{}

	for scanner.Scan() {
		if i == 0 {
			if strings.Contains(scanner.Text(), "TITLE") {
				// fmt.Println("#" + strconv.Itoa(i) + " " + scanner.Text())
				i += 1
				continue
			} else {
				fmt.Println("This is not .edl file")
				break
			}
		}
		if i >= 3 {
			timecodeline_splitted := strings.Split(scanner.Text(), " ")
			if len(timecodeline_splitted) > 1 {
				num, err := strconv.ParseInt(timecodeline_splitted[0], 10, 64)
				if err != nil {

				}

				if num != 0 {
					item.Number = num
					time_splited := strings.Split(scanner.Text(), ":")
					item.Timecode = time_splited[1] + ":" + time_splited[2]
				}

			}
			titleline_splitted := strings.Split(scanner.Text(), "|")
			if len(titleline_splitted) > 2 {
				item.Title = titleline_splitted[2][2:]

				timeline = append(timeline, item)
				item = Timeline{}
			}

		}
		i += 1

	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Помилка при скануванні файлу:", err)
		return nil
	}
	return timeline
}

func release_scenario(w http.ResponseWriter, r *http.Request) {
	sc := Scenario_delete{}
	err := json.NewDecoder(r.Body).Decode(&sc)
	if err != nil {
		logger.Warn("Wrong json")
		return
	}
	req := "UPDATE allnews.scenario SET date_released = CURRENT_DATE, released = true WHERE id =$1;"
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
		return
	}
	_, err = db.Exec(req, sc.ID)
	if err != nil {
		logger.Warn(err)
		return
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Сценарій успішно випущено"
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Scenario was released with ID %d", sc.ID))
}
func release_episode(w http.ResponseWriter, r *http.Request) {
	ep := Delete_episode{}
	err := json.NewDecoder(r.Body).Decode(&ep)
	if err != nil {
		logger.Warn("Wrong json")
		return
	}
	req := "update allnews.episode set released=true where id=$1;"
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
		return
	}
	_, err = db.Exec(req, ep.ID)
	if err != nil {
		logger.Warn(err)
		return
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Випуск успішно випущено"
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Episode was released with ID %d", ep.ID))

}
func get_youtube_statistics(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		logger.Warn("Error loading .env file:", err)
		return
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		logger.Warn("API key not found in .env file")
		return
	}
	channelID := "UCSNmY29-UaPIPhYGY3lV2Kg"

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=statistics&id=%s&key=%s", channelID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		logger.Warn("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("Error reading response:", err)
		return
	}

	var data Data_from_youtube
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	if len(data.Items) > 0 {
		statistics := data.Items[0].Statistics
		logger.Info("Get youtube statistics")
		json.NewEncoder(w).Encode(statistics)

	} else {
		logger.Warn("No statistics found for the channel")
		json.NewEncoder(w).Encode("error")
	}

}
func update_scenario(w http.ResponseWriter, r *http.Request) {
	var s Scenario_update
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		logger.Warn(err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := `UPDATE allnews.scenario SET number = $1, title = $2, body = $3 WHERE id = $4;`
	_, err = db.Exec(req, s.Number, s.Title, s.Body, s.ID)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = fmt.Sprintf("Успішно оновлено сценарій із ID %d", s.ID)
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Update notation with ID %d", s.ID))
}
func get_scenarios_by_id(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str_id := vars["id"]
	id, err := strconv.Atoi(str_id)

	if err != nil {
		panic(err)
	}
	s := Scenario{}
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := `select id, number, title, body, date , date_released, released from allnews.scenario where id =$1;`

	rows, err := db.Query(req, id)
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Number, &s.Title, &s.Body, &s.Date, &s.Date_released, &s.Released)
		if err != nil {
			logger.Warn(err)
		}
	}
	db.Close()
	logger.Info(fmt.Sprintf("Get scenario with ID %d", id))
	json.NewEncoder(w).Encode(s)

}
func get_scenarios(w http.ResponseWriter, r *http.Request) {
	var s []Scenarios

	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	rows, err := db.Query("select id, number, title, date , date_released, released from allnews.scenario ORDER BY id DESC")
	if err != nil {
		logger.Warn(err)
	}
	for rows.Next() {
		var scenario Scenarios
		err = rows.Scan(&scenario.ID, &scenario.Number, &scenario.Title, &scenario.Date, &scenario.Date_released, &scenario.Released)
		if err != nil {
			logger.Warn(err)
		}
		s = append(s, scenario)
	}
	db.Close()
	logger.Info("Get all scenarios")
	json.NewEncoder(w).Encode(s)
}

func delete_scenario(w http.ResponseWriter, r *http.Request) {
	var scenario_id Scenario_delete
	json.NewDecoder(r.Body).Decode(&scenario_id)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "DELETE FROM allnews.scenario WHERE id = $1;"
	_, err = db.Exec(req, scenario_id.ID)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	mess := fmt.Sprintf("Успішно видалено сценарій із ID %d", scenario_id.ID)
	resp.Message = mess
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Delete scenario with ID %d", scenario_id.ID))

}
func add_scenario(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	s := Scenario_create{}
	req := `INSERT INTO allnews.scenario (number, title, date, date_released, body) 
	VALUES ($1, $2, CURRENT_DATE, CURRENT_DATE, '');`
	json.NewDecoder(r.Body).Decode(&s)
	_, err = db.Exec(req, s.Number, s.Title)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Новий сценарій створено"
	json.NewEncoder(w).Encode(resp)
	logger.Info("New scenario has been created")
}

func get_statistics(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	s := Statistics{}
	req := `
	SELECT
	  (SELECT COUNT(id) FROM allnews.games_news) AS games_news,
	  (select count(dc.id) as last_month from allnews.games_news dc
		left join (SELECT EXTRACT(EPOCH FROM NOW() - INTERVAL '30 days')::bigint AS unix_timestamp) t on 1=1
		where time > t.unix_timestamp
		group by t.unix_timestamp) as last_30days,
		( select count(id) from allnews.games_news
	 left join  ( SELECT EXTRACT(EPOCH FROM (DATE_TRUNC('MONTH', CURRENT_DATE)::DATE))::BIGINT AS unix_timestamp) t on 1=1
	 where time>t.unix_timestamp) as this_month_from_first,
	  (select count(id) from allnews.games_news
	left join (SELECT EXTRACT(epoch FROM DATE_TRUNC('MONTH', CURRENT_DATE - INTERVAL '1 month')) AS first_last_month) l on 1=1
	left join (SELECT EXTRACT(EPOCH FROM DATE_TRUNC('month', CURRENT_DATE))::int as first_curr_month) c on 1=1
	where time>l.first_last_month and time <c.first_curr_month
	) as lastmonth,
	  (SELECT COUNT(DISTINCT origin) FROM allnews.games_news) AS origins,
	  (SELECT COUNT(id) FROM allnews.episode) AS episodes,
	  (SELECT COUNT(id) FROM allnews.episode where released=true) AS episodes_released,
	  (SELECT COUNT(n.id) FROM allnews.notation n
			   left join allnews.episode e on e.id = n.episode_id
	   where n.deleted=false and e.released=true) AS notation,
	  (SELECT COUNT(id) FROM allnews.notation where deleted=true) AS deleted_notation;	`
	rows, err := db.Query(req)
	for rows.Next() {
		m := Statistics_main{}
		err = rows.Scan(&m.Games_news, &m.Last_30day, &m.This_month_from_first, &m.Last_month, &m.Origins, &m.Episodes, &m.Episodes_released, &m.Notations, &m.Deleted_notations)
		if err != nil {
			logger.Warn(err)
		}
		s.Main = m
	}

	req = `WITH month_intervals AS (
		SELECT 
		  generate_series(
			date_trunc('month', CURRENT_DATE - interval '11 months'),
			date_trunc('month', CURRENT_DATE),
			interval '1 month'
		  ) AS start_of_month
	  )
	  SELECT
		TO_CHAR(start_of_month, 'Month') AS month_name,
		EXTRACT(YEAR FROM start_of_month) AS year,
		EXTRACT(EPOCH FROM start_of_month)::int AS start_of_month_unix_timestamp,
		EXTRACT(EPOCH FROM (start_of_month + interval '1 month' - interval '1 day'))::int AS end_of_month_unix_timestamp,
		(SELECT COUNT(id) FROM allnews.games_news 
		 WHERE time >= EXTRACT(EPOCH FROM start_of_month) 
		   AND time <= EXTRACT(EPOCH FROM (start_of_month + interval '1 month' - interval '1 day'))) AS count_news
	  FROM month_intervals;`
	rows, err = db.Query(req)

	for rows.Next() {
		m := Statistics_by_month_db{}
		err = rows.Scan(&m.Month_name, &m.Year, &m.Start_of_month_unix_timestamp, &m.End_of_month_unix_timestamp, &m.Count_news)
		if err != nil {
			logger.Warn(err)
		}
		s.By_month = append(s.By_month, m)

	}

	req = `select count(id) as news, origin from allnews.games_news group by origin order by 1 desc`
	rows, err = db.Query(req)

	for rows.Next() {
		origin := ""
		count := int64(0)
		err = rows.Scan(&count, &origin)
		if err != nil {
			logger.Warn(err)
		}
		s.By_origin.Count = append(s.By_origin.Count, count)
		s.By_origin.Origins = append(s.By_origin.Origins, origin)

	}

	req = `select origin, count(id) from allnews.games_news
	where time>
	(SELECT EXTRACT(epoch FROM date_trunc('month', CURRENT_DATE))::int AS unix_timestamp_of_first_day_of_this_month)
	group by origin
	order by 2 desc`

	rows, err = db.Query(req)

	for rows.Next() {
		origin := ""
		count := int64(0)
		err = rows.Scan(&origin, &count)
		if err != nil {
			logger.Warn(err)
		}
		s.By_origin_this_month.Count = append(s.By_origin_this_month.Count, count)
		s.By_origin_this_month.Origins = append(s.By_origin_this_month.Origins, origin)
	}

	req = `select max(c.origin) as origin, count(*) from allnews.notation dc
	left join allnews.games_news c on c.id = dc.news_id
	left join allnews.episode c1 on c1.id = dc.episode_id
	where dc.deleted=false and c1.released = true
	group by c.origin
	order by count DESC`
	n := News_by_origin_in_released_episodes{}
	rows, err = db.Query(req)
	for rows.Next() {

		var o string
		var c int64
		err = rows.Scan(&o, &c)
		if err != nil {
			logger.Warn(err)
		}
		n.Origin = append(n.Origin, o)
		n.Count = append(n.Count, c)

	}

	s.By_origin_in_released = n

	db.Close()

	json.NewEncoder(w).Encode(s)
	logger.Info("Get statistics")

}
func delete_notation(w http.ResponseWriter, r *http.Request) {
	var id Notation_del
	json.NewDecoder(r.Body).Decode(&id)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "update allnews.notation set deleted=true where id=$1;"
	_, err = db.Exec(req, id.ID)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Успішно видалено нотатку"
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Delete notation with ID %d", id.ID))
}
func update_notation(w http.ResponseWriter, r *http.Request) {
	var notation Notation
	json.NewDecoder(r.Body).Decode(&notation)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "update allnews.notation set notation=$1 where id=$2;"
	_, err = db.Exec(req, notation.Notation, notation.ID)
	if err != nil {
		logger.Warn(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	resp.Message = "Успішно оновлено нотатку"
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Update notation with ID %d", notation.ID))

}
func get_episode_by_id(w http.ResponseWriter, r *http.Request) {
	var notations []Episode_notation
	var response Episode_notation_responce

	vars := mux.Vars(r)
	str_id := vars["id"]
	id, err := strconv.Atoi(str_id)

	if err != nil {
		panic(err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := `select dc.id, news.title, news.url, news.short, news.origin, news.preview, dc.notation
	from allnews.notation dc
	left join allnews.games_news news on dc.news_id = news.id
	where dc.deleted=false and dc.episode_id=`
	req += str_id
	req += ` order by 2`
	rows, err := db.Query(req)
	for rows.Next() {
		e := Episode_notation{}
		err = rows.Scan(&e.ID, &e.Title, &e.Url, &e.Short, &e.Origin, &e.Preview, &e.Notation)
		if err != nil {
			logger.Warn(err)
		}
		notations = append(notations, e)
	}
	response.Notation = notations
	req = fmt.Sprintf("select id,name,number,date,released, intro, ending, description from allnews.episode where id=%d", id)
	rows, err = db.Query(req)
	for rows.Next() {
		e := Episode_db{}
		err = rows.Scan(&e.ID, &e.Name, &e.Number, &e.Date, &e.Released, &e.Intro, &e.Ending, &e.Description)
		if err != nil {
			logger.Warn(err)
		}
		response.Episode = e
	}
	db.Close()
	json.NewEncoder(w).Encode(response)
	logger.Info("Get episode with ID ", id)

}
func generate_pdf_for_episode(w http.ResponseWriter, r *http.Request) {
	var notations []Episode_notation
	var response Episode_notation_responce

	vars := mux.Vars(r)
	str_id := vars["id"]
	id, err := strconv.Atoi(str_id)

	if err != nil {
		panic(err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := `select dc.id, news.title, news.url, news.short, news.origin, news.preview, dc.notation
	from allnews.notation dc
	left join allnews.games_news news on dc.news_id = news.id
	where dc.deleted=false and dc.episode_id=`
	req += str_id
	req += ` order by 2`
	rows, err := db.Query(req)
	for rows.Next() {
		e := Episode_notation{}
		err = rows.Scan(&e.ID, &e.Title, &e.Url, &e.Short, &e.Origin, &e.Preview, &e.Notation)
		if err != nil {
			logger.Warn(err)
		}
		notations = append(notations, e)
	}
	response.Notation = notations
	req = fmt.Sprintf("select id,name,number,date,released, intro, ending, description from allnews.episode where id=%d", id)
	rows, err = db.Query(req)
	for rows.Next() {
		e := Episode_db{}
		err = rows.Scan(&e.ID, &e.Name, &e.Number, &e.Date, &e.Released, &e.Intro, &e.Ending, &e.Description)
		if err != nil {
			logger.Warn(err)
		}
		response.Episode = e
	}
	db.Close()
	pdf := gen_pdf_episode(response)
	header := fmt.Sprintf("attachment; filename=%s.pdf", response.Episode.Name)
	w.Header().Set("Content-Type", "application/pdf; charset=utf-8")
	w.Header().Set("Content-Disposition", header)
	if _, err := w.Write(pdf); err != nil {
		http.Error(w, "Помилка при відправленні PDF", http.StatusInternalServerError)
		logger.Warn(err)
		return
	}
	logger.Info(fmt.Sprintf("Generate PDF for episode with ID %d", id))
}

func gen_pdf_episode(episode Episode_notation_responce) []byte {
	var buf bytes.Buffer
	pwd, _ := os.Getwd()
	pdf := gofpdf.New("P", "mm", "A4", pwd+"/font")
	pdf.AddUTF8Font("Roboto", "", "Roboto-Regular.ttf")

	pdf.AddPage()
	pdf.SetFont("Roboto", "", 16)
	pdf.MultiCell(190, 10, episode.Episode.Name, "", "C", false)
	//формуємо інтро
	intro := delete_html_tags(*episode.Episode.Intro)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(190, 5, "Вступ", "", "C", false)
	pdf.SetFont("Roboto", "", 8)
	pdf.MultiCell(190, 5, intro, "", "J", false)
	pdf.MultiCell(190, 5, "\n", "0", "0", false)
	//формуємо новини
	for _, item := range episode.Notation {
		item.Notation = delete_html_tags(item.Notation)
		pdf.SetFont("Roboto", "", 12)
		pdf.MultiCell(190, 5, item.Title, "", "C", false)
		pdf.SetFont("Roboto", "", 8)
		pdf.MultiCell(190, 5, item.Notation, "", "J", false)
		pdf.MultiCell(190, 5, "\n", "0", "0", false)
	}
	//формуємо закінчення
	ending := delete_html_tags(*episode.Episode.Ending)
	pdf.SetFont("Roboto", "", 12)
	pdf.MultiCell(190, 5, "Закінчення", "", "C", false)
	pdf.SetFont("Roboto", "", 8)
	pdf.MultiCell(190, 5, ending, "", "J", false)
	pdf.MultiCell(190, 5, "\n", "0", "0", false)

	err := pdf.Output(&buf)
	if err != nil {
		return nil
	}
	return buf.Bytes()

}
func delete_html_tags(str string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(str, "")
}

func get_all_episodes(w http.ResponseWriter, r *http.Request) {
	var episodes []Episode_db
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		logger.Warn(err)
	}
	req := "select id,name,number,date,released from allnews.episode order by number, name asc;"
	rows, err := db.Query(req)
	for rows.Next() {
		e := Episode_db{}
		err = rows.Scan(&e.ID, &e.Name, &e.Number, &e.Date, &e.Released)
		if err != nil {
			log.Fatal(err)
		}
		episodes = append(episodes, e)
	}
	db.Close()
	json.NewEncoder(w).Encode(episodes)
	logger.Info("Get all episodes")

}
func add_news_from_favorit(w http.ResponseWriter, r *http.Request) {
	var news Add_news
	var news_count int
	news_count = 0
	json.NewDecoder(r.Body).Decode(&news)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	req :=
		`select id, title, short, origin, url, preview, time, favorit from allnews.games_news
	where favorit = true
	`
	rows, err := db.Query(req)
	var gaming_news []Games_News
	for rows.Next() {
		f := Games_News{}
		err = rows.Scan(&f.Id, &f.Title, &f.Short, &f.Origin, &f.Url, &f.Preview, &f.Time, &f.Favorit)
		if err != nil {
			log.Fatal(err)
		}
		gaming_news = append(gaming_news, f)
	}

	for _, value := range gaming_news {
		req = "select count(news_id) from allnews.notation where news_id=$1 and episode_id=$2;"
		var count int
		rows, err := db.Query(req, value.Id, news.Episode_ID)
		if err != nil {
			logger.Error(err)
		}
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				logger.Error(err)
			}

		}
		if count == 0 {
			news_count += 1
			req = "INSERT INTO allnews.notation (episode_id, news_id, notation) VALUES ($1, $2, '');"
			_, err = db.Exec(req, news.Episode_ID, value.Id)
			if err != nil {
				logger.Error(err)
			}
		}

	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	var mess string
	if news_count == 0 {
		mess = fmt.Sprintf("Новин не додано до випуску %d", news.Episode_ID)
	} else {
		mess = fmt.Sprintf("Новини додано у випуск %d у кількості %d", news.Episode_ID, news_count)
	}

	resp.Message = mess
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Add news to episode with ID %d with count %d", news.Episode_ID, news_count))

}
func delete_episode(w http.ResponseWriter, r *http.Request) {
	var episode Delete_episode
	json.NewDecoder(r.Body).Decode(&episode)
	db, err := sql.Open("postgres", "postgres://postgres:12345678@localhost/news?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	req := "DELETE FROM allnews.episode WHERE id = $1;"
	_, err = db.Exec(req, episode.ID)
	if err != nil {
		panic(err)
	}
	db.Close()
	var resp Episode_response
	resp.Status = "200, OK"
	mess := fmt.Sprintf("Успішно видалено епізод із ID %d", episode.ID)
	resp.Message = mess
	json.NewEncoder(w).Encode(resp)
	logger.Info(fmt.Sprintf("Delete episode with ID %d", episode.ID))
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
	var query string = "select id, title, short, origin, url, preview, time, favorit from allnews.games_news WHERE origin = '" + origin + "' ORDER BY id DESC LIMIT 20  OFFSET " + strconv.FormatInt(int64(offset), 10)
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
	query := "select id, title, short, origin, url, preview, time, favorit from allnews.games_news WHERE favorit = True ORDER BY id DESC"
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
	query := "select id, title, short, origin, url, preview, time, favorit from allnews.games_news WHERE origin = '" + origin + "' ORDER BY id DESC LIMIT 20"
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
	rows, err := db.Query("select id, title, short, origin, url, preview, time, favorit from allnews.games_news ORDER BY id DESC LIMIT 20")
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
	var query string = "select id, title, short, origin, url, preview, time, favorit from allnews.games_news ORDER BY id DESC LIMIT 20  OFFSET " + strconv.FormatInt(int64(offset), 10)
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
