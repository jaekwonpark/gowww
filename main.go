package main

import (
	"fmt"
	"os"
	"github.com/stianeikeland/go-rpio"
	"time"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
	"log"
)

var (
	waterStations = []int{22, 23, 24}
	Garage = 25
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SESSION")))
	webkey = os.Getenv("WEBKEY")
	cert = os.Getenv("CERT_PATH")+"bundle.crt"
	certkey = os.Getenv("CERT_PATH")+"private.key"
	cookieName = "ctrl"
	sessionName = "certified"
)

func ctrl(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookieName)

	if auth, ok := session.Values[sessionName].(string); !ok || (auth != webkey) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	fmt.Fprintln(w, "you are allowed")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookieName)
	session.Values[sessionName] = webkey
	session.Save(r,w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookieName)
	session.Values[sessionName] = ""
	session.Save(r,w)
}

func main() {


	// declare mux router
	r := mux.NewRouter()

	r.HandleFunc("/ctrl", ctrl)
	//r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)




	err := http.ListenAndServeTLS(":443", cert, certkey, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}


	
	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// sprinkler
	//turnOn(waterStations, 5, 1)
	//turnOn(waterStations, 5, 1)

	// garage
	//toggle(Garage, 5)
	//toggle(Garage, 5)

}

func turnOn(pins []int, min time.Duration, sleep time.Duration) {
	for _, v := range pins {
		pin := rpio.Pin(v)
		pin.Output()
		pin.Low()
		time.Sleep(time.Second*min)
		pin.High()
		time.Sleep(time.Second*sleep)
	}
}

func toggle(pinNo int, sec time.Duration) {
	pin := rpio.Pin(pinNo)
	pin.Output()
	pin.Low()
	time.Sleep(time.Second*sec)
	pin.High()
}
