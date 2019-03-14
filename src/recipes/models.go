package recipes

import (
	"strconv"

	// External imports
	"gopkg.in/couchbase/gocb.v1"
)

const (
	lockTime = 3 // seconds
)

// The Recipe entity is used to marshall/unmarshall JSON.
type Recipe struct {
	Name       string  `json:"name"`
	PrepTime   float32 `json:"preptime"`
	Difficulty int     `json:"difficulty"`
	Vegetarian bool    `json:"vegetarian"`
	Ratings    []int   `json:"ratings"`
}

// The N1qlRecipe entity is used to retrieve query data from Couchbase.
type N1qlRecipe struct {
	ID     string `json:"id"`
	Recipe Recipe `json:"recipe"`
}

// The RecipeRated entity is used to marshall/unmarshall JSON.
type RecipeRated struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PrepTime   float32 `json:"preptime"`
	Difficulty int     `json:"difficulty"`
	Vegetarian bool    `json:"vegetarian"`
	AvgRating  float32 `json:"avg_rating"`
}

// The RecipeRating entity is used to marshall/unmarshall JSON.
type RecipeRating struct {
	RecipeID int `json:"recipe_id"`
	Rating   int `json:"rating"`
}

// GetRecipe returns a single specified recipe.
func (r *Recipe) GetRecipe(id string, db *gocb.Bucket) error {
	_, err := db.Get(id, r)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRecipe is used to modify a specific recipe.
// The recipe ratings will not be changed.
func (r *Recipe) UpdateRecipe(id string, db *gocb.Bucket) error {

	var recipe Recipe

	// Get document, lock for specified number of seconds
	cas, err := db.GetAndLock(id, lockTime, &recipe)
	if err != nil {
		return err
	}

	recipe.Name = r.Name
	recipe.PrepTime = r.PrepTime
	recipe.Difficulty = r.Difficulty
	recipe.Vegetarian = r.Vegetarian

	// Mutating unlocks the document
	_, err = db.Replace(id, recipe, cas, 0)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecipe is used to delete a specific recipe.
func (r *Recipe) DeleteRecipe(id string, db *gocb.Bucket) error {
	_, err := db.Remove(id, 0)
	if err != nil {
		return err
	}
	return nil
}

// CreateRecipe is used to create a single recipe.
func (r *Recipe) CreateRecipe(db *gocb.Bucket) error {

	// For automatically getting the next sequence number:
	// increment by 1, initialize at 1 if counter not found,
	// do not expire (set expiry to 0). Returns uint64, Cas, error.
	newID, _, err := db.Counter("idGeneratorForRecipes", 1, 1, 0)
	if err != nil {
		return err
	}

	rID := int(newID)

	id := strconv.Itoa(rID)

	_, err = db.Insert(id, r, 0)
	if err != nil {
		return err
	}
	return nil
}

// GetRecipes returns a collection of known recipes.
func GetRecipes(db *gocb.Bucket, start int, count int) ([]N1qlRecipe, error) {

	getRecipesN1ql := "SELECT META().id, * FROM recipes AS recipe LIMIT $1 OFFSET $2"
	getRecipesQuery := gocb.NewN1qlQuery(getRecipesN1ql).AdHoc(false)

	var params []interface{}
	params = append(params, count)
	params = append(params, start)

	rows, err := db.ExecuteN1qlQuery(getRecipesQuery, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := []N1qlRecipe{}

	var row N1qlRecipe

	for rows.Next(&row) {
		recipes = append(recipes, row)
		row = N1qlRecipe{}
	}
	return recipes, nil
}

// GetRecipesRated returns a collection of rated recipes.
func GetRecipesRated(db *gocb.Bucket, start int, count int, preptime float32) ([]RecipeRated, error) {

	listRecipesN1ql := "SELECT META().id, * FROM recipes AS recipe WHERE preptime < $3 LIMIT $1 OFFSET $2"
	listRecipesQuery := gocb.NewN1qlQuery(listRecipesN1ql).AdHoc(false)

	var params []interface{}
	params = append(params, count)
	params = append(params, start)
	params = append(params, preptime)

	rows, err := db.ExecuteN1qlQuery(listRecipesQuery, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipesRated := []RecipeRated{}

	var row N1qlRecipe

	for rows.Next(&row) {
		recipeRated := RecipeRated{}
		recipeRated.ID = row.ID
		recipeRated.Name = row.Recipe.Name
		recipeRated.PrepTime = row.Recipe.PrepTime
		recipeRated.Difficulty = row.Recipe.Difficulty
		recipeRated.Vegetarian = row.Recipe.Vegetarian
		var avgRating float32
		lenRatings := len(row.Recipe.Ratings)
		if lenRatings > 0 {
			total := 0
			for _, r := range row.Recipe.Ratings {
				total += r
			}
			avgRating = float32(total) / float32(lenRatings)
		}
		recipeRated.AvgRating = avgRating
		recipesRated = append(recipesRated, recipeRated)
		row = N1qlRecipe{}
	}
	return recipesRated, nil
}

// AddRecipeRating adds a rating for a specific recipe.
// There can be many ratings for any specific recipe
// and the ratings are never overwritten.
func (rr *RecipeRating) AddRecipeRating(db *gocb.Bucket) error {

	id := strconv.Itoa(int(rr.RecipeID))

	var recipe Recipe

	// Get document, lock for specified number of seconds
	cas, err := db.GetAndLock(id, lockTime, &recipe)
	if err != nil {
		return err
	}

	ratings := recipe.Ratings
	ratings = append(ratings, rr.Rating)
	recipe.Ratings = ratings

	// Mutating unlocks the document
	_, err = db.Replace(id, recipe, cas, 0)
	if err != nil {
		return err
	}

	return nil
}
