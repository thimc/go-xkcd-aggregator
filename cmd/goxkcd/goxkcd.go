package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/thimc/go-xkcd-aggregator/pkg/xkcdstore"
)

func usage() {
	fmt.Printf("usage: %s <command> <args>\n", os.Args[0])
	fmt.Printf("  help           - prints the usage for this program\n")
	fmt.Printf("  download       - downloads and indexes all missing xkcd entries\n")
	fmt.Printf("  search <term>  - looks up any xkcd entries matching \"term\"\n")
	os.Exit(1)
}

func main() {
	store, err := xkcdstore.New("database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	args := os.Args[1:]
	if len(args) < 1 {
		usage()
	}

	switch args[0] {
	case "help":
		usage()
	case "download":
		resultch := make(chan xkcdstore.XkcdComic, 5)
		wg := &sync.WaitGroup{}

		latest, err := store.Fetch(-1)
		log.Printf("There are %d entries in total\n", latest.Num)

		current, err := store.Current()
		if err != nil {
			log.Println("Assuming there are no previous entries")
		} else {
			log.Printf("%d are already stored", current)
		}

		wg.Add(latest.Num-current)

		for i := current; i < latest.Num; i++ {
			go func(ID int) {
				entry, err := store.Fetch(ID)
				if err != nil {
					log.Printf("Error when fetching %d: %s", ID, err)
					wg.Done()
					return
				}
				if err := store.Insert(entry); err != nil {
					log.Printf("Error when inserting %d: %s", entry.Num, err)
					wg.Done()
					return
				}
				log.Printf("Inserted #%d %s\n", entry.Num, entry.Title)
				wg.Done()
			}(i)
		}

		wg.Wait()
		close(resultch)

	case "search":
		entries, err := store.Search(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		for _, entry := range *entries {
			fmt.Printf("%d\t%s\t%s\t%s\n", entry.Num, entry.Title, entry.Alt, entry.Image)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: \"%s\".", args[0])
		os.Exit(1)
	}

}
