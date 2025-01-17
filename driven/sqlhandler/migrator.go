package sqlhandler

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type migrOpts struct {
	path *string
}

type Migrator struct {
	stderr  io.Writer
	db      *sql.DB
	options migrOpts
}

const SigMigr string = "sqlhandler migrator"

func NewMigrator(stderr io.Writer, db *sql.DB, opts ...MigrOption) *Migrator {
	if stderr == nil {
		panic(fmt.Sprintf("%s: stderr cannot be nil", SigMigr))
	}

	m := &Migrator{stderr: stderr, db: db}
	for _, opt := range opts {
		opt(&m.options)
	}

	return m
}

func (m *Migrator) SetDB(db *sql.DB) {
	fmt.Fprintf(m.stderr, "%s: setting new db", SigMigr)
	if isConnected(m.db) {
		panic(fmt.Sprintf("%s: cannot change connected connection", SigMigr))
	}

	if !isConnected(db) {
		panic(fmt.Sprintf("%s: cannot change connection to a closed one", SigMigr))
	}

	m.db = db
}

func (m *Migrator) init() error {
	fmt.Fprintf(m.stderr, "%s: executing a query in init", SigMigr)
	_, err := m.db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("%s: failed to initialize migrations table: %w", SigMigr, err)
	}
	return nil
}

func (m *Migrator) findLastID() (int, error) {
	var lastID int
	fmt.Fprintf(m.stderr, "%s: executing query row in find last id", SigMigr)
	err := m.db.QueryRow(`
        SELECT id 
        FROM migrations 
        ORDER BY id DESC 
        LIMIT 1
    `).Scan(&lastID)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%s: failed to find last migration ID: %w", SigMigr, err)
	}
	return lastID, nil
}

type Migration struct {
	ID   int
	Name string
	SQL  string
}

func (m *Migrator) load(path string, inverse bool, steps int) ([]Migration, error) {
	direction := map[bool]string{true: "down", false: "up"}

	if steps < 0 {
		return nil, fmt.Errorf("%s: steps cannot be negative, got %d", SigMigr, steps)
	}

	lastID, err := m.findLastID()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find last migration ID: %w", SigMigr, err)
	}

	filenames, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s.sql", direction[inverse])))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get migration files: %w", SigMigr, err)
	}

	maxIDs := len(filenames)

	// Calcular steps si es 0
	if steps == 0 {
		if !inverse {
			steps = maxIDs - lastID
		} else {
			steps = lastID
		}
	}

	if steps == 0 {
		return []Migration{}, nil
	}

	if steps < 0 {
		return nil, fmt.Errorf("%s: step calculation failed, got %d", SigMigr, steps)
	}

	// Calcular rangos de migración
	var fromID, toID int
	if !inverse {
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
		return []Migration{}, nil
	}

	migrations := make([]Migration, toID-fromID+1)
	for _, filename := range filenames {
		_, name := filepath.Split(filename)
		noSuffix := strings.TrimSuffix(name, fmt.Sprintf(".%s.sql", direction[inverse]))
		nameParts := strings.Split(noSuffix, "_")

		if len(nameParts) < 2 {
			return nil, fmt.Errorf("%s: invalid migration filename format: %s", SigMigr, filename)
		}

		id, err := strconv.Atoi(nameParts[0])
		if err != nil {
			return nil, fmt.Errorf("%s: invalid migration ID in filename %s: %w", SigMigr, filename, err)
		}
		if id == 0 {
			return nil, fmt.Errorf("%s: migration ID cannot be 0 in file: %s", SigMigr, filename)
		}

		if id < fromID || id > toID {
			continue
		}

		// Leer contenido del archivo SQL
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read migration file %s: %w", SigMigr, filename, err)
		}
		if len(content) == 0 {
			return nil, fmt.Errorf("%s: migration file is empty: %s", SigMigr, filename)
		}

		// Verificar que existe el archivo opuesto
		counterpartPath := filepath.Join(path, fmt.Sprintf("%s.%s.sql", noSuffix, direction[!inverse]))
		counterpartContent, err := os.ReadFile(counterpartPath)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read counterpart file for %s: %w", SigMigr, filename, err)
		}
		if len(counterpartContent) == 0 {
			return nil, fmt.Errorf("%s: counterpart migration file is empty: %s", SigMigr, counterpartPath)
		}

		index := id - fromID
		migrations[index] = Migration{
			ID:   id,
			Name: strings.Join(nameParts[1:], "_"),
			SQL:  string(content),
		}
	}

	// Ordenar migraciones por ID
	sort.Slice(migrations, func(i, j int) bool {
		if !inverse {
			return migrations[i].ID < migrations[j].ID
		}
		return migrations[i].ID > migrations[j].ID
	})

	return migrations, nil
}

