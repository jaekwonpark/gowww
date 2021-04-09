package main

import (
  "fmt"
  "github.com/gorilla/mux"
  "github.com/gorilla/sessions"
  "github.com/stianeikeland/go-rpio"
  "io"
  "log"
  "net/http"
  "os"
  "strconv"
  "sync"
  "time"
  "strings"
)

var (
  waterStations = []int{22, 23, 24}
  Garage        = 25
  store         = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SESSION")))
  webkey        = os.Getenv("WEBKEY")
  cert          = os.Getenv("CERT_PATH") + "bundle.crt"
  certkey       = os.Getenv("CERT_PATH") + "private.key"
  cookieName    = "ctrl"
  sessionName   = "certified"
  gIndexHtmlFile = &ThreadSafeFile {}
)

type ThreadSafeFile struct {
  mu sync.Mutex
  handle *os.File
}

func (f *ThreadSafeFile) write(out http.ResponseWriter) {
  // lock to prevent the case where the file pointer is rewound while another thread is reading
  f.mu.Lock()
  // rewind
  var err error
  if _, err = f.handle.Seek(0, 0); err != nil {
    log.Fatal(err)
  }
  if _, err = io.Copy(out, f.handle); err != nil {
    log.Fatal(err)
  }
  f.mu.Unlock()

}

func (f *ThreadSafeFile) close() {
  if err := f.handle.Close(); err != nil {
    log.Fatal(err)
  }
}

func init() {
  indexFileName := "./static/index.html"
  var err error

  gIndexHtmlFile.handle, err = os.Open(indexFileName)
  if err != nil { log.Fatal(err) }
}


func sprinkler(w http.ResponseWriter, r *http.Request) {
  if !auth(r) {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
  }
  num, ok := r.URL.Query()["no"]
  if !ok || len(num[0]) < 1 {
    go ToggleSprinkler(waterStations, 5, 1)
  } else {
    i, err := strconv.Atoi(num[0])
    //i = 0(all), 1 (station 1), 2 (station 2)
    if err != nil || i < 0 || i > len(waterStations) {
      log.Printf("given station number is out of range")
    } else {
      if i == 0 {
        go ToggleSprinkler(waterStations, 5, 1)
      } else {
        // 1 -> array of waterStations[0]
        go ToggleSprinkler([]int{waterStations[i-1]}, 5, 1)
      }
    }
  }
  http.Redirect(w, r, "/ctrl", http.StatusFound)
}

func garage(w http.ResponseWriter, r *http.Request) {
  if !auth(r) {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
  }
  go ToggleGarageDoor(Garage, 1)
  http.Redirect(w, r, "/ctrl", http.StatusFound)
}

func auth(r *http.Request) bool {
  session, _ := store.Get(r, cookieName)
  if auth, ok := session.Values[sessionName].(string); !ok || (auth != webkey) {
    log.Printf("auth=%s|webkey=%s\n", auth, webkey)
    return false
  }
  return true
}

func ctrl(w http.ResponseWriter, r *http.Request) {
  if !auth(r) {
	log.Printf("r.RemoteAddr:%s", r.RemoteAddr)
	// strings.Split guarantees it returns at least one element array
	// so no need to check null
	remote_ip := strings.Split(r.RemoteAddr, ":")
    if (remote_ip[0] == "192.168.1.1") {
      login(w, r)
	} else {
      http.Error(w, "Forbidden", http.StatusForbidden)
      return
   }
  }
  gIndexHtmlFile.write(w)
  // flush output buffer
}

func login(w http.ResponseWriter, r *http.Request) {
  // for 32 bit, max expiry date is 1-19-2038
  // so, don't set expiry date beyond that day
  store.MaxAge(86400 * 365 * 10)
  session, _ := store.Get(r, cookieName)
  session.Values[sessionName] = webkey
  if err := session.Save(r, w); err != nil {
    log.Fatal(err)
  }
}

func logout(w http.ResponseWriter, r *http.Request) {
  session, _ := store.Get(r, cookieName)
  session.Values[sessionName] = ""
  if err := session.Save(r, w); err != nil {
    log.Fatal(err)
  }
}

func main() {

  // declare mux router
  r := mux.NewRouter()

  r.HandleFunc("/ctrl", ctrl)
  //r.HandleFunc("/login", login)
  r.HandleFunc("/logout", logout)
  r.HandleFunc("/sprinkler", sprinkler)
  r.HandleFunc("/garage", garage)
  //r.PathPrefix("/static/").Handler(http.FileServer(http.Dir("./")))
  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

  // Open and map memory to access gpio, check for errors
  if err := rpio.Open(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  // Unmap gpio memory when done
  defer func() {
    if err := rpio.Close(); err != nil {
      log.Fatal(err)
    }
  }()

  err := http.ListenAndServeTLS(":443", cert, certkey, r)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

func ToggleSprinkler(pins []int, min time.Duration, sleep time.Duration) {
  for _, v := range pins {
    pin := rpio.Pin(v)
    state := pin.Read()
    pin.Output()
    if state == rpio.High { // if the sprinkler is off
      pin.Low()
      time.Sleep(time.Minute * min)
    }
    // if the sprinkler is on, turn off
    pin.High()
    time.Sleep(time.Second * sleep)
  }
}

func ToggleGarageDoor(pinNo int, sec time.Duration) {
  pin := rpio.Pin(pinNo)
  pin.Output()
  pin.Low()
  time.Sleep(time.Second * sec)
  pin.High()
}
