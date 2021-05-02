package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type sendIndex struct {
	Posts    []Post
	Erreur   string
	Post     string
	Login    string
	PasLogin string
}

type Post struct {
	Id       int
	Title    string
	Content  string
	Date     string
	Like     int
	Category string
}

type User struct {
	id        int
	name      string
	firstname string
	mail      string
	password  string
	pseudo    string
	cookie    string
}

var Erreur2, Post2, Log2, notLogin2 string

func main() {
	fs := http.FileServer(http.Dir("./templates/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	port := "8099"
	createDatabase()
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/newpost", newPost)
	http.HandleFunc("/posts", postList)
	http.HandleFunc("/post", post)
	http.HandleFunc("/wantpost", wantpost)
	http.HandleFunc("/looklog", looklog)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/like", like)
	println("Le serveur web se lance sur le port " + port)
	http.ListenAndServe(":"+port, nil)
}

func like(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		Erreur2 = "Erreur, Vous devez être connecté pour écrire un poste"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		ID := r.FormValue("id")
		fmt.Println(ID)
		myCookie, err := r.Cookie("sessionId")
		if err != nil {
			//mettre une erreur
		}
		sessionId := myCookie.Value
		z := strings.Split(sessionId, ":")
		pseudo := z[0]
		database, _ := sql.Open("sqlite3", "./databases/likes.db")
		rows, _ := database.Query("SELECT pseudo FROM likes WHERE post = '" + ID + "'")
		for rows.Next() {
			var Pseudo string
			rows.Scan(&Pseudo)
			if Pseudo == pseudo {
				Erreur2 = "Erreur, Vous ne pouvez mettre qu'un seul like par post"
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
		rows.Close()
		statement, _ := database.Prepare("INSERT INTO likes (pseudo, post) VALUES (?, ?)")
		statement.Exec(pseudo, ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		fmt.Println("liké")
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	myCookie, err := r.Cookie("sessionId")
	if err != nil {
		//mettre une erreur
	}
	sessionId := myCookie.Value
	z := strings.Split(sessionId, ":")
	pseudo := z[0]
	database, _ := sql.Open("sqlite3", "./databases/session.db")
	rows, _ := database.Prepare("DELETE FROM session WHERE pseudo = ?")
	_, _ = rows.Exec(pseudo)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func looklog(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		notLogin2 = "notLog"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		Log2 = "Log"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func wantpost(w http.ResponseWriter, r *http.Request) {
	if !isLogged(r) {
		Erreur2 = "Erreur, Vous devez être connecté pour ércire un poste"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		Post2 = "Connecté"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	database, _ := sql.Open("sqlite3", "./databases/posts.db")
	rows, _ := database.Query("SELECT * FROM posts  WHERE id = '" + id + "' ")
	var post Post
	if rows.Next() {
		rows.Scan(&post.Id, &post.Title, &post.Content, &post.Date, &post.Like, &post.Category)
	}
	rows.Close()
	files := []string{"./templates/post.html", "./templates/template.html"}

	tpl, err := template.ParseFiles(files...)
	if err != nil {
		print(err.Error())
	} else {
		tpl.Execute(w, &post)
	}
}

func postList(w http.ResponseWriter, r *http.Request) {
	database, _ := sql.Open("sqlite3", "./databases/posts.db")
	rows, _ := database.Query("SELECT * FROM posts ")
	var tabpost []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.Id, &post.Title, &post.Content, &post.Date, &post.Like, &post.Category)
		tabpost = append(tabpost, post)
	}
	rows.Close()
	database2, _ := sql.Open("sqlite3", "./databases/likes.db")
	for i2 := 0; i2 < 3; i2++ {
		idPost := strconv.Itoa(tabpost[i2].Id)

		rows2, _ := database2.Query("SELECT id FROM likes WHERE post = '" + idPost + "' ")
		var nblike int
		for rows2.Next() {
			var id1 int
			rows2.Scan(&id1)
			nblike++
		}
		tabpost[i2].Like = nblike
		rows2.Close()

	}
	files := []string{"./templates/postList.html", "./templates/template.html"}

	tpl, err := template.ParseFiles(files...)
	if err != nil {
		print(err.Error())
	} else {
		tpl.Execute(w, &tabpost)
	}
}

func newPost(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("titlePost")
	content := r.FormValue("contentPost")
	category := r.FormValue("categoryPost")
	date := strconv.Itoa(time.Now().Day()) + "/" + strconv.Itoa(int(time.Now().Month())) + "/" + strconv.Itoa(time.Now().Year())
	database, _ := sql.Open("sqlite3", "./databases/posts.db")
	statement, _ := database.Prepare("INSERT INTO posts (title, content, date, like, category) VALUES (?, ?, ?, ?, ?)")
	statement.Exec(title, content, date, 0, category)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

//Probleme avec les selects
func login(w http.ResponseWriter, r *http.Request) {
	mail := r.FormValue("email")
	password := r.FormValue("psw")
	database, _ := sql.Open("sqlite3", "./databases/users.db")
	rows, _ := database.Query("SELECT mail, password, pseudo FROM users WHERE mail = '" + mail + "'")
	var tempUser User
	if rows.Next() {
		rows.Scan(&tempUser.mail, &tempUser.password, &tempUser.pseudo)
	}
	rows.Close()
	if tempUser.mail != mail || tempUser.password == "" {
		Erreur2 = "Erreur lors de la connection, Mot de passe ou mail incorrect bza"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	err := bcrypt.CompareHashAndPassword([]byte(tempUser.password), []byte(password))
	if err != nil {
		Erreur2 = "Erreur lors de la connection, Mot de passe ou mail incorrect"
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		addCookie(w, tempUser.pseudo)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
func register(w http.ResponseWriter, r *http.Request) {
	firstname := r.FormValue("firstname")
	pseudo := r.FormValue("pseudo")
	mail := r.FormValue("email")
	password := r.FormValue("psw")
	confirmpassword := r.FormValue("confirmpsw")
	if password != confirmpassword {
		Erreur2 = "Erreur lors de l'inscription, les mots de passes ne sont pas identiques"
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
	database, _ := sql.Open("sqlite3", "./databases/users.db")
	rows, _ := database.Query("SELECT mail, pseudo FROM users WHERE mail = '" + mail + "'")
	for rows.Next() {
		var Email string
		var Pseudo string
		rows.Scan(&Email, &Pseudo)
		if Pseudo == pseudo && Email == mail {
			Erreur2 = "Erreur lors de l'inscription, le pseudo et l'adresse mail saisi sont déja utilisé Veuillez vous connecter"
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else if Pseudo == pseudo {
			Erreur2 = "Erreur lors de l'inscription, le pseudo saisi est déja utilisé Veuillez vous connecter"
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else if Email == mail {
			Erreur2 = "Erreur lors de l'inscription, l'adresse mail saisi est déja utilisé Veuillez vous connecter"
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	rows.Close()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	statement, _ := database.Prepare("INSERT INTO users (firstname, mail, password, pseudo) VALUES (?, ?, ?, ?)")
	statement.Exec(firstname, mail, hashedPassword, pseudo)
	addCookie(w, pseudo)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func index(w http.ResponseWriter, r *http.Request) {
	database, _ := sql.Open("sqlite3", "./databases/posts.db")
	rows, _ := database.Query("SELECT * FROM posts ")
	var send sendIndex
	var tabpost []Post
	for i := 0; i < 4; rows.Next() {
		if i > 0 {
			var post Post
			rows.Scan(&post.Id, &post.Title, &post.Content, &post.Date, &post.Like, &post.Category)
			tabpost = append(tabpost, post)
		}
		i++
	}
	rows.Close()
	database2, _ := sql.Open("sqlite3", "./databases/likes.db")
	for i2 := 0; i2 < 3; i2++ {
		idPost := strconv.Itoa(tabpost[i2].Id)

		rows2, _ := database2.Query("SELECT id FROM likes WHERE post = '" + idPost + "' ")
		var nblike int
		for rows2.Next() {
			var id1 int
			rows2.Scan(&id1)
			nblike++
		}
		tabpost[i2].Like = nblike
		rows2.Close()
	}
	send.Posts = tabpost

	if Erreur2 == "" {
		send.Erreur = ""
		Erreur2 = ""
	} else {
		send.Erreur = Erreur2
		Erreur2 = ""
	}
	if Post2 == "" {
		send.Post = ""
		Post2 = ""
	} else {
		send.Post = Post2
		Post2 = ""
	}
	if Log2 == "" {
		send.Login = ""
		Log2 = ""
	} else {
		send.Login = Log2
		Log2 = ""
	}
	if notLogin2 == "" {
		send.PasLogin = ""
		notLogin2 = ""
	} else {
		send.PasLogin = notLogin2
		notLogin2 = ""
	}
	files := []string{"./templates/index.html", "./templates/template.html"}
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		print(err.Error())
	} else {
		tpl.Execute(w, &send)
	}
}

func isLogged(r *http.Request) bool {
	type Cook struct {
		id     int
		pseudo string
		uuid   string
	}
	myCookie, err := r.Cookie("sessionId")
	if err != nil {
		return false
	}
	sessionId := myCookie.Value
	z := strings.Split(sessionId, ":")
	uuid := z[1]
	pseudo := z[0]
	database, _ := sql.Open("sqlite3", "./databases/session.db")
	rows, _ := database.Query("SELECT * FROM session WHERE pseudo = '" + pseudo + "'")
	var id int
	var user string
	var value string
	tabCookie := []Cook{}
	for rows.Next() {
		rows.Scan(&id, &user, &value)
		cook := Cook{
			id:     id,
			pseudo: user,
			uuid:   value,
		}
		tabCookie = append(tabCookie, cook)
	}
	for _, v := range tabCookie {
		if v.uuid == uuid {
			return true
		}
	}
	rows.Close()
	return false
}

func addCookie(w http.ResponseWriter, pseudo string) {
	uuid, _ := uuid.NewV4()
	value := pseudo + ":" + uuid.String()
	expire := time.Now().Add(1 * time.Hour)
	cookie := http.Cookie{
		Name:    "sessionId",
		Value:   value,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	database, _ := sql.Open("sqlite3", "./databases/session.db")
	statement, _ := database.Prepare("INSERT INTO session (pseudo, cookie) VALUES (?, ?)")
	statement.Exec(pseudo, uuid.String())
}

func createDatabase() {
	database, _ := sql.Open("sqlite3", "./databases/users.db")
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, firstname VARCHAR(23) NOT NULL, mail VARCHAR(255) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, pseudo VARCHAR(25) UNIQUE NOT NULL)")
	statement.Exec()
	database2, _ := sql.Open("sqlite3", "./databases/posts.db")
	statement2, _ := database2.Prepare("CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, title VARCHAR(100) NOT NULL, content TEXT NOT NULL, date VARCHAR(10) NOT NULL, like INTEGER NOT NULL, category TEXT NOT NULL)")
	statement2.Exec()
	database3, _ := sql.Open("sqlite3", "./databases/session.db")
	statement3, _ := database3.Prepare("CREATE TABLE IF NOT EXISTS session (id INTEGER PRIMARY KEY AUTOINCREMENT, pseudo VARCHAR(25) NOT NULL, cookie VARCHAR(100) NOT NULL)")
	statement3.Exec()
	database4, _ := sql.Open("sqlite3", "./databases/likes.db")
	statement4, _ := database4.Prepare("CREATE TABLE IF NOT EXISTS likes (id INTEGER PRIMARY KEY AUTOINCREMENT, pseudo TEXT NOT NULL, post TEXT NOT NULL)")
	statement4.Exec()
}