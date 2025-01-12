package data

import (
	"fmt"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"github.com/zooper-corp/CoinWatch/tools"
	"log"
	"strings"
	"time"
)

const (
	balanceCollection = "balance"
)

type Db struct {
	settings sqlite.ConnectionURL
}

type BalanceQueryOptions struct {
	Days int
}

func FromFile(path string) (Db, error) {
	var settings = sqlite.ConnectionURL{
		Database: tools.ExpandPath(path),
	}
	exists, err := tools.PathExists(path)
	if exists {
		sess, err := sqlite.Open(settings)
		if err != nil {
			log.Fatalf("Unable to open DB '%v' it might be corrupted: %v", settings, err)
			return Db{}, err
		}
		_ = sess.Close()
	}
	return Db{
		settings: settings,
	}, err
}

func GetTestDb() Db {
	d, _ := FromFile(GetTestDbPath())
	return d
}

func GetTestDbPath() string {
	return "/tmp/coinwatch.db"
}

func (d *Db) GetSession() (db.Session, error) {
	sess, err := sqlite.Open(d.settings)
	if err != nil {
		log.Fatalf("Unable to open DB '%v' it might be corrupted: %v", d.settings, err)
		return nil, err
	}
	return sess, nil
}

func (d *Db) InsertBalance(balance Balance) error {
	sess, err := d.GetSession()
	if err != nil {
		return err
	}
	defer func(sess db.Session) {
		_ = sess.Close()
	}(sess)
	collection := sess.Collection(balanceCollection)
	// Table not there just create one
	exists, _ := collection.Exists()
	if !exists {
		log.Printf("Create main balance table")
		_, err = sess.SQL().Exec(fmt.Sprintf(`
        CREATE TABLE %v (
            ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			wallet TEXT,
			token TEXT,
			address TEXT,
			balance REAL,
			balance_locked REAL,
			fiat_value REAL
        )`, balanceCollection))
		if err != nil {
			log.Fatalf("Unable to create main balance table")
			return err
		}
	}
	// Insert
	_, err = collection.Insert(balance)
	return err
}

func (d *Db) GetBalances(options BalanceQueryOptions) (Balances, error) {
	sess, err := d.GetSession()
	if err != nil {
		return Balances{}, err
	}
	defer func(sess db.Session) {
		err := sess.Close()
		if err != nil {

		}
	}(sess)
	collection := sess.Collection(balanceCollection)
	// Table not there just return an empty set
	exists, _ := collection.Exists()
	if !exists {
		return Balances{}, nil
	}
	// Do search
	var result []Balance
	q := sess.SQL().
		SelectFrom(balanceCollection).
		OrderBy("ts desc")
	// Day limit
	if options.Days > 0 {
		q = q.Where(fmt.Sprintf("ts BETWEEN datetime('now', '-%d days') "+
			"AND datetime('now', 'localtime')", options.Days))
	}
	// Query
	log.Printf(q.String())
	if err := q.All(&result); err != nil {
		return Balances{}, err
	}
	// Log the results
	log.Printf("Query returned %d results", len(result))
	return Balances{
		entries: result,
	}, nil
}

func (d *Db) GetBalancesInRange(from, to time.Time, interval time.Duration, mode string) ([]map[string]interface{}, error) {
	sess, err := d.GetSession()
	if err != nil {
		return nil, err
	}
	defer func(sess db.Session) {
		_ = sess.Close()
	}(sess)
	collection := sess.Collection(balanceCollection)
	// Check if the table exists
	exists, _ := collection.Exists()
	if !exists {
		return nil, nil
	}
	// Query for balances within the time range
	fromStr := from.Format(time.RFC3339) // ISO 8601 format
	toStr := to.Format(time.RFC3339)     // ISO 8601 format
	var result []Balance
	q := sess.SQL().
		SelectFrom(balanceCollection).
		Where(fmt.Sprintf("ts BETWEEN '%s' AND '%s'", fromStr, toStr)).
		OrderBy("ts ASC")
	log.Printf(q.String())
	if err := q.All(&result); err != nil {
		return nil, err
	}
	// Log the results
	log.Printf("Query returned %d results", len(result))
	// Organize data by interval
	var balances []map[string]interface{}
	currentTime := from
	// Keep track of the previous values for each token
	previousValues := make(map[string]float64)
	for currentTime.Before(to) {
		point := map[string]interface{}{
			"timestamp": currentTime.Format(time.RFC3339),
		}
		hasChanges := false
		// Loop through the balances and filter by token
		for _, entry := range result {
			if entry.Timestamp.After(currentTime) && entry.Timestamp.Before(currentTime.Add(interval)) {
				token := strings.ToLower(entry.Token)
				var value float64
				if mode == "token" {
					value = entry.Balance // Use Balance when mode is "token"
				} else if mode == "fiat_value" {
					value = entry.FiatValue // Use FiatValue when mode is "fiat_value"
				}
				// Check if the value has changed
				if previousValue, exists := previousValues[token]; !exists || value != previousValue {
					point[token] = value
					previousValues[token] = value
					hasChanges = true
				}
			}
		}
		// Only append the point if there were changes
		if hasChanges {
			balances = append(balances, point)
		}
		currentTime = currentTime.Add(interval)
	}
	return balances, nil
}