func isConnected(db *sql.DB) bool {
	return db != nil && db.Ping() == nil
}

func (m *Migrator) up(migrations []Migration) error {
	for _, mig := range migrations {
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("%s: failed to start transaction: %w", SigMigr, err)
		}

		// Ejecutar statements
		stmts := strings.Split(mig.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				rollErr := tx.Rollback()
				if rollErr != nil {
					// Aquí retornamos ambos errores ya que es crítico saber si falló tanto la migración como el rollback
					return fmt.Errorf("%s: migration %d failed: %v, additionally rollback failed: %v", SigMigr, mig.ID, err, rollErr)
				}
				return fmt.Errorf("%s: migration %d failed: %w", SigMigr, mig.ID, err)
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, mig.ID, mig.Name); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("%s: failed to register migration %d: %v,SigMigr, additionally rollback failed: %v", SigMigr, mig.ID, err, rollErr)
			}
			return fmt.Errorf("%s: failed to register migration %d: %w", SigMigr, mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("%s: failed to commit migration %d: %w", SigMigr, mig.ID, err)
		}
	}

	return nil
}

func (m *Migrator) down(migrations []Migration) error {
	for _, mig := range migrations {
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("%s: failed to start transaction: %w", SigMigr, err)
		}

		// Ejecutar statements
		stmts := strings.Split(mig.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				rollErr := tx.Rollback()
				if rollErr != nil {
					return fmt.Errorf("%s: migration %d rollback failed: %v,SigMigr, additionally transaction rollback failed: %v", SigMigr, mig.ID, err, rollErr)
				}
				return fmt.Errorf("%s: migration %d rollback failed: %w", SigMigr, mig.ID, err)
			}
		}

		// Eliminar registro de migración
		if _, err := tx.Exec(`DELETE from MIGRATIONS WHERE id = ?`, mig.ID); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("%s: failed to remove migration %d record: %v,SigMigr, additionally rollback failed: %v", SigMigr, mig.ID, err, rollErr)
			}
			return fmt.Errorf("%s: failed to remove migration %d record: %w", SigMigr, mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("%s: failed to commit migration %d rollback: %w", SigMigr, mig.ID, err)
		}
	}

	return nil
}

func (m *Migrator) Version() (int, error) {
	if !isConnected(m.db) {
		return 0, fmt.Errorf("%s: db in migrations is desconnected", SigMigr)
	}

	version, err := m.findLastID()
	return version, err
}

func (m *Migrator) Move(steps int, inverse bool) error {
	if !isConnected(m.db) {
		return fmt.Errorf("%s: db in migrations is desconnected", SigMigr)
	}
	// Inicializar tabla de migraciones si no existe
	if err := m.init(); err != nil {
		return fmt.Errorf("%s: failed to initialize migrations: %w", SigMigr, err)
	}

	// Cargar migraciones
	migrations, err := m.load(*m.options.path, inverse, steps)

	if err != nil {
		return err
	}

	// Verificar si hay migraciones para ejecutar
	if len(migrations) == 0 {
		return fmt.Errorf("%s: no migrations to run", SigMigr)
	}

	// Ejecutar migraciones según la dirección
	if !inverse {
		if err := m.up(migrations); err != nil {
			return fmt.Errorf("%s: failed to run up migrations: %w", SigMigr, err)
		}
		return nil
	}

	if err := m.down(migrations); err != nil {
		return fmt.Errorf("%s: failed to run down migrations: %w", SigMigr, err)
	}
	return nil
}
