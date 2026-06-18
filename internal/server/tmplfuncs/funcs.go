package tmplfuncs

import (
	"fmt"
	"html/template"
	"reflect"
)

// FuncMap returns the custom template functions used in HTML templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"dict":        dict,
		"seq":         seq,
		"add":         add,
		"subtract":    subtract,
		"eq_int":      eqInt,
		"proficiencies": proficiencies,
		"signedCost":  signedCost,
	}
}

// signedCost formats a perk's add cost from the player's perspective.
// Positive add cost = you spend points = shown as "-X pts" (you lose them).
// Negative add cost = you gain points back = shown as "+X pts" (refund).
func signedCost(n int) string {
	if n > 0 {
		return fmt.Sprintf("-%d pts", n)
	} else if n < 0 {
		return fmt.Sprintf("+%d pts", -n)
	}
	return "0 pts"
}

// proficiencies returns all purchasable proficiency level names and their int values.
func proficiencies() []map[string]interface{} {
	levels := []struct {
		Name string
		Val  int
	}{
		{"Clumsy", 0},
		{"Untrained", 1},
		{"Trained", 2},
		{"Expert", 3},
		{"Master", 4},
	}
	result := make([]map[string]interface{}, len(levels))
	for i, l := range levels {
		result[i] = map[string]interface{}{
			"Name": l.Name,
			"Val":  l.Val,
		}
	}
	return result
}

// eqInt compares two integers for equality (works with Proficiency int type).
func eqInt(a, b interface{}) bool {
	ai := toInt(a)
	bi := toInt(b)
	return ai == bi
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	default:
		// Handle named int types (like models.Proficiency)
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Int || rv.Kind() == reflect.Int8 || rv.Kind() == reflect.Int16 || rv.Kind() == reflect.Int32 || rv.Kind() == reflect.Int64 {
			return int(rv.Int())
		}
		return -999
	}
}

// dict creates a map from alternating key/value pairs for passing to templates.
func dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict requires even number of arguments")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m
}

// seq generates a sequence of integers from 0 to n-1.
func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// add returns a + b.
func add(a, b int) int {
	return a + b
}

// subtract returns a - b.
func subtract(a, b int) int {
	return a - b
}
