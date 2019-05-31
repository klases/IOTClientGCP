package gcpfunc

import (
	"fmt"
	"net/http"
)

/*
To deploy run:
gcloud functions deploy HelloGet --runtime go111 --trigger-http
*/

func HelloGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
