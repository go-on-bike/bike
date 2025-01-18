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

func (h *sqlHandler) initMigrations() error {
	fmt.Fprintf(h.stderr, "%s init migrations: executing a query in init", Sig)
	_, err := h.db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("%s init migrations: failed to initialize migrations table: %w", Sig, err)
	}
	return nil
}

func (h *sqlHandler) findMigrationLastID() (int, error) {
	var lastID int
	fmt.Fprintf(h.stderr, "%s find last migration id: executing query row in find last id", Sig)
	err := h.db.QueryRow(`
        SELECT id 
        FROM migrations 
        ORDER BY id DESC 
        LIMIT 1
    `).Scan(&lastID)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%s find last migration id: failed to find last migration ID: %w", Sig, err)
	}
	return lastID, nil
}

type Migration struct {
	ID   int
	Name string
	SQL  string
}

func (h *sqlHandler) load(path string, inverse bool, steps int) ([]Migration, error) {
	direction := map[bool]string{true: "down", false: "up"}

	if steps < 0 {
		return nil, fmt.Errorf("%s load: steps cannot be negative, got %d", Sig, steps)
	}

	lastID, err := h.findMigrationLastID()
	if err != nil {
		return nil, fmt.Errorf("%s load: failed to find last migration ID: %w", Sig, err)
	}

	filenames, err := filepath.Glob(filepath.Join(path, fmt.Sprintf("*.%s.sql", direction[inverse])))
	if err != nil {
		return nil, fmt.Errorf("%s load: failed to get migration files: %w", Sig, err)
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
		return nil, fmt.Errorf("%s load: step calculation failed, got %d", Sig, steps)
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
			return nil, fmt.Errorf("%s load: invalid migration filename format: %s", Sig, filename)
		}

		id, err := strconv.Atoi(nameParts[0])
		if err != nil {
			return nil, fmt.Errorf("%s load: invalid migration ID in filename %s: %w", Sig, filename, err)
		}
		if id == 0 {
			return nil, fmt.Errorf("%s load: migration ID cannot be 0 in file: %s", Sig, filename)
		}

		if id < fromID || id > toID {
			continue
		}

		// Leer contenido del archivo SQL
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("%s load: failed to read migration file %s: %w", Sig, filename, err)
		}
		if len(content) == 0 {
			return nil, fmt.Errorf("%s load: migration file is empty: %s", Sig, filename)
		}

		// Verificar que existe el archivo opuesto
		counterpartPath := filepath.Join(path, fmt.Sprintf("%s.%s.sql", noSuffix, direction[!inverse]))
		counterpartContent, err := os.ReadFile(counterpartPath)
		if err != nil {
			return nil, fmt.Errorf("%s load: failed to read counterpart file for %s: %w", Sig, filename, err)
		}
		if len(counterpartContent) == 0 {
			return nil, fmt.Errorf("%s load: counterpart migration file is empty: %s", Sig, counterpartPath)
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

func (h *sqlHandler) up(migrations []Migration) error {
	for _, mig := range migrations {
		tx, err := h.db.Begin()
		if err != nil {
			return fmt.Errorf("%s up: failed to start transaction: %w", Sig, err)
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
					return fmt.Errorf("%s up: migration %d failed: %v, additionally rollback failed: %v", Sig, mig.ID, err, rollErr)
				}
				return fmt.Errorf("%s up: migration %d failed: %w", Sig, mig.ID, err)
			}
		}

		// Registrar migración
		if _, err := tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, mig.ID, mig.Name); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("%s up: failed to register migration %d: %v,Sig, additionally rollback failed: %v", Sig, mig.ID, err, rollErr)
			}
			return fmt.Errorf("%s up: failed to register migration %d: %w", Sig, mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("%s up: failed to commit migration %d: %w", Sig, mig.ID, err)
		}
	}

	return nil
}

func (h *sqlHandler) down(migrations []Migration) error {
	for _, mig := range migrations {
		tx, err := h.db.Begin()
		if err != nil {
			return fmt.Errorf("%s down: failed to start transaction: %w", Sig, err)
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
					return fmt.Errorf("%s down: migration %d rollback failed: %v,Sig, additionally transaction rollback failed: %v", Sig, mig.ID, err, rollErr)
				}
				return fmt.Errorf("%s down: migration %d rollback failed: %w", Sig, mig.ID, err)
			}
		}

		// Eliminar registro de migración
		if _, err := tx.Exec(`DELETE from MIGRATIONS WHERE id = ?`, mig.ID); err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				return fmt.Errorf("%s down: failed to remove migration %d record: %v,Sig, additionally rollback failed: %v", Sig, mig.ID, err, rollErr)
			}
			return fmt.Errorf("%s down: failed to remove migration %d record: %w", Sig, mig.ID, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("%s down: failed to commit migration %d rollback: %w", Sig, mig.ID, err)
		}
	}

	return nil
}

func (h *sqlHandler) Version() (int, error) {
	if !isConnected(h.db) {
		return 0, fmt.Errorf("%s version: db in migrations is desconnected", Sig)
	}

	version, err := h.findMigrationLastID()
	return version, err
}

func (h *sqlHandler) RunMigrations(steps int, inverse bool) error {
	if !isConnected(h.db) {
		return fmt.Errorf("%s run migrations: db in migrations is desconnected", Sig)
	}
	// Inicializar tabla de migraciones si no existe
	if err := h.initMigrations(); err != nil {
		return fmt.Errorf("%s run migrations: failed to initialize migrations: %w", Sig, err)
	}

	// Cargar migraciones
	migrations, err := h.load(*h.options.path, inverse, steps)

	if err != nil {
		return err
	}

	// Verificar si hay migraciones para ejecutar
	if len(migrations) == 0 {
		return fmt.Errorf("%s run migrations: no migrations to run", Sig)
	}

	// Ejecutar migraciones según la dirección
	if !inverse {
		if err := h.up(migrations); err != nil {
			return fmt.Errorf("%s run migrations: failed to run up migrations: %w", Sig, err)
		}
		return nil
	}

	if err := h.down(migrations); err != nil {
		return fmt.Errorf("%s run migrations: failed to run down migrations: %w", Sig, err)
	}
	return nil
}
