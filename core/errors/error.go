package errors

type FieldError struct {
	FieldName string
	Message   string
}

type Error struct {
	StatusCode  int
	ErrorCode   string
	Message     string
	FieldErrors []FieldError
}
