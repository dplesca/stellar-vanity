package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/stellar/go/keypair"
)

const maxConcurrency = 10

const txtTemplate = `
Address: {{.Address}}
Seed:    {{.Seed}}
`

var positions = []string{
	"anywhere",
	"end",
	"start",
}
var position string
var wg sync.WaitGroup
var throttle = make(chan int, maxConcurrency)
var verbose, writeToFile bool
var start time.Time

func main() {

	var RootCmd = &cobra.Command{
		Use:   "stellar-vanity-address [flags] text",
		Short: "Stellar vanity address generator",
		Long:  `Generate a stellar vanity address.`,
		RunE:  search,
	}

	RootCmd.PersistentFlags().StringVarP(&position, "position", "p", "anywhere", "position of searched text in the address: anywhere, start, end")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output that adds a bit of extra info, like the number or pairs searched")
	RootCmd.PersistentFlags().BoolVarP(&writeToFile, "writetofile", "f", false, "flag for writing resulted address/seed pair into a file")
	RootCmd.Execute()

}

func search(cmd *cobra.Command, args []string) error {
	// regex that checks that seaerchstring does not contain any invalid chacters
	isValid := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
	if len(args) == 0 || !isValid(args[0]) {
		return errors.New("Please enter a valid string to search for")
	}

	// check that the position has a predefined value, it can and should be improved
	if !strings.Contains(strings.Join(positions, ","), position) {
		return errors.New("Position to search at can be: anywhere, start, end")
	}

	var index int64
	index = 1
	start = time.Now()
	fmt.Println("Start search for", args[0], "at position:", position)
	u := strings.ToUpper(args[0])
	for {

		if verbose == true && index%100000 == 0 {
			fmt.Printf("\rSearched %s pairs", humanize.Comma(index))
		}
		throttle <- 1
		wg.Add(1)
		go generatePair(u, index)
		index++
	}
	return nil
}

func generatePair(text string, index int64) {

	var r bool
	pair, err := keypair.Random()
	if err != nil {
		log.Fatal(err)
	}

	switch position {
	case "start":
		r = checkStart(pair.Address(), text)
	case "end":
		r = checkEnd(pair.Address(), text)
	default:
		r = checkMiddle(pair.Address(), text)
	}
	if r == true {
		writeFinalMessage(pair, index, text)
	}

	wg.Done()
	<-throttle
}

func checkStart(s, substr string) bool {
	if strings.HasPrefix(substr, "GA") || strings.HasPrefix(substr, "GB") || strings.HasPrefix(substr, "GC") || strings.HasPrefix(substr, "GD") {
		return strings.HasPrefix(s, substr)
	}
	s = s[1:]
	if strings.HasPrefix(substr, "A") || strings.HasPrefix(substr, "B") || strings.HasPrefix(substr, "C") || strings.HasPrefix(substr, "D") {
		return strings.HasPrefix(s, substr)
	}
	return strings.HasPrefix(s[1:], substr)
}

func checkMiddle(s, substr string) bool {
	return strings.Contains(s, substr)
}

func checkEnd(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func writeFinalMessage(pair *keypair.Full, index int64, u string) {
	end := time.Now().Sub(start)
	if verbose == true {
		d, _ := durafmt.ParseString(end.String())
		fmt.Printf("\rChecked a total of %s pairs in: %s\n", humanize.Comma(index), d)
	}
	c := strings.Split(pair.Address(), u)
	fmt.Printf("\nSearch successful! Results:\n\nAddress: %s%s%s\nSeed:\t %s\n", c[0], aurora.Green(u), c[1], pair.Seed())

	if writeToFile == true {
		f, err := os.OpenFile("result.txt", os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		t := template.Must(template.New("t2").Parse(txtTemplate))

		if err := t.Execute(f, struct{ Address, Seed string }{pair.Address(), pair.Seed()}); err != nil {
			log.Fatal(err)
		}
	}

	//this could probably be handled more gracefully, ¯\_(ツ)_/¯
	os.Exit(0)
}
