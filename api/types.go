package api

type AIAPI interface {
	GetResponse(context string) (string, error)
}
