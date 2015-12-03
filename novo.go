package apinovo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/bgentry/que-go"
	"github.com/jackc/pgx"
	"github.com/pmylund/go-cache"
)

type Config struct {
	WorkserSleep time.Duration
	MaxWOrkers   int
	CacheExpire  time.Duration
	CacheClean   time.Duration
	LinkExpire   time.Duration
	DBConn       string
}

type DocIndex struct {
	URL  string
	Text []string
}

type Nova struct {
	queue     *que.Client
	fetch     *fetchbot.Fetcher
	mu        sync.RWMutex
	cfg       *Config
	stop      chan struct{}
	Q         *fetchbot.Queue
	workerMap que.WorkMap
	cache     *cache.Cache
	pgPool    *pgx.ConnPool
	pgConfig  pgx.ConnConfig
}

func NewNova(cfg *Config) (*Nova, error) {
	pgxcfg, err := pgx.ParseURI(cfg.DBConn)
	if err != nil {
		return nil, err
	}
	pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:   pgxcfg,
		AfterConnect: que.PrepareStatements,
	})
	if err != nil {
		return nil, err
	}
	n := &Nova{
		queue:    que.NewClient(pgxpool),
		cfg:      cfg,
		stop:     make(chan struct{}),
		cache:    cache.New(cfg.CacheExpire, cfg.CacheClean),
		pgPool:   pgxpool,
		pgConfig: pgxcfg,
	}
	n.Init()
	return n, nil
}

func (n *Nova) Init() {
	n.setWorker("indexer", n.Indexer)
	go n.startWorkers()
	mux := fetchbot.NewMux()
	mux.Response().Method("GET").
		ContentType("text/html").
		Handler(fetchbot.HandlerFunc(n.ProcessLink))
	n.fetch = fetchbot.New(mux)
	n.Q = n.fetch.Start()
}

func (n *Nova) setWorker(name string, worker que.WorkFunc) {
	n.mu.RLock()
	n.workerMap[name] = worker
	n.mu.Unlock()
}

func (n *Nova) startWorkers() {
	workers := que.NewWorkerPool(n.queue, n.workerMap, n.cfg.MaxWOrkers)
	go workers.Start()
END:
	for {
		select {
		case <-n.stop:
			workers.Shutdown()
			break END
		}
	}
}

func (n *Nova) ProcessLink(ctx *fetchbot.Context, res *http.Response, err error) {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		// log this?
	}
	n.mu.RLock()
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")
		// Resolve address
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			fmt.Printf("error: resolve URL %s - %s\n", val, err)
			return
		}
		if !n.LinkDuplicate(u.String()) {
			if _, err := ctx.Q.SendStringHead(u.String()); err != nil {
				fmt.Printf("error: enqueue head %s - %s\n", u, err)
			} else {
				n.LinkDone(u.String())
				data, err := json.Marshal(DocIndex{URL: u.String(), Text: TextFromHTML(res.Body)})
				if err != nil {
					log.Fatal(err)
				}

				job := &que.Job{
					Type: "index",
					Args: data,
				}
				err = n.queue.Enqueue(job)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	})
	n.mu.Unlock()
}

func (n *Nova) LinkDuplicate(link string) bool {
	_, ok := n.cache.Get(link)
	return ok
}

func (n *Nova) LinkDone(link string) {
	n.cache.Set(link, true, n.cfg.LinkExpire)
}

func (n Nova) Indexer(j *que.Job) error {
	return nil
}

func (n *Nova) ShutDown() {
	n.stop <- struct{}{}
	n.Q.Close()
}
