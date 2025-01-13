package main

import (
	"encoding/json" // для сервера
	"fmt"
	"html/template"
	"net/http"

	//"strconv"

	//"context"
	"log"

	//"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"

	//"go.mongodb.org/mongo-driver/mongo/readpref"
	"database/sql"

	"time"

	_ "github.com/go-sql-driver/mysql"
	//"github.com/gorilla/mux"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
)

/*
type ViewData struct{

	    Title string
	    Users []string
	}
*/
type Post struct {
	Id int
	//Id          bson.ObjectId `bson:"_id"`
	Date        string //`bson:"date"`
	Sender      string
	Recipient   string
	Finder      string
	Title       string  //`bson:"title"`
	Description string  //`bson:"description"`
	Lost        string  //`bson:"lost"`
	Lat         float64 //`bson:"latitude"`
	Lon         float64 //`bson:"longitude"`
	IsFound     bool
	IsConfirmed bool
	Image       string
}
type User struct {
	Id       int
	Name     string
	Lastname string
	Password string
	Email    string
	Vk       string
	Telegram string
	Rating   int
}
type ViewData struct {
	Posts  []Post
	IsAuth bool
	Name   string
}

var database *sql.DB
var isAuth bool = false
var user User
var SECRET = "123"

func main() {

	db, err := sql.Open("mysql", "root:Xookauf3Ow@/area")

	if err != nil {
		log.Println(err)
	}
	database = db
	defer db.Close()

	//http.HandleFunc("/", IndexHandler)
	//http.HandleFunc("/add", AddHandler)
	//http.HandleFunc("/position", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "static/position.html")
	//})

	http.HandleFunc("/", IndexHandler)

	http.HandleFunc("/api/posts", ApiIndexHandler)
	handler := http.HandlerFunc(handleRequest)
	http.Handle("/api/token", handler)
	http.HandleFunc("/api/rating", ApiRatingHandler)

	http.HandleFunc("/add", loginСheck(AddHandler))
	http.HandleFunc("/register", loginСheck(RegisterHandler))
	http.HandleFunc("/login", loginСheck(LoginHandler))
	http.HandleFunc("/logout", loginСheck(logoutHandler))
	http.HandleFunc("/position", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/position.html")
	})
	http.HandleFunc("/prfl", loginСheck(ProfileHandler))
	http.HandleFunc("/rating", RatingHandler)
	//http.Handle("/", router)
	http.HandleFunc("/found", loginСheck(FoundHandler))
	http.HandleFunc("/confirm", loginСheck(ConfirmHandler))
	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}

var store = sessions.NewCookieStore([]byte("SESSION_KEY"))

func ProfileHandler(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")
	row := database.QueryRow("select * from area.Users where id=?", id)

	err := row.Scan(&user.Id, &user.Name, &user.Lastname, &user.Password, &user.Email, &user.Vk, &user.Telegram, &user.Rating)
	if err != nil {
		fmt.Println(err)

	}
	data := user
	tmpl, _ := template.ParseFiles("templates/profile.html")
	tmpl.Execute(w, data)
	//http.ServeFile(w, r, "templates/profile.html")

	//http.Redirect(w, r, "/", 301)

}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		row := database.QueryRow("select * from area.Users where email=? and password_=?", email, password)

		err = row.Scan(&user.Id, &user.Name, &user.Lastname, &user.Password, &user.Email, &user.Vk, &user.Telegram, &user.Rating)
		if err != nil {
			fmt.Println(err)

		}
		if user.Email == email && user.Password == password {
			log.Println("Succses!")

			//isAuth = true
			// Получаем сессию или создаём, если её не существует
			session, _ := store.Get(r, "session-cookie")

			// Добавляем данные в сессию
			session.Values["username"] = email
			session.Values["is_auth"] = true
			session.Values["user_id"] = user.Id
			// Сохраняем сессию в ответ
			session.Save(r, w)
			http.Redirect(w, r, "/", 301)

		} else {
			fmt.Fprintf(w, "Invalid login and/or password.")
		}
	} else {
		http.ServeFile(w, r, "static/login.html")
	}
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	session.Options.MaxAge = -1
	session.Save(r, w)

	// Перенаправляем на страницу входа
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		name := r.FormValue("name")
		lastname := r.FormValue("lastname")
		password := r.FormValue("password")
		email := r.FormValue("email")
		//vk := r.FormValue("vk")
		//telegram := r.FormValue("telegram")

		_, err = database.Exec("insert into area.Users (name_, lastname, password_, email, vk, telegram, rating) values (?, ?, ?, ?, ?, ?, ?)",
			name, lastname, password, email, "none", "none", 0)

		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, "/", 301)
	} else {
		http.ServeFile(w, r, "static/register.html")
	}
}

