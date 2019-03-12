package application

import (
	// native packages
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	// local packages
	"recipes"
	// external packages
	"github.com/gorilla/mux"
	"gopkg.in/couchbase/gocb.v1"
)

// App represents the application
type App struct {
	Router  *mux.Router
	Manager *gocb.BucketManager
	DB      *gocb.Bucket
}

func (a *App) getRecipeEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	recipeID := params["id"]
	r := recipes.Recipe{}
	if err := r.GetRecipe(recipeID, a.DB); err != nil {
		if gocb.IsKeyNotFoundError(err) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) getRecipesEndpoint(w http.ResponseWriter, req *http.Request) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}
	recipes, err := recipes.GetRecipes(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, recipes)
}

func (a *App) createRecipeEndpoint(w http.ResponseWriter, req *http.Request) {
	var r recipes.Recipe
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := r.CreateRecipe(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, r)
}

func (a *App) modifyRecipeEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	recipeID := params["id"]
	var r recipes.Recipe
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := r.UpdateRecipe(recipeID, a.DB); err != nil {
		if gocb.IsKeyNotFoundError(err) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) deleteRecipeEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	recipeID := params["id"]
	r := recipes.Recipe{}
	if err := r.DeleteRecipe(recipeID, a.DB); err != nil {
		if gocb.IsKeyNotFoundError(err) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) addRatingEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	recipeID, err := strconv.Atoi(params["recipe_id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}
	rr := recipes.RecipeRating{RecipeID: recipeID}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&rr); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := rr.AddRecipeRating(a.DB); err != nil {
		if gocb.IsKeyNotFoundError(err) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respondWithJSON(w, http.StatusCreated, rr)
}

func (a *App) searchRecipesEndpoint(w http.ResponseWriter, req *http.Request) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	var preptime32 float32
	if req.FormValue("preptime") == "" {
		preptime32 = 9999.99 // random large value
	} else {
		preptime64, _ := strconv.ParseFloat(req.FormValue("preptime"), 32)
		preptime32 = float32(preptime64)
	}

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	recipesRated, err := recipes.GetRecipesRated(a.DB, start, count, preptime32)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, recipesRated)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}

// Initialize sets up the database connection, router, and routes for the app
func (a *App) Initialize(user, password, dbname string) {

	cluster, err := gocb.Connect("couchbase://couchbase")
	if err != nil {
		log.Fatal("Failed to connect to couchbase: ", err)
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: user,
		Password: password,
	})
	a.DB, err = cluster.OpenBucket(dbname, "")
	if err != nil {
		log.Fatal("Failed to get bucket from couchbase: ", err)
	}
	a.Manager = a.DB.Manager(user, password)

	a.Router = mux.NewRouter()

	v1 := a.Router.PathPrefix("/v1").Subrouter()

	v1.HandleFunc("/recipes", a.getRecipesEndpoint).Methods("GET")
	v1.HandleFunc("/recipes", a.createRecipeEndpoint).Methods("POST")
	v1.HandleFunc("/recipes/{id:[0-9]+}", a.getRecipeEndpoint).Methods("GET")
	v1.HandleFunc("/recipes/{id:[0-9]+}", a.modifyRecipeEndpoint).Methods("PUT")
	v1.HandleFunc("/recipes/{id:[0-9]+}", a.modifyRecipeEndpoint).Methods("PATCH")
	v1.HandleFunc("/recipes/{id:[0-9]+}", a.deleteRecipeEndpoint).Methods("DELETE")
	v1.HandleFunc("/recipes/{recipe_id:[0-9]+}/rating", a.addRatingEndpoint).Methods("POST")
	v1.HandleFunc("/recipes/search", a.searchRecipesEndpoint).Methods("POST")
}

// Run starts the app and serves on the specified port
func (a *App) Run(port string) {
	log.Print("Now serving recipes ...")
	log.Fatal(http.ListenAndServe(":"+port, a.Router))
}
