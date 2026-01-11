package persistence

import (
	"io"
)

// DBSaver defines the interface for saving database to disk
// Using interface{} to avoid circular import
type DBSaver interface {
	SaveDB(db interface{}, filename string) error
	SaveDBToWriter(db interface{}, writer io.Writer) error
}

// SaverRegistry holds the registered saver
var saver DBSaver

// RegisterSaver registers a database saver implementation
func RegisterSaver(s DBSaver) {
	saver = s
}

// GetSaver returns the registered saver
func GetSaver() DBSaver {
	return saver
}

// SaveDatabase saves the database using the registered saver
func SaveDatabase(db interface{}, filename string) error {
	if saver == nil {
		return nil // No saver registered, skip silently
	}
	return saver.SaveDB(db, filename)
}

// SaveDatabaseToWriter saves the database to an io.Writer
func SaveDatabaseToWriter(db interface{}, writer io.Writer) error {
	if saver == nil {
		return nil // No saver registered, skip silently
	}
	return saver.SaveDBToWriter(db, writer)
}