// получаем измененные данные и сохраняем их в БД

// функция добавления данных
func AddHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		title := r.FormValue("title")
		desc := r.FormValue("desc")
		lost := r.FormValue("lost")
		lat := r.FormValue("lat")
		lon := r.FormValue("lon")
		img := r.FormValue("img")
		tokenString := r.FormValue("token")

		if tokenString != "" {
			// Выполняем проверку токена и извелечение данных
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Сюда можно добавить дополнительные проверки токена
				return []byte(SECRET), nil
			})

			// Достаём данные из токена
			payload, ok := token.Claims.(jwt.MapClaims)
			if ok && token.Valid {
				fmt.Printf("login: %v\n", payload["login"])
				fmt.Printf("password: %v\n", payload["password"])
				fmt.Printf("Действителен до (UNIX): %v\n", payload["expires_at"])
				t := int64(payload["expires_at"].(float64))
				fmt.Printf("Действителен до (timestamp): %v\n", time.Unix(t, 0))

				email := payload["login"].(string)
				password := payload["password"].(string)

				row := database.QueryRow("select email, password_ from area.Users where email=? and password_=?", email, password)

				err = row.Scan(&user.Email, &user.Password)
				if err != nil {
					fmt.Println(err)

				}
				if user.Email == email && user.Password == password {
					log.Println("Succses!")

					_, err = database.Exec("insert into area.Posts (title, sender, finder, recipient, date_, description_, lost, latitude, longitude, is_found, is_confirmed, image) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
						title, user.Email, "none", "none", time.Now().Unix(), desc, lost, lat, lon, false, false, img)

					if err != nil {
						log.Println(err)
					}

				} else {
					fmt.Fprintf(w, "Invalid login and/or password.")
					w.WriteHeader(401)
				}

			} else {
				fmt.Println(err)
			}
		} else {

			_, err = database.Exec("insert into area.Posts (title, sender, finder, recipient, date_, description_, lost, latitude, longitude, is_found, is_confirmed, image) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				title, session.Values["username"], "none", "none", time.Now().Unix(), desc, lost, lat, lon, false, false, img)

			if err != nil {
				log.Println(err)
			}
			http.Redirect(w, r, "/", 301)
		}
	} else {
		http.ServeFile(w, r, "static/add.html")
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("select * from area.Posts")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	posts := []Post{}
	var timestamp int64
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Id, &timestamp, &p.Sender, &p.Recipient, &p.Finder, &p.Title, &p.Description, &p.Lost, &p.Lat, &p.Lon, &p.IsFound, &p.IsConfirmed, &p.Image)
		if err != nil {
			fmt.Println(err)
			continue
		}
		t := time.Unix(timestamp, 0)
		p.Date = t.Format(time.UnixDate)

		posts = append(posts, p)
	}
	session, _ := store.Get(r, "session-cookie")

	/*res, ok := session.Values["is_auth"].(bool)

	if ok != false {
		// insert error handling here
		fmt.Println("Error")
	}
	data := ViewData{
		Posts:  posts,
		IsAuth: res,
		Name:   user.Name,
	}*/
	data := make(map[string]interface{})
	data["Posts"] = posts
	data["IsAuth"] = session.Values["is_auth"]
	data["Name"] = session.Values["username"]
	data["userID"] = session.Values["user_id"]

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, data)
}
func ApiIndexHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := database.Query("select * from area.Posts")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	posts := []Post{}
	var timestamp int64
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Id, &timestamp, &p.Sender, &p.Recipient, &p.Finder, &p.Title, &p.Description, &p.Lost, &p.Lat, &p.Lon, &p.IsFound, &p.IsConfirmed, &p.Image)
		if err != nil {
			fmt.Println(err)
			continue
		}
		t := time.Unix(timestamp, 0)
		p.Date = t.Format(time.UnixDate)

		posts = append(posts, p)
	}
	session, _ := store.Get(r, "session-cookie")

	/*res, ok := session.Values["is_auth"].(bool)

	if ok != false {
		// insert error handling here
		fmt.Println("Error")
	}
	data := ViewData{
		Posts:  posts,
		IsAuth: res,
		Name:   user.Name,
	}*/
	data := make(map[string]interface{})
	data["Posts"] = posts
	data["IsAuth"] = session.Values["is_auth"]
	data["Name"] = session.Values["username"]
	data["userID"] = session.Values["user_id"]

	response, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Указываем тип контента json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Статус код 200
	w.WriteHeader(http.StatusCreated)
	// Записываем байты в ответ
	w.Write(response)
}
func ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("test")
	//vars := mux.Vars(r)
	//id := vars["id"]

	//var username string

	if r.Method == "POST" {
		id := r.FormValue("id")
		tokenString := r.FormValue("token")
		// Выполняем проверку токена и извелечение данных
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Сюда можно добавить дополнительные проверки токена
			return []byte(SECRET), nil
		})

		// Достаём данные из токена
		payload, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			fmt.Printf("login: %v\n", payload["login"])
			fmt.Printf("password: %v\n", payload["password"])
			fmt.Printf("Действителен до (UNIX): %v\n", payload["expires_at"])
			t := int64(payload["expires_at"].(float64))
			fmt.Printf("Действителен до (timestamp): %v\n", time.Unix(t, 0))

			email := payload["login"].(string)
			password := payload["password"].(string)

			row := database.QueryRow("select email, password_ from area.Users where email=? and password_=?", email, password)

			err = row.Scan(&user.Email, &user.Password)
			if err != nil {
				fmt.Println(err)

			}
			if user.Email == email && user.Password == password {
				log.Println("Succses!")

				//username = user.Email
				row := database.QueryRow("select sender from area.Posts where id=?", id)
				var sender string
				err := row.Scan(&sender)
				//log.Println(sender == usr.Email)
				if err != nil {
					fmt.Println(err)

				}
				row = database.QueryRow("select is_confirmed from area.Posts where id=?", id)
				var confirmed bool
				err = row.Scan(&confirmed)
				if err != nil {
					fmt.Println(err)

				}
				row = database.QueryRow("select is_found from area.Posts where id=?", id)
				var found bool
				err = row.Scan(&found)
				if err != nil {
					fmt.Println(err)

				}
				row = database.QueryRow("select finder from area.Posts where id=?", id)
				var finder string
				err = row.Scan(&finder)
				if err != nil {
					fmt.Println(err)

				}
				//log.Println(finder)
				state := true

				if sender == email && confirmed != true && found == true {

					_, err := database.Exec("update area.Posts set is_confirmed=? where id = ?", state, id)

					if err != nil {
						log.Println(err)
					}
					log.Println("Пользователь ", email, " подтвердил нахойженную вещь под номером ", id, "(Моб. прил.)")
					_, err = database.Exec("update area.Users set rating=rating + 1 where email = ?", finder)
					log.Println("Пользователь ", finder, " получает +1 к рейтингу (Моб. прил.)")

					if err != nil {
						log.Println(err)
					}

				}
				confirmed = false
				http.Redirect(w, r, "/", 301)

			} else {
				fmt.Fprintf(w, "Invalid login and/or password.")
				w.WriteHeader(401)
			}

		} else {
			fmt.Println(err)
		}
	} else {
		id := r.URL.Query().Get("id")
		//username = session.Values["username"].(string)

		//usr := user

		// Получаем сессию или создаём, если её не существует
		session, _ := store.Get(r, "session-cookie")
		row := database.QueryRow("select sender from area.Posts where id=?", id)
		var sender string
		err := row.Scan(&sender)
		//log.Println(sender == usr.Email)
		if err != nil {
			fmt.Println(err)

		}
		row = database.QueryRow("select is_confirmed from area.Posts where id=?", id)
		var confirmed bool
		err = row.Scan(&confirmed)
		if err != nil {
			fmt.Println(err)

		}
		row = database.QueryRow("select is_found from area.Posts where id=?", id)
		var found bool
		err = row.Scan(&found)
		if err != nil {
			fmt.Println(err)

		}
		row = database.QueryRow("select finder from area.Posts where id=?", id)
		var finder string
		err = row.Scan(&finder)
		if err != nil {
			fmt.Println(err)

		}
		//log.Println(finder)
		state := true

		if sender == session.Values["username"] && confirmed != true && found == true {

			_, err := database.Exec("update area.Posts set is_confirmed=? where id = ?", state, id)
			log.Println("Пользователь ", session.Values["username"], " подтвердил нахойженную вещь под номером ", id, "(Браузер)")
			if err != nil {
				log.Println(err)
			}
			_, err = database.Exec("update area.Users set rating=rating + 1 where email = ?", finder)
			log.Println("Пользователь ", finder, " получает +1 к рейтингу (Браузер)")
			if err != nil {
				log.Println(err)
			}
		}
		confirmed = false
		http.Redirect(w, r, "/", 301)
	}
}

func FoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("test")
	//vars := mux.Vars(r)
	//id := vars["id"]
	//id := r.URL.Query().Get("id")
	//tokenString := r.FormValue("token")
	//tokenString := r.FormValue("token")
	//var username string
	//usr := user
	// Получаем сессию или создаём, если её не существует

	if r.Method == "POST" {
		id := r.FormValue("id")
		tokenString := r.FormValue("token")
		// Выполняем проверку токена и извелечение данных
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Сюда можно добавить дополнительные проверки токена
			return []byte(SECRET), nil
		})

		// Достаём данные из токена
		payload, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			fmt.Printf("login: %v\n", payload["login"])
			fmt.Printf("password: %v\n", payload["password"])
			fmt.Printf("Действителен до (UNIX): %v\n", payload["expires_at"])
			t := int64(payload["expires_at"].(float64))
			fmt.Printf("Действителен до (timestamp): %v\n", time.Unix(t, 0))

			email := payload["login"].(string)
			password := payload["password"].(string)

			row := database.QueryRow("select email, password_ from area.Users where email=? and password_=?", email, password)

			err = row.Scan(&user.Email, &user.Password)
			if err != nil {
				fmt.Println(err)

			}
			if user.Email == email && user.Password == password {
				log.Println("Succses!")

				//username = user.Email
				row := database.QueryRow("select sender from area.Posts where id=?", id)
				var sender string
				err := row.Scan(&sender)
				//log.Println(sender == usr.Email)
				if err != nil {
					fmt.Println(err)

				}

				//log.Println(finder)
				state := true
				log.Println(email)
				if sender != email {
					_, err = database.Exec("update area.Posts set  is_found=? where id = ?", state, id)

					if err != nil {
						log.Println(err)
					}
					_, err = database.Exec("update area.Posts set finder=? where id = ?", email, id)

					if err != nil {
						log.Println(err)
					}
					log.Println("Пользователь ", email, " нашёл вещь под номером ", id, "(Моб. прил.)")

				}

			} else {
				fmt.Fprintf(w, "Invalid login and/or password.")
				w.WriteHeader(401)
			}

		} else {
			fmt.Println(err)
		}
	} else {
		id := r.URL.Query().Get("id")

		session, _ := store.Get(r, "session-cookie")
		//username = session.Values["username"].(string)
		row := database.QueryRow("select sender from area.Posts where id=?", id)
		var sender string
		err := row.Scan(&sender)
		//log.Println(sender == usr.Email)
		if err != nil {
			fmt.Println(err)

		}

		//log.Println(finder)
		state := true

		if sender != session.Values["username"] {
			_, err = database.Exec("update area.Posts set  is_found=? where id = ?", state, id)

			if err != nil {
				log.Println(err)
			}
			_, err = database.Exec("update area.Posts set finder=? where id = ?", session.Values["username"], id)
			log.Println("Пользователь ", session.Values["username"], " нашёл вещь под номером ", id, "(Браузер)")
			if err != nil {
				log.Println(err)
			}

		}
		http.Redirect(w, r, "/", 301)
	}

	// Добавляем данные в сессию

}
func RatingHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("select name_, lastname, rating from AREA.Users WHERE rating > 0 ORDER BY rating DESC ")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		usr := User{}
		err := rows.Scan(&usr.Name, &usr.Lastname, &usr.Rating)
		if err != nil {
			fmt.Println(err)
			continue
		}

		users = append(users, usr)
	}
	tmpl, _ := template.ParseFiles("templates/rating.html")
	tmpl.Execute(w, users)
}
func ApiRatingHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("select name_, lastname, rating from AREA.Users WHERE rating > 0 ORDER BY rating DESC ")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		usr := User{}
		err := rows.Scan(&usr.Name, &usr.Lastname, &usr.Rating)
		if err != nil {
			fmt.Println(err)
			continue
		}

		users = append(users, usr)
	}
	response, err := json.Marshal(users)
	if err != nil {
		panic(err)
	}

	// Указываем тип контента json
	w.Header().Set("Content-Type", "application/json")
	// Статус код 200
	w.WriteHeader(http.StatusCreated)
	// Записываем байты в ответ
	w.Write(response)

}

