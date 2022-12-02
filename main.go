package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	// Route untuk menginisialisasi folder public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/project", project).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/project-detail/{id}", projectDetail).Methods("GET")
	route.HandleFunc("/add-project", addProjects).Methods("POST")
	route.HandleFunc("/delete-project/{index}", deleteProjects).Methods("GET")
	route.HandleFunc("/edit-project/{in}", editProject).Methods("GET")
	route.HandleFunc("/edit-project/{in}", formEditProject).Methods("POST")

	route.HandleFunc("/register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")

	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server sedang berjalan pada port 5000")
	http.ListenAndServe("localhost:5000", route)
}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type Project struct {
	TitleSessions string
	IsLogin       bool
	UserName      string
	FlashData     string

	Id              int
	Title           string
	Description     string
	Technologies    []string
	StartDate       time.Time
	EndDate         time.Time
	Duration        string
	FormatStartDate string
	FormatEndDate   string
	NodeJs          string
	ReactJs         string
	JavaScript      string
	TypeScript      string
}

var Data = Project{
	TitleSessions: "Personal Web",
}

func addProjects(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	startDate := r.PostForm.Get("startdate")
	endDate := r.PostForm.Get("enddate")

	//Checkbock Technologies
	nodejs := r.PostForm.Get("nodejs")
	reactjs := r.PostForm.Get("reactjs")
	javascript := r.PostForm.Get("javascript")
	typescript := r.PostForm.Get("typescript")

	checked := []string{
		nodejs,
		reactjs,
		javascript,
		typescript,
	}

	_, errQuery := connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_projects(title, start_date, end_date, description, technologies) VALUES($1, $2, $3, $4, $5)", title, startDate, endDate, description, checked)

	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	// var newProject = Project{
	// 	Title:       title,
	// 	Description: description,
	// 	StartDate:   startDate,
	// 	EndDate:     endDate,
	// }

	// Untuk Push ke Array projects
	// projects = append(projects, newProject)
	fmt.Println(startDate)
	fmt.Println(endDate)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	dataProject, errQuery := connection.Conn.Query(context.Background(), "SELECT Id, title, start_date, end_date, description, technologies FROM public.tb_projects")

	if errQuery != nil {
		fmt.Println("Message2 : " + errQuery.Error())
		return
	}

	var result []Project

	for dataProject.Next() {
		var each = Project{}

		err := dataProject.Scan(&each.Id, &each.Title, &each.StartDate, &each.EndDate, &each.Description, &each.Technologies)
		if err != nil {
			fmt.Println("Message dataProject : " + err.Error())
			return
		}

		// Time
		diff := each.EndDate.Sub(each.StartDate)
		days := diff.Hours() / 24
		mount := math.Floor(diff.Hours() / 24 / 30)

		dy := strconv.FormatFloat(days, 'f', 0, 64)
		mo := strconv.FormatFloat(mount, 'f', 0, 64)

		if days < 30 {
			each.Duration = dy + " Days"
		} else if days > 30 {
			each.Duration = mo + " Month"
		}

		// Technologies
		each.NodeJs = ""
		each.ReactJs = ""
		each.JavaScript = ""
		each.TypeScript = ""

		if each.Technologies[0] == "true" {
			each.NodeJs = "/public/img/nodejs.png"
		}
		if each.Technologies[1] == "true" {
			each.ReactJs = "/public/img/reactjs.png"
		}
		if each.Technologies[2] == "true" {
			each.JavaScript = "/public/img/javaScript.png"
		}
		if each.Technologies[3] == "true" {
			each.TypeScript = "/public/img/typeScript.png"
		}

		each.FormatStartDate = each.StartDate.Format("2 January 2006")
		each.FormatEndDate = each.EndDate.Format("2 January 2006")

		result = append(result, each)

	}

	//sessions
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Names"].(string)
	}

	fm := session.Flashes("Message")

	var flashes []string

	if len(fm) > 0 {

		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	resData := map[string]interface{}{
		"Projects": result,
		"Data":     Data,
	}
	tmpt.Execute(w, resData)

}

func projectDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/projectDetail.html")

	if err != nil {
		w.Write([]byte("Message1 : " + err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// w.Write([]byte("Message : " + err.Error()))

	var ProjectDetail = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description FROM tb_projects WHERE id = $1", id).Scan(&ProjectDetail.Id, &ProjectDetail.Title, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message2: " + err.Error()))
	}

	ProjectDetail.FormatStartDate = ProjectDetail.StartDate.Format("2 january 2006")
	ProjectDetail.FormatEndDate = ProjectDetail.EndDate.Format("2 january 2006")

	dataDetail := map[string]interface{}{
		"Project": ProjectDetail,
	}

	tmpt.Execute(w, dataDetail)
}

// ---------------------

func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/addProject.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	tmpt.Execute(w, nil)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/editProject.html")
	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	in, _ := strconv.Atoi(mux.Vars(r)["in"])

	var EditProject = Project{}

	errQuery := connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description FROM public.tb_projects WHERE id = $1", in).Scan(&EditProject.Id, &EditProject.Title, &EditProject.StartDate, &EditProject.EndDate, &EditProject.Description)

	if errQuery != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	dataEdit := map[string]interface{}{
		"Project": EditProject,
	}

	EditProject.FormatStartDate = EditProject.StartDate.Format("2 January 2006")
	EditProject.FormatEndDate = EditProject.EndDate.Format("2 January 2006")

	tmpt.Execute(w, dataEdit)
}
func formEditProject(w http.ResponseWriter, r *http.Request) {
	in, _ := strconv.Atoi(mux.Vars(r)["in"])
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	startDate := r.PostForm.Get("startdate")
	endDate := r.PostForm.Get("enddate")

	_, errQuery := connection.Conn.Exec(context.Background(), "UPDATE public.tb_projects SET title=$1, start_date=$2, end_date=$3, description=$4 WHERE id=$5;", title, startDate, endDate, description, in)

	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProjects(w http.ResponseWriter, r *http.Request) {

	index, _ := strconv.Atoi(mux.Vars(r)["index"])

	_, errQuery := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id = $1", index)

	if errQuery != nil {
		fmt.Println("Message : " + errQuery.Error())
		return
	}

	// projects = append(projects[:index], projects[index+1:]...)

	http.Redirect(w, r, "/", http.StatusFound)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}
	tmpt.Execute(w, nil)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")

	password := r.PostForm.Get("password")
	paswordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES($1,$2,$3)", name, email, paswordHash)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email = $1", email).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	session.Values["IsLogin"] = true
	session.Values["Names"] = user.Name
	session.Options.MaxAge = 10800 //3Hours

	session.AddFlash("Sueccessfully login", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logout")

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")
	session.Options.MaxAge = -1

	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
