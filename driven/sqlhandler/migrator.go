package sqlhandler

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migrator struct {
	db      *sql.DB
	options migrOpts
}

func NewMigrator(db *sql.DB, opts ...MigrOption) *Migrator {
	m := &Migrator{db: db}
	for _, opt := range opts {
		opt(&m.options)
	}

	return m
}

func (m *Migrator) SetDB(db *sql.DB) {
    if isConnected(m.db) {
        panic("cannot change connected connection")
    }

    if !isConnected(db){
        panic("cannot change connection to a closed one")
    }

	m.db = db
}

func (m *Migrator) init() error {
	_, err := m.db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) findLastID() (int, error) {
	var lastID int
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
		return 0, fmt.Errorf("failed to find last migration ID: %w", err)
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
		return nil, fmt.Errorf("steps cannot be negative, got %d", steps)
	}

	lastID, err := m.findLastID()
	if err != nil {
		return nil, fmt.Errorf("failed to find last migration ID: %w", err)
	}

	filenames, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s.sql", direction[inverse])))
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
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
		return nil, fmt.Errorf("step calculation failed, got %d", steps)
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
			return nil, fmt.Errorf("invalid migration filename format: %s", filename)
		}

		id, err := strconv.Atoi(nameParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration ID in filename %s: %w", filename, err)
		}
		if id == 0 {
			return nil, fmt.Errorf("migration ID cannot be 0 in file: %s", filename)
		}

		if id < fromID || id > toID {
			continue
		}

		// Leer contenido del archivo SQL
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}
		if len(content) == 0 {
			return nil, fmt.Errorf("migration file is empty: %s", filename)
		}

		// Verificar que existe el archivo opuesto
		counterpartPath := filepath.Join(path, fmt.Sprintf("%s.%s.sql", noSuffix, direction[!inverse]))
		counterpartContent, err := os.ReadFile(counterpartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read counterpart file for %s: %w", filename, err)
		}
		if len(counterpartContent) == 0 {
			return nil, fmt.Errorf("counterpart migration file is empty: %s", counterpartPath)
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
			return fmt.Errorf("failed to start transaction: %w", err)
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
					return fmt.Errorf("migration %d failed: %v, additionally rollback failed: %v", mig.ID, err, rollErr)
				}
				return fmt.Errorf("migration %d failed: %w", mig.ID, err)
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, mig.ID, mig.Name); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("failed to register migration %d: %v, additionally rollback failed: %v", mig.ID, err, rollErr)
			}
			return fmt.Errorf("failed to register migration %d: %w", mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", mig.ID, err)
		}
	}

	return nil
}

func (m *Migrator) down(migrations []Migration) error {
	for _, mig := range migrations {
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
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
					return fmt.Errorf("migration %d rollback failed: %v, additionally transaction rollback failed: %v", mig.ID, err, rollErr)
				}
				return fmt.Errorf("migration %d rollback failed: %w", mig.ID, err)
			}
		}

		// Eliminar registro de migración
		if _, err := tx.Exec(`DELETE from MIGRATIONS WHERE id = ?`, mig.ID); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("failed to remove migration %d record: %v, additionally rollback failed: %v", mig.ID, err, rollErr)
			}
			return fmt.Errorf("failed to remove migration %d record: %w", mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d rollback: %w", mig.ID, err)
		}
	}

	return nil
}

func (m *Migrator) Version() (int, error) {
	if !isConnected(m.db) {
		return 0, fmt.Errorf("db in migrations is desconnected")
	}

	version, err := m.findLastID()
	return version, err
}

func (m *Migrator) Move(steps int, inverse bool) error {
	if !isConnected(m.db) {
		return fmt.Errorf("db in migrations is desconnected")
	}
	// Inicializar tabla de migraciones si no existe
	if err := m.init(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Cargar migraciones
	migrations, err := m.load(*m.options.path, inverse, steps)

	if err != nil {
		return err
	}

	// Verificar si hay migraciones para ejecutar
	if len(migrations) == 0 {
		return fmt.Errorf("no migrations to run")
	}

	// Ejecutar migraciones según la dirección
	if !inverse {
		if err := m.up(migrations); err != nil {
			return fmt.Errorf("failed to run up migrations: %w", err)
		}
		return nil
	}

	if err := m.down(migrations); err != nil {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}
	return nil
}