//var store = sessions.NewCookieStore([]byte("SESSION_KEY"))

// Проверяет существует ли сессия
func isSessionExist(r *http.Request) bool {
	session, _ := store.Get(r, "session-cookie")
	return !session.IsNew
}

func loginСheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.FormValue("token")
		if tokenString == "" {

			if !isSessionExist(r) && r.URL.Path != "/login" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if isSessionExist(r) && r.URL.Path == "/login" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			next(w, r)
		} else {
			next(w, r)
		}
	}
}
func handleRequest(w http.ResponseWriter, r *http.Request) {

	u, p, ok := r.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic auth")
		w.WriteHeader(401)
		return
	}
	row := database.QueryRow("select email, password_ from area.Users where email=? and password_=?", u, p)

	err := row.Scan(&user.Email, &user.Password)
	if err != nil {
		fmt.Println(err)

	}
	if user.Email == u && user.Password == p {
		log.Println("Succses!")

		fmt.Printf("Username: %s\n", u)
		fmt.Printf("Password: %s\n", p)
		// Определяем время жизни токена +24 часа от момента создания
		tokeExpiresAt := time.Now().Add(time.Hour * time.Duration(24))

		// Заполняем данными полезную нагрузку
		payload := jwt.MapClaims{
			"login":      u,
			"password":   p,
			"expires_at": tokeExpiresAt.Unix(), // Не обязательно
		}

		// Создаём токен с методом шифрования HS256
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

		// Подписываем токен секретным ключом
		tokenString, err := token.SignedString([]byte(SECRET))

		fmt.Printf("Токен: %v\n", tokenString)
		fmt.Printf("Действителен до: %v\n", tokeExpiresAt)
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Fprintf(w, tokenString)
		w.WriteHeader(200)
		return

	} else {
		fmt.Fprintf(w, "Invalid login and/or password.")
		w.WriteHeader(401)
		return
	}
	/*if u != username {
		fmt.Printf("Username provided is correct: %s\n", u)
		w.WriteHeader(401)
		return
	}
	if p != password {
		fmt.Printf("Password provided is correct: %s\n", u)

	}*/

}

