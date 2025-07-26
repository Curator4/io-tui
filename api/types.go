package api

type AIAPI interface {
	GetResponse(prompt string) (string, error)
}
