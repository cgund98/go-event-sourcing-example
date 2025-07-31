package pg

import "fmt"

// dslError generates an error caused by synthesizing a SQL query
func ErrorDsl(err error) error {
	return fmt.Errorf("error generating sql statement: %v", err)
}

func ErrorDb(err error) error {
	return fmt.Errorf("error running SQL query: %v", err)
}

func ErrorUnmarshal(err error) error {
	return fmt.Errorf("error marshaling SQL query: %v", err)
}
