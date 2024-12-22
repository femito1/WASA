package main

import (
	"fmt"
	"io"
	"net/http"
)

/*func evenRandomNumber(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	var p = rand.Int()

	if p%2 == 0 {
		fmt.Fprintf(w, "%d is even", p)
	} else {
		fmt.Fprintf(w, "%d is odd", p)
	}
}
*/

func hi(w http.ResponseWriter, r *http.Request) {
	// name := r.URL.Query().Get("name")
	body, _ := io.ReadAll(r.Body)

	name := string(body)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hello %s", name)

}

func main() {
	http.HandleFunc("/", hi)
	fmt.Println("Starting web sever at http://localhost:8090")
	http.ListenAndServe(":8090", nil)

}
