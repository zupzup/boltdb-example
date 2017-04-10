package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

// Config type
type Config struct {
	Height   float64   `json:"height"`
	Birthday time.Time `json:"birthday"`
}

// Entry type
type Entry struct {
	Calories int    `json:"calories"`
	Food     string `json:"food"`
}

func main() {
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	conf := Config{Height: 186.0, Birthday: time.Now()}
	err = setConfig(db, conf)
	if err != nil {
		log.Fatal(err)
	}
	err = addWeight(db, "85.0", time.Now())
	if err != nil {
		log.Fatal(err)
	}
	err = addEntry(db, 100, "apple", time.Now())
	if err != nil {
		log.Fatal(err)
	}

	err = addEntry(db, 100, "orange", time.Now().AddDate(0, 0, -2))
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		conf := tx.Bucket([]byte("DB")).Get([]byte("CONFIG"))
		fmt.Printf("Config: %s\n", conf)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("WEIGHT"))
		b.ForEach(func(k, v []byte) error {
			fmt.Println(string(k), string(v))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("DB")).Bucket([]byte("ENTRIES")).Cursor()
		min := []byte(time.Now().AddDate(0, 0, -7).Format(time.RFC3339))
		max := []byte(time.Now().AddDate(0, 0, 0).Format(time.RFC3339))
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			fmt.Println(string(k), string(v))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("WEIGHT"))
		if err != nil {
			return fmt.Errorf("could not create weight bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("ENTRIES"))
		if err != nil {
			return fmt.Errorf("could not create days bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	fmt.Println("DB Setup Done")
	return db, nil
}

func setConfig(db *bolt.DB, config Config) error {
	confBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not marshal config json: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Put([]byte("CONFIG"), confBytes)
		if err != nil {
			return fmt.Errorf("could not set config: %v", err)
		}
		return nil
	})
	fmt.Println("Set Config")
	return err
}

func addWeight(db *bolt.DB, weight string, date time.Time) error {
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("WEIGHT")).Put([]byte(date.Format(time.RFC3339)), []byte(weight))
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})
	fmt.Println("Added Weight")
	return err
}

func addEntry(db *bolt.DB, calories int, food string, date time.Time) error {
	entry := Entry{Calories: calories, Food: food}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("could not marshal entry json: %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("ENTRIES")).Put([]byte(date.Format(time.RFC3339)), entryBytes)
		if err != nil {
			return fmt.Errorf("could not insert entry: %v", err)
		}

		return nil
	})
	fmt.Println("Added Entry")
	return err
}
