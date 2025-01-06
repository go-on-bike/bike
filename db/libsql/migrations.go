package libsql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/matisin/bike/assert"
)

type Migration struct {
	ID   int
	Name string
	SQL  string
}

func (lsql *LibSqlOperator) initMigrations() {
	_, err := lsql.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	assert.ErrNil(err, "Init migrations failed")
}

func (lsql *LibSqlOperator) findLastMigrationID() int {
	var lastID int
	err := lsql.db.QueryRow(`
    SELECT id 
    FROM migrations 
    ORDER BY id DESC 
    LIMIT 1
`).Scan(&lastID)

	if err == sql.ErrNoRows {
		return 0
	}
	assert.ErrNil(err, "Find last migration failed")
	return lastID
}

func (lsql *LibSqlOperator) loadMigrations(path string, direction string, steps int) []Migration {
	lastID := lsql.findLastMigrationID()

	filenames, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s.sql", direction)))
	assert.ErrNil(err, "Getting files from migration path failed")

	dirSwitch := map[string]string{
		"up":   "down",
		"down": "up",
	}

	maxIDs := len(filenames)

	if steps == 0 {
		if direction == "up" {
			steps = maxIDs - lastID
		} else {
			steps = lastID
		}
	}

	if steps == 0 {
		return []Migration{}
	}
	assert.IntGreater(steps, 0, "steps is 0")

	var fromID, toID int
	if direction == "up" {
		fromID = lastID + 1
		toID = lastID + steps
		if toID > maxIDs {
			toID = maxIDs
		}
	} else {
		toID = lastID
		fromID = lastID - steps
		if fromID < 1 {
			fromID = 1
		}
	}
    if fromID > toID {
        return []Migration{}
    }
	assert.IntGeq(toID, fromID, "toId is not Greater or equal than fromId")
	assert.IntGreater(toID, 0, "toID is 0")

	migrations := make([]Migration, toID-fromID+1)
	for _, filename := range filenames {
		_, name := filepath.Split(filename)
		noSuffix := strings.TrimSuffix(name, fmt.Sprintf(".%s.sql", direction))
		nameParts := strings.Split(noSuffix, "_")
		assert.StringArrayMin(nameParts, 2)

		id, err := strconv.Atoi(nameParts[0])
		assert.ErrNil(err, "Failed getting id from migration file")
		assert.IntNot(id, 0)

		if id < fromID || id > toID {
			continue
		}

		// Leer contenido del archivo SQL
		content, err := os.ReadFile(filename)
		assert.ErrNil(err, "Failed to read file")
		assert.Bytes(content)

		// Verificar que exista direccion contraria
		c, err := os.ReadFile(filepath.Join(path, fmt.Sprintf("%s.%s.sql", noSuffix, dirSwitch[direction])))
		assert.ErrNil(err, "Failed to read oposite file")
		assert.Bytes(c)

		index := id - fromID
		migrations[index] = Migration{
			ID:   id,
			Name: strings.Join(nameParts[1:], "_"),
			SQL:  string(content),
		}
	}
	// Ordenar migraciones por ID
	sort.Slice(migrations, func(i, j int) bool {
		if direction == "up" {
			return migrations[i].ID < migrations[j].ID
		}
		return migrations[i].ID > migrations[j].ID
	})

	assert.SliceIsSorted(migrations, func(i, j int) bool {
		if direction == "up" {
			return migrations[i].ID < migrations[j].ID
		}
		return migrations[i].ID > migrations[j].ID
	})

	return migrations
}

func (lsql *LibSqlOperator) runMigrationsUp(migrations []Migration) {
	// lsql.logger.Info(fmt.Sprintf("Running %d up migrations", len(migrations)))
	for _, m := range migrations {
		// lsql.logger.Info("Running migration", "ID", m.ID, "name", m.Name)

		tx, err := lsql.db.Begin()
		assert.ErrNil(err, "Failed fo start transaction")

		// Ejecutar statements
		stmts := strings.Split(m.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			// lsql.logger.Info("statement", "s", s)
			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				tx.Rollback()
				assert.ErrNil(err, fmt.Sprintf("migration %d failed", m.ID))
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, m.ID, m.Name); err != nil {
			tx.Rollback()
			assert.ErrNil(err, fmt.Sprintf("failed to register migration %d", m.ID))
		}

		if err := tx.Commit(); err != nil {
			assert.ErrNil(err, fmt.Sprintf("failed to commit migration %d", m.ID))
		}
	}
}

func (lsql *LibSqlOperator) runMigrationsDown(migrations []Migration) {
	// lsql.logger.Info(fmt.Sprintf("Running %d down migrations", len(migrations)))
	for _, m := range migrations {
		// lsql.logger.Info("Running migration", "ID", m.ID, "name", m.Name)

		tx, err := lsql.db.Begin()
		assert.ErrNil(err, "Failed fo start transaction")

		// Ejecutar statements
		stmts := strings.Split(m.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			// lsql.logger.Info("statement", "s", s)
			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				tx.Rollback()
				assert.ErrNil(err, fmt.Sprintf("migration %d failed", m.ID))
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`DELETE from MIGRATIONS WHERE id = ?`, m.ID); err != nil {
			tx.Rollback()
			assert.ErrNil(err, fmt.Sprintf("failed to register migration %d", m.ID))
		}

		if err := tx.Commit(); err != nil {
			assert.ErrNil(err, fmt.Sprintf("failed to commit migration %d", m.ID))
		}
	}
}

func (lsql *LibSqlOperator) RunMigrations(path string, direction string, steps int) error {
	lsql.initMigrations()

	assert.StringAllowedValues(direction, "up", "down")

	migrations := lsql.loadMigrations(path, direction, steps)

	if len(migrations) == 0 {
        return fmt.Errorf("There's no migrations to run")
	}

	assert.IntGreater(len(migrations), 0, fmt.Sprintf("Migrations array is %d", len(migrations)))

	if direction == "up" {
		lsql.runMigrationsUp(migrations)
		return nil
	}

	lsql.runMigrationsDown(migrations)
    return nil
}