/*func main() {

	//data := []Post{Post{Title: "Пропала шапка брата", Description: "Маленькая шапка", Lost: "В школе", Lat: 44.979323, Lon: 34.096847}, Post{Title: "Пропала моя карточка", Description: "Карта банка РНКБ", Lost: "В автобусе", Lat: 44.979323, Lon: 34.096847}}

	// Create client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://admin:Xookauf3Ow@127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Create connect
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	posts_collection := client.Database("test").Collection("posts")
	users_collection := client.Database("test").Collection("users")
	results := []Post{}
	router := mux.NewRouter()
	router.HandleFunc("/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		number, err := strconv.Atoi(id)
		if err != nil {
			// insert error handling here
		}
		data := results[number-1]
		isFound := r.URL.Query().Get("is_found")
		if isFound != "" {
			filter := bson.D{{data.Title}}

			update := bson.D{
				{"$set", bson.D{
					{"isfound", 1},
				}},
			}
			updateResult, err := posts_collection.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
		}
		tmpl, _ := template.ParseFiles("templates/post.html")
		tmpl.Execute(w, data)

	})
	http.Handle("/post", router)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		results = nil
		// Pass these options to the Find method
		options := options.Find()
		// options.SetLimit(2)
		filter := bson.M{}

		// Passing nil as the filter matches all documents in the collection
		cur, err := posts_collection.Find(context.TODO(), filter, options)
		if err != nil {
			log.Fatal(err)
		}

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(context.TODO()) {

			// create a value into which the single document can be decoded
			var elem Post
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(elem)
			results = append(results, elem)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		// Close the cursor once finished
		cur.Close(context.TODO())
		fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)

		//fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
		tmpl, _ := template.ParseFiles("templates/index.html")
		tmpl.Execute(w, results)
	})
	http.HandleFunc("/position", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/position.html")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			name := r.FormValue("name")
			lastname := r.FormValue("lastname")
			password := r.FormValue("password")
			email := r.FormValue("email")
			vk := r.FormValue("vk")
			telegram := r.FormValue("telegram")

			user := User{Name: name, Lastname: lastname, Password: password, Email: email, Vk: vk, Telegram: telegram}
			// добавляем один объект
			insertResult, err := users_collection.InsertOne(context.TODO(), user)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			http.Redirect(w, r, "/", 301)
		} else {
			http.ServeFile(w, r, "static/register.html")
		}
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {

			email := r.FormValue("email")
			password := r.FormValue("password")

			// create a value into which the result can be decoded
			filter := bson.D{{"email", email}, {"password", password}}
			var result User

			err = users_collection.FindOne(context.TODO(), filter).Decode(&result)
			if err != nil {
				//log.Fatal(err)
				fmt.Fprint(w, "Неправильный логин и/или пароль")
			}

			fmt.Printf("Found a single document: %+v\n", result)


			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			http.Redirect(w, r, "/", 301)
		} else {
			http.ServeFile(w, r, "static/login.html")
		}
	})
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			title := r.FormValue("title")
			desc := r.FormValue("desc")
			lost := r.FormValue("lost")
			lat := r.FormValue("lat")
			lon := r.FormValue("lon")

			//fmt.Fprintf(w, "Имя: %s Возраст: %s", name, age)

			// открываем соединение
			latitude, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				// insert error handling here
			}
			longitude, err := strconv.ParseFloat(lon, 64)
			if err != nil {
				// insert error handling here
			}
			post := Post{Title: title, Description: desc, Lost: lost, Lat: latitude, Lon: longitude}
			// добавляем один объект
			insertResult, err := posts_collection.InsertOne(context.TODO(), post)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Inserted a single document: ", insertResult.InsertedID)
			http.Redirect(w, r, "/", 301)
		} else {
			http.ServeFile(w, r, "static/add.html")
		}
	})
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	cur.Close(context.TODO())
	fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, results)
	//})
	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}*/
