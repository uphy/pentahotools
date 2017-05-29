package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// User represent a user on Pentaho.
type User struct {
	name     string
	password string
}

// ReadUsersFile reads users from a csv file.
func ReadUsersFile(file string) ([]User, error) {
	// Create user from CSV file.
	file1, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer file1.Close()
	bar := pb.StartNew(0)
	bar.Prefix("Scanning CSV file...")
	users := []User{}
	reader := csv.NewReader(transform.NewReader(file1, japanese.ShiftJIS.NewDecoder()))
	reader.LazyQuotes = true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		switch len(record) {
		case 0:
			continue
		case 2:
			users = append(users, User{record[0], record[1]})
			if err != nil {
				return nil, fmt.Errorf("Failed to create user. (user=%s, err=%s)", record[0], err)
			}
		default:
			return nil, fmt.Errorf("Unsupported format line. (line=%s)", record)
		}
	}
	return users, nil
}
