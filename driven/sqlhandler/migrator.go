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
	DB      *sql.DB
	options migrOpts
}

func NewMigrator(db *sql.DB, opts ...MigrOption) *Migrator {
	m := &Migrator{}
	for _, opt := range opts {
		opt(&m.options)
	}

	if db == nil {
		panic("cannot create migrator with nil database connection")
	}
    m.DB = db

	return m
}

type Migration struct {
	ID   int
	Name string
	SQL  string
}

func (migrator *Migrator) init() error {
	_, err := migrator.DB.Exec(`
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

func (migrator *Migrator) findLastID() (int, error) {
	var lastID int
	err := migrator.DB.QueryRow(`
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

func (migrator *Migrator) load(path string, down bool, steps int) ([]Migration, error) {
	direction := map[bool]string{true: "down", false: "up"}

	if steps < 0 {
		return nil, fmt.Errorf("steps cannot be negative, got %d", steps)
	}

	lastID, err := migrator.findLastID()
	if err != nil {
		return nil, fmt.Errorf("failed to find last migration ID: %w", err)
	}

	filenames, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s.sql", direction[down])))
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	maxIDs := len(filenames)

	// Calcular steps si es 0
	if steps == 0 {
		if !down {
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
	if !down {
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
		noSuffix := strings.TrimSuffix(name, fmt.Sprintf(".%s.sql", direction[down]))
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
		counterpartPath := filepath.Join(path, fmt.Sprintf("%s.%s.sql", noSuffix, direction[!down]))
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
		if !down {
			return migrations[i].ID < migrations[j].ID
		}
		return migrations[i].ID > migrations[j].ID
	})

	return migrations, nil
}

func (migrator *Migrator) up(migrations []Migration) error {
	for _, m := range migrations {
		tx, err := migrator.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Ejecutar statements
		stmts := strings.Split(m.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				rollErr := tx.Rollback()
				if rollErr != nil {
					// Aquí retornamos ambos errores ya que es crítico saber si falló tanto la migración como el rollback
					return fmt.Errorf("migration %d failed: %v, additionally rollback failed: %v", m.ID, err, rollErr)
				}
				return fmt.Errorf("migration %d failed: %w", m.ID, err)
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, m.ID, m.Name); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("failed to register migration %d: %v, additionally rollback failed: %v", m.ID, err, rollErr)
			}
			return fmt.Errorf("failed to register migration %d: %w", m.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", m.ID, err)
		}
	}

	return nil
}

func (migrator *Migrator) down(migrations []Migration) error {
	for _, m := range migrations {
		tx, err := migrator.DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Ejecutar statements
		stmts := strings.Split(m.SQL, ";")
		for _, s := range stmts[:len(stmts)-1] {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			if _, err := tx.Exec(fmt.Sprintf("%s;", s)); err != nil {
				rollErr := tx.Rollback()
				if rollErr != nil {
					return fmt.Errorf("migration %d rollback failed: %v, additionally transaction rollback failed: %v", m.ID, err, rollErr)
				}
				return fmt.Errorf("migration %d rollback failed: %w", m.ID, err)
			}
		}

		// Eliminar registro de migración
		if _, err := tx.Exec(`DELETE from MIGRATIONS WHERE id = ?`, m.ID); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("failed to remove migration %d record: %v, additionally rollback failed: %v", m.ID, err, rollErr)
			}
			return fmt.Errorf("failed to remove migration %d record: %w", m.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d rollback: %w", m.ID, err)
		}
	}

	return nil
}

func (migrator *Migrator) Move(down bool, steps int) error {
	// Inicializar tabla de migraciones si no existe
	if err := migrator.init(); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Cargar migraciones
	migrations, err := migrator.load(*migrator.options.path, down, steps)

	if err != nil {
		return err
	}

	// Verificar si hay migraciones para ejecutar
	if len(migrations) == 0 {
		return fmt.Errorf("no migrations to run")
	}

	// Ejecutar migraciones según la dirección
	if !down {
		if err := migrator.up(migrations); err != nil {
			return fmt.Errorf("failed to run up migrations: %w", err)
		}
		return nil
	}

	if err := migrator.down(migrations); err != nil {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}
	return nil
}
