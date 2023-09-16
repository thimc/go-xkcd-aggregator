package xkcdstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	recentURL string = "https://xkcd.com/info.0.json"
	entryURL  string = "https://xkcd.com/%d/info.0.json"
)

type XkcdComic struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	Image      string `json:"img"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	// Month      string `json:"month"`
	// Day        string `json:"day"`
	// Year       string `json:"year"`
	// Link       string `json:"link"`
	// News       string `json:"news"`
	// SafeTitle  string `json:"json_title"`
}

type XkcdStore struct {
	db      *sql.DB
	mut     sync.RWMutex
	Entries []XkcdComic
}

func New(fileName string) (*XkcdStore, error) {
	entries := []XkcdComic{}

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}

	store := &XkcdStore{
		db:      db,
		Entries: entries,
	}
	if err := store.init(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *XkcdStore) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS xkcd (
    	num INTEGER UNIQUE,
		title TEXT,
		img TEXT,
		alt TEXT,
		transcript TEXT
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		log.Printf("Database init: %s\n", err)
	}

	return nil
}

func (s *XkcdStore) Close() error {
	return s.db.Close()
}

func (s *XkcdStore) Fetch(ID int) (*XkcdComic, error) {
	var (
		entry  = &XkcdComic{}
		client = http.Client{}
		url    string
	)

	if ID <= 0 {
		url = recentURL
	} else {
		url = fmt.Sprintf(entryURL, ID)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *XkcdStore) Insert(entry *XkcdComic) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	query := `
	INSERT INTO xkcd(num, title, img, alt, transcript)
	VALUES (?, ?, ?, ?, ?);`

	_, err := s.db.Exec(query, entry.Num, entry.Title, entry.Image, entry.Alt, entry.Transcript)
	if err != nil {
		return err
	}

	return nil
}

func (s *XkcdStore) Current() (int, error) {
	var out int = 1
	query := `SELECT COUNT(num) FROM xkcd LIMIT 1;`
	res, err := s.db.Query(query)
	if err != nil {
		return out, err
	}

	defer res.Close()
	for res.Next() {
		if err := res.Scan(&out); err != nil {
			return out, err
		}
		break
	}

	return out, nil
}

func (s *XkcdStore) Search(term string) (*[]XkcdComic, error) {
	var entries []XkcdComic
	query := `
	SELECT *
	FROM xkcd
	WHERE title LIKE "%?%"
	OR img LIKE "%?%"
	OR alt LIKE "%?%"
	or transcript LIKE "%?%"
	ORDER BY num ASC;`
	query = strings.ReplaceAll(query, "?", term)

	if term == "-" {
		query = `SELECT * FROM xkcd ORDER BY num ASC;`
	}

	res, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer res.Close()
	for res.Next() {
		var entry XkcdComic
		err := res.Scan(&entry.Num,
			&entry.Title,
			&entry.Image,
			&entry.Alt,
			&entry.Transcript)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return &entries, nil
}
