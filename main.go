package main

import (
	"fmt"
	"sync"

	"gopkg.in/mgo.v2"
)

var db *mgo.Session

type poll struct {
	Options []string
}

type vote struct {
	Vote  string
	Tweet tweet
}

var wg sync.WaitGroup

func main() {
	votesChannel := make(chan vote)
	quit := make(chan struct{})

	var seattleVotes, portlandVotes int
	wg.Add(1)
	go readFromTwitter(votesChannel, []string{"Portland", "Seattle"}, quit, &wg)
	for x := range votesChannel {
		fmt.Println(x)
		switch x.Vote {
		case "Seattle":
			seattleVotes++
		case "Portland":
			portlandVotes++
		}
		fmt.Println("Seattle: ", seattleVotes, "Portland: ", portlandVotes)
		if seattleVotes >= 10000 || portlandVotes >= 10000 {
			fmt.Println("Done")
			close(quit)
			break
		}
	}

	wg.Wait()
	fmt.Println("exit")
}

func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("pools").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

func dialdb() error {
	var err error
	fmt.Println("dialing mongod: localhost")
	db, err = mgo.Dial("localhost")
	return err
}

func closedb() {
	db.Close()
}
