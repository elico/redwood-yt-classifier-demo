package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/asaskevich/govalidator"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"regexp"
	"github.com/twinj/uuid"
)

var YouTubeIDregex = regexp.MustCompile("(?:youtube(?:-nocookie)?\\.com\\/(?:[^\\/\n\\s]+\\/\\S+\\/|(?:v|e(?:mbed)?)\\/|\\S*?[?&]v=)|youtu\\.be\\/)([a-zA-Z0-9_-]{11})(?:\\&|\\?|$)")

var (
	DBCon *sql.DB
	conf  Config

	configFile string
	debug      int
)

type Config struct {
	Database database
	Server   server
}

type database struct {
	Server   string
	Port     string
	Database string
	User     string
	Password string
	Timeout  uint
}

type server struct {
	Listen string
}

type Tag struct {
	VideoID string `json:"videoid"`
	Title   string `json:"title"`
	Rated   int    `json:"rated"`
}

func tomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestXRay := uuid.NewV4()
	w.Header().Set("Request-X-Rat", requestXRay.String())

	switch {
	case "GET" == r.Method || "POST" == r.Method:
		result := make(map[string]int)
		reuesrUrl := r.FormValue("url")
		validURLStr := govalidator.IsRequestURL(reuesrUrl)
		matchID := ""
		if validURLStr {
			res := YouTubeIDregex.FindStringSubmatch(reuesrUrl)
			if len(res) == 2 {
				matchID = res[1]
				var tag Tag
				w.Header().Set("YT-ID-Match", matchID)
				row := DBCon.QueryRow("SELECT videoid, title, rated FROM videos WHERE videoid = ?", matchID)
				err := row.Scan(&tag.VideoID, &tag.Title, &tag.Rated)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				w.Header().Set("YT-ID-Rate", fmt.Sprintf("%d", tag.Rated))

				if tag.Rated > 60 {
					result["nudity"] = 1000

				}

			} else {
				if debug > 1 {
					fmt.Println("Request:", requestXRay.String(), "YouTube ID is not present in the URL")
				}
			}
		} else {
			if debug > 1 {
				fmt.Println("Request:", requestXRay.String(), "Has Invalid URL")
			}
		}

		j, _ := json.MarshalIndent(result, "", "  ")

		w.Write(j)
		w.Write([]byte("\n"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{ \"msg\": \"Method is not allowed.\" }\n")
	}
}

func init() {
	flag.StringVar(&configFile, "f", "config.toml", "Toml config file location")
	flag.IntVar(&debug, "debug", 0, "Debug level")

	flag.Parse()

	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		fmt.Println(err)
	}
	if debug > 0 {
		fmt.Printf("%#v\n", conf)
	}
}

func main() {
	var err error

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%ds", conf.Database.User, conf.Database.Password, conf.Database.Server, conf.Database.Port, conf.Database.Database, conf.Database.Timeout)
	DBCon, err = sql.Open("mysql", connString)

	if err != nil {
		panic(err.Error())
	}
	defer DBCon.Close()
	http.HandleFunc("/tom", tomHandler)

	log.Println("Go!")
	panic(http.ListenAndServe(conf.Server.Listen, nil))
}
