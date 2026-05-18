package builddb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	// 1️⃣ Chemin de la base
	dbPath := "./data/RideUp.db"

	// 2️⃣ Créer le dossier si nécessaire
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("impossible de créer le dossier data: %w", err)
	}

	// 3️⃣ Vérifier si la base existe déjà
	dbExists := true
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbExists = false
	}

	// 4️⃣ Ouvrir ou créer la base
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir la DB: %w", err)
	}

	// 5️⃣ Activer les foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("impossible d'activer les foreign keys: %w", err)
	}

	// 6️⃣ Si la DB n’existe pas, exécuter le script SQL
	if !dbExists {
		fmt.Println("📀 Création d'une nouvelle base de données...")

		schema, err := os.ReadFile("./data/schemaRideUp.sql")
		if err != nil {
			return nil, fmt.Errorf("impossible de lire le fichier SQL: %w", err)
		}

		if _, err := db.Exec(string(schema)); err != nil {
			return nil, fmt.Errorf("erreur lors de la création des tables: %w", err)
		}

		fmt.Println("✅ Base de données créée avec succès !")
	} else {
		fmt.Println("🗂️  Base existante détectée, aucune recréation nécessaire.")
	}

	// 7️⃣ Retourner la DB ouverte (sans la fermer ici)
	return db, nil
}
