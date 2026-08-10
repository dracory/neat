package data

type Blog struct {
	ID         int64  `db:"id"`
	Title      string `db:"title"`
	CategoryID int64  `db:"category_id"`
}

type Category struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

var Blogs = []Blog{
	{ID: 1, Title: "Hello World", CategoryID: 1},
	{ID: 2, Title: "Go Tips", CategoryID: 2},
	{ID: 3, Title: "Advanced Patterns", CategoryID: 2},
}

var Categories = []Category{
	{ID: 1, Name: "General"},
	{ID: 2, Name: "Programming"},
}
