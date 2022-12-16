package handlers

import (
	"fmt"
	"github.com/cschloer/tischtennis/common/database"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"sort"
	"strconv"
)

type IndexPageData struct {
	Version          string
	Title            string
	People           []database.Person
	AlphSortedPeople []database.Person
}

type PersonPageData struct {
	Version string
	Person  database.Person
	People  []database.Person
	Games   []database.Game
}

type AdminPageData struct {
	Version string
	People  []database.Person
	Games   []database.Game
}

var version = "1.0"

var IndexTemplate = template.Must(
	template.New("").Funcs(template.FuncMap{
		"mul": Mul,
	}).ParseFiles("templates/index.html", "templates/base.html"))
var PersonTemplate = template.Must(
	template.New("").Funcs(template.FuncMap{
		"mul": Mul,
	}).ParseFiles("templates/person.html", "templates/base.html"))
var AdminTemplate = template.Must(
	template.New("").Funcs(template.FuncMap{
		"mul": Mul,
	}).ParseFiles("templates/admin.html", "templates/base.html"))

func Mul(param1 float64, param2 float64) string {
	return fmt.Sprintf("%.2f", param1*param2)
}

func IndexPageHandler(response http.ResponseWriter, request *http.Request) {

	people, err := database.GetPeople()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	alphSortedPeople := make([]database.Person, len(people))
	copy(alphSortedPeople, people)
	sort.Slice(alphSortedPeople, func(i, j int) bool {
		return alphSortedPeople[i].Name < alphSortedPeople[j].Name
	})
	data := IndexPageData{
		Version:          version,
		Title:            "Tischtennis",
		People:           people,
		AlphSortedPeople: alphSortedPeople,
	}
	IndexTemplate.ExecuteTemplate(response, "base", data)
}

func PersonPageHandler(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	personIdStr := vars["personId"]
	personId, err := strconv.Atoi(personIdStr)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	person, err := database.GetPerson(personId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	people, err := database.GetPeople()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	alphSortedPeople := make([]database.Person, len(people))
	copy(alphSortedPeople, people)
	sort.Slice(alphSortedPeople, func(i, j int) bool {
		return alphSortedPeople[i].Name < alphSortedPeople[j].Name
	})

	games, err := database.GetPersonGames(personId, 10)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	data := PersonPageData{
		Version: version,
		Person:  person,
		People:  alphSortedPeople,
		Games:   games,
	}
	PersonTemplate.ExecuteTemplate(response, "base", data)
}

func AdminPageHandler(response http.ResponseWriter, request *http.Request) {
	/*
		err := database.CreateDatabase()
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = database.ComputeScores()
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
	*/
	people, err := database.GetPeople()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	idSortedPeople := make([]database.Person, len(people))
	copy(idSortedPeople, people)
	sort.Slice(idSortedPeople, func(i, j int) bool {
		return idSortedPeople[i].Id < idSortedPeople[j].Id
	})
	games, err := database.GetGames(100)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	data := AdminPageData{
		Version: version,
		People:  idSortedPeople,
		Games:   games,
	}
	AdminTemplate.ExecuteTemplate(response, "base", data)
}
