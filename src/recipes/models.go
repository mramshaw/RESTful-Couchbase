package recipes

import (
	"strconv"

	// External imports
	"gopkg.in/couchbase/gocb.v1"
)

// The Recipe entity is used to marshall/unmarshall JSON.
type Recipe struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	PrepTime   float32 `json:"preptime"`
	Difficulty int     `json:"difficulty"`
	Vegetarian bool    `json:"vegetarian"`
	Ratings    []int   `json:"ratings"`
}

// The n1qlRecipe entity is used to retrieve query data from Couchbase.
type n1qlRecipe struct {
	Recipe Recipe `json:"recipe"`
}

// The n1qlRecipe entity is used to retrieve query data from Couchbase.
type n1qlRatings struct {
	Ratings []int `json:"ratings"`
}

// The RecipeRated entity is used to marshall/unmarshall JSON.
type RecipeRated struct {
	ID         int     `json:"id"`
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
func (r *Recipe) UpdateRecipe(id string, db *gocb.Bucket) error {
	_, err := db.Upsert(id, r, 0)
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
	r.ID = rID

	id := strconv.Itoa(rID)

	_, err = db.Insert(id, r, 0)
	if err != nil {
		return err
	}
	return nil
}

// GetRecipes returns a collection of known recipes.
func GetRecipes(db *gocb.Bucket, start int, count int) ([]Recipe, error) {

	listRecipesQuery := gocb.NewN1qlQuery("SELECT * FROM recipes AS recipe LIMIT $1 OFFSET $2").AdHoc(false)

	var params []interface{}
	params = append(params, count)
	params = append(params, start)

	rows, err := db.ExecuteN1qlQuery(listRecipesQuery, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := []Recipe{}

	var row n1qlRecipe

	for rows.Next(&row) {
		recipes = append(recipes, row.Recipe)
		row = n1qlRecipe{}
	}
	return recipes, nil
}

// GetRecipesRated returns a collection of rated recipes.
func GetRecipesRated(db *gocb.Bucket, start int, count int, preptime float32) ([]RecipeRated, error) {

	listRecipesQuery := gocb.NewN1qlQuery("SELECT * FROM recipes AS recipe WHERE preptime < $3 LIMIT $1 OFFSET $2").AdHoc(false)

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

	var row n1qlRecipe

	for rows.Next(&row) {
		recipeRated := RecipeRated{}
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
		row = n1qlRecipe{}
	}
	return recipesRated, nil
}

// AddRecipeRating adds a rating for a specific recipe.
// There can be many ratings for any specific recipe
// and the ratings are never overwritten.
func (rr *RecipeRating) AddRecipeRating(db *gocb.Bucket) error {

	getRecipeQuery := gocb.NewN1qlQuery("SELECT ratings FROM recipes AS results USE KEYS $1").AdHoc(false)

	id := strconv.Itoa(int(rr.RecipeID))

	var params []interface{}
	params = append(params, id)

	rs, err := db.ExecuteN1qlQuery(getRecipeQuery, params)
	if err != nil {
		return err
	}
	defer rs.Close()

	var row n1qlRatings
	rs.One(&row)

	ratings := row.Ratings
	ratings = append(ratings, rr.Rating)

	updateRecipeQuery := gocb.NewN1qlQuery("UPDATE recipes USE KEYS $1 SET ratings = $2").AdHoc(false)

	params = append(params, ratings)

	_, err = db.ExecuteN1qlQuery(updateRecipeQuery, params)
	if err != nil {
		return err
	}

	return nil
}
