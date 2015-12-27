package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"crypto/rand"
	"crypto/md5"
	"html/template"
	"net/http"
	"time"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
)

/* Database settings  */
const (
	DBNAME string = "chaupar"
	DBUSER string = "hoter"
	DBPASS string = ""
	ADDR string = ":8081"
	TOKEN string = "token"
)

var db *sql.DB

/* Player structure  */
type Player struct {
	ID int
	Name string
}

/* Game structure  */
type Game struct {
	ID int
	Players [4]Player
}

func main() {
	var err error
	db, err = sql.Open("mysql", DBUSER + ":" + DBPASS + "@/" + DBNAME + "?charset=utf8");
	checkErr(err)
	defer db.Close()
	
	http.HandleFunc("/", homeHandler)
	http.Handle("/s", websocket.Handler(testSocket))
	http.HandleFunc("/templates/", func (c http.ResponseWriter, r *http.Request) {
		http.ServeFile(c, r, r.URL.Path[1:])
	})

	err = http.ListenAndServe(ADDR, nil)
	checkErr(err)
}

func testSocket(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

/**
 * Generate the homepage html file
 */
func homeHandler(c http.ResponseWriter, r *http.Request) {
	var message string
	templateFile := template.Must(template.ParseFiles("templates/login.html"))
	if _, err := r.Cookie(TOKEN); err != nil {
		if r.Method == "POST" {
			r.ParseForm()
			if len(r.Form["name"][0]) == 0 {
				message = "Please enter your nickname"
			} else {
				token := getRandToken()
				message = "Name: " + r.Form["name"][0] + " password: " + r.Form["pass"][0]
				cookie := http.Cookie{Name: TOKEN, Value: token}
				http.SetCookie(c, &cookie)
				
				stmt, err := db.Prepare("INSERT users SET name=?, pass=?, token=?, created=?")
				checkErr(err)
				pass := md5.Sum([]byte(r.Form["pass"][0]))
				_, err = stmt.Exec(r.Form["name"][0], string(pass[:]), token, time.Now().Format("2006-01-02 15:04:05"))
				checkErr(err)
				
				templateFile = template.Must(template.ParseFiles("templates/main.html", "templates/game.html"))
			}
		}
	} else {
		templateFile = template.Must(template.ParseFiles("templates/main.html", "templates/game.html"))
	}

	if message != "" {
		params := struct {
			Message string
		}{message}
		templateFile.Execute(c, params)
	} else {
		templateFile.Execute(c, nil)	
	}
	
}

/**
 * Panic if the error is existed
 */
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

/**
 * Get a new random token.
 */
func getRandToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
