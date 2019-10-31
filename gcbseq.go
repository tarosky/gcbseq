package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/urfave/cli"
)

const description = "This generates an incremented sequential number every time this program runs. " +
	"The number can be used for build number on Google Cloud Build. " +
	"This program uses Google Cloud Datastore to store the counter."

type flags struct {
	datastoreNamespace string
	projectID          string
	initialNumber      int64
	out                string
}

func flagsFromCLI(c *cli.Context) *flags {
	return &flags{
		datastoreNamespace: c.String("datastore-namespace"),
		projectID:          c.String("project-id"),
		initialNumber:      c.Int64("initial-number"),
		out:                c.String("out"),
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "gcbseq"
	app.Description = description
	app.Usage = "This generates a number suitable for sequential build number."
	app.UsageText = "gcbseq --datastore-namespace cloudbuild_myproject --out artifact/BUILD_NUMBER"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "datastore-namespace,n",
			Usage: "Datastore namespace used for storing counter",
			Value: "cloudbuild_gcbseq",
		},
		cli.StringFlag{
			Name:  "project-id,p",
			Usage: "Google Cloud Project ID to be used",
			Value: os.Getenv("PROJECT_ID"),
		},
		cli.Int64Flag{
			Name:  "initial-number,i",
			Usage: "Initialize the build number as this value when the counter is empty",
			Value: 1,
		},
		cli.StringFlag{
			Name:  "out,o",
			Usage: "Output the new build number to `FILE`",
			Value: "BUILD_NUMBER",
		},
	}

	app.Action = func(c *cli.Context) error {
		return run(flagsFromCLI(c))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Entity holds an entity on Datastore.
type Entity struct {
	Counter int64 `datastore:",noindex"`
}

func writeToFile(counter int64, out string) error {
	return ioutil.WriteFile(out, []byte(strconv.FormatInt(counter, 10)), 0644)
}

func run(f *flags) error {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, f.projectID)
	if err != nil {
		return err
	}

	key := &datastore.Key{
		Kind:      "GCBSeq",
		Name:      "gcbsec",
		Parent:    nil,
		Namespace: f.datastoreNamespace,
	}

	entity := &Entity{}

	if _, err := client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		if err := tx.Get(key, entity); err != nil {
			if err != datastore.ErrNoSuchEntity {
				return err
			}
			entity.Counter = f.initialNumber
		} else {
			entity.Counter++
		}

		tx.Put(key, entity)

		return nil
	}); err != nil {
		return err
	}

	if err := writeToFile(entity.Counter, f.out); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Build number: %d\n", entity.Counter)

	return nil
}
