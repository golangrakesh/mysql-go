package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	bcrypt "golang.org/x/crypto/bcrypt"
)

var db *sql.DB

var err error

var tpl *template.Template

type user struct {
ID int64
Username string
FirstName string
LastName string
Password []byte
}

func init() {
	db, err = sql.Open("mysql", "root@tcp(localhost:3306)/student")
	if err != nil {
		panic(err.Error())  
	}
	
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func index(w http.ResponseWriter, req *http.Request) {
	rows, e := db.Query(
	`SELECT id,
	username,
	first_name,
	last_name,
	password
	FROM users;`)
	if e != nil {
	log.Println(e)
	http.Error(w, e.Error(), http.StatusInternalServerError)
	return
	}
	users := make([]user, 0)
	for rows.Next() {
	usr := user{}
	rows.Scan(&usr.ID, &usr.Username, &usr.FirstName, &usr.LastName, &usr.Password)
	users = append(users, usr)
	}
	log.Println(users)
	tpl.ExecuteTemplate(w, "index.gohtml", users)
}

func userForm(w http.ResponseWriter, req *http.Request) {
	err = tpl.ExecuteTemplate(w, "userForm.gohtml", nil)
	if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
	}
}

func createUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		usr := user{}
		usr.Username = req.FormValue("username")
		usr.FirstName = req.FormValue("firstName")
		usr.LastName = req.FormValue("lastName")
		bPass, e := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.MinCost)
		if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
		}
		usr.Password = bPass
		_, e = db.Exec(
		"INSERT INTO users (username, first_name, last_name, password) VALUES (?, ?, ?, ?)",
		usr.Username,
		usr.FirstName,
		usr.LastName,
		usr.Password,
		)
		if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
		}
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
		http.Error(w, "Method Not Supported", http.StatusMethodNotAllowed)
	

}

func editUsers(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	rows, err := db.Query(
	`SELECT id,
	username,
	first_name,
	last_name
	FROM users
	WHERE id = ` + id + `;`)
	if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
	}
	usr := user{}
	for rows.Next() {
	rows.Scan(&usr.ID, &usr.Username, &usr.FirstName, &usr.LastName)
	}
	tpl.ExecuteTemplate(w, "editUser.gohtml", usr)
}

func updateUsers(w http.ResponseWriter, req *http.Request) {
	_, er := db.Exec(
	"UPDATE users SET username = ?, first_name = ?, last_name = ? WHERE id = ? ",
	req.FormValue("username"),
	req.FormValue("firstName"),
	req.FormValue("lastName"),
	req.FormValue("id"),
	)
	if er != nil {
	http.Error(w, er.Error(), http.StatusInternalServerError)
	return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
}


func deleteUsers(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
	http.Error(res, "Please send ID", http.StatusBadRequest)
	return
	}
	_, er := db.Exec("DELETE FROM users WHERE id = ?", id)
	if er != nil {
	http.Error(res, er.Error(), http.StatusBadRequest)
	return
	}
	http.Redirect(res, req, "/", http.StatusSeeOther)
}


func main() {
	defer db.Close()
	http.HandleFunc("/", index)
	http.HandleFunc("/userForm", userForm)
	http.HandleFunc("/createUsers", createUsers)
	http.HandleFunc("/editUsers", editUsers)
	http.HandleFunc("/deleteUsers", deleteUsers)
	http.HandleFunc("/updateUsers", updateUsers)
	log.Println("Server is up on 8080 port")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}