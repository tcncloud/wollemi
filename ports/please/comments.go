package please

type Comments struct {
	Before []Comment
	Suffix []Comment
	After  []Comment
}

type Comment struct {
	Token string
}
