package node

import "net/http"

func Run(dataDir string) error {
	return http.ListenAndServe(":8080", nil)
}
