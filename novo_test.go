package apinovo

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNova(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile("fixtures/home.html")
		if err != nil {
			t.Error(err)
		}
		w.Write(b)
	})
	ts := httptest.NewServer(m)
	defer ts.Close()
	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		t.Skip("No database connection")
	}
	cfg := &Config{
		DBConn:       dbConn,
		WorkserSleep: time.Millisecond,
	}
	n, err := NewNova(cfg)
	if err != nil {
		t.Fatal(err)
	}
	n.Q.SendString(ts.URL)
	time.Sleep(time.Second)
	n.ShutDown()
}
