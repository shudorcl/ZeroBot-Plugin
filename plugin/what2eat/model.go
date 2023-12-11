package what2eat

// var db = &sql.Sqlite{}
// var mu sync.RWMutex

type picture struct {
	ID   uint64 `db:"id"`
	NAME string `db:"name"`
}
