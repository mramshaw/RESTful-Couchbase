package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
	// local import
	"application"
	// external import
	"gopkg.in/couchbase/gocb.v1"
)

var app application.App

func TestMain(m *testing.M) {
	app = application.App{}
	app.Initialize(
		os.Getenv("COUCHBASE_USER"),
		os.Getenv("COUCHBASE_PASS"),
		os.Getenv("COUCHBASE_DB"))
	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func TestEmptyTables(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentRecipe(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes/11", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Recipe not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Recipe not found'. Got '%s'", m["error"])
	}
}

func TestCreateRecipe(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test recipe" {
		t.Errorf("Expected recipe name to be 'test recipe'. Got '%v'", m["name"])
	}

	if m["preptime"] != 0.1 {
		t.Errorf("Expected recipe price to be '0.1'. Got '%v'", m["preptime"])
	}

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["difficulty"] != 2.0 {
		t.Errorf("Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])
	}

	if m["vegetarian"] != true {
		t.Errorf("Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])
	}
}

func TestGetRecipe(t *testing.T) {
	clearTables()
	addRecipes(1, 1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)
}

func TestUpdatePutRecipe(t *testing.T) {
	clearTables()
	addRecipes(1, 1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - put","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PUT", "/v1/recipes/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PUT): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] == originalRecipe["name"] {
		t.Errorf("Expected the name to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["name"], m["name"])
	}
	if m["preptime"] == originalRecipe["preptime"] {
		t.Errorf("Expected the price to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["preptime"], m["preptime"])
	}
	if m["difficulty"] == originalRecipe["difficulty"] {
		t.Errorf("Expected the difficulty to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["difficulty"], m["difficulty"])
	}
	if m["vegetarian"] == originalRecipe["vegetarian"] {
		t.Errorf("Expected vegetarian to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["vegetarian"], m["vegetarian"])
	}
}

func TestUpdatePatchRecipe(t *testing.T) {
	clearTables()
	addRecipes(1, 1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - patch","preptime":22.22,"difficulty":4,"vegetarian":false}`)

	req, err = http.NewRequest("PATCH", "/v1/recipes/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PATCH): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] == originalRecipe["name"] {
		t.Errorf("Expected the name to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["name"], m["name"])
	}
	if m["preptime"] == originalRecipe["preptime"] {
		t.Errorf("Expected the price to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["preptime"], m["preptime"])
	}
	if m["difficulty"] == originalRecipe["difficulty"] {
		t.Errorf("Expected the difficulty to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["difficulty"], m["difficulty"])
	}
	if m["vegetarian"] == originalRecipe["vegetarian"] {
		t.Errorf("Expected vegetarian to change from '%[01]v' to '%[02]v'. Got '%[02]v'", originalRecipe["vegetarian"], m["vegetarian"])
	}
}

func TestDeleteRecipe(t *testing.T) {
	clearTables()
	addRecipes(1, 1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	req, err = http.NewRequest("DELETE", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (DELETE): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	req, err = http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (Second GET): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected int, response *httptest.ResponseRecorder) {
	actual := response.Code
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
		log.Println("checkResponseCode: Response body = ", response.Body)
	}
}

func ensureTablesExist() {
	err := app.Manager.CreatePrimaryIndex("", true, false)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	if err := app.Manager.Flush(); err != nil {
		log.Fatal(err)
	}
	ensureTablesExist()
}

func TestAddRating(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (1st POST): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (2nd POST): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)
}

func TestSearch(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (1st POST): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (2nd POST): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)

	payload = []byte(`{"rating":2}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (3rd POST): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response)

	// Sleep for 2 seconds to allow Couchbase time to commit
	time.Sleep(2 * time.Second)

	var bb bytes.Buffer
	mw := multipart.NewWriter(&bb)
	mw.WriteField("count", "1")
	mw.WriteField("start", "0")
	mw.WriteField("preptime", "50.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (4th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response)

	var m map[string]interface{}
	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)
	if len(mm) == 0 {
		t.Errorf("Expected results, but got an empty resultset")
	} else {
		// we only want the first result
		m = mm[0]

		if m["name"] != "test recipe" {
			t.Errorf("Expected recipe name to be 'test recipe'. Got '%v'", m["name"])
		}

		if m["preptime"] != 0.1 {
			t.Errorf("Expected recipe price to be '0.1'. Got '%v'", m["preptime"])
		}

		// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
		//     floats (float64), when the target is a map[string]interface{}
		if m["difficulty"] != 2.0 {
			t.Errorf("Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])
		}

		if m["vegetarian"] != true {
			t.Errorf("Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])
		}

		if m["avg_rating"] != 2.5 {
			t.Errorf("Expected average recipe rating to be '2.5'. Got '%v'", m["id"])
		}
	}

	addRecipes(2, 12)

	// Sleep for 2 seconds to allow Couchbase time to commit
	time.Sleep(2 * time.Second)

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (5th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 10 {
		t.Errorf("Expected '10' recipes. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.WriteField("preptime", "30.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (6th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 2 {
		t.Errorf("Expected '2' recipes. Got '%v'", len(mm))
	}
}

func addRecipes(start, count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {

		insertRecipe := gocb.NewN1qlQuery("INSERT INTO recipes (KEY, VALUE) VALUES ($1, {'name':$2,'preptime':$3,'difficulty':$4,'vegetarian':$5})")

		id := strconv.Itoa(start)
		nameID := strconv.Itoa(i + 1)
		var params []interface{}
		params = append(params, id)
		params = append(params, "Recipe "+nameID)
		params = append(params, (i+1.0)*10)
		params = append(params, i%3+1)
		params = append(params, true)

		_, err := app.DB.ExecuteN1qlQuery(insertRecipe, params)
		if err != nil {
			log.Printf("Did not load recipe %v, key: %v, error : %v\n", i, start, err)
		}
		start++
	}
}
