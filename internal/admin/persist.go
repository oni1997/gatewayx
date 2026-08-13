package admin

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Persistence struct {
	db *sql.DB
}

func NewPersistence(path string) (*Persistence, error) {
	if path == "" {
		path = filepath.Join("data", "gatewayx.db")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	p := &Persistence{db: db}

	if err := p.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return p, nil
}

func (p *Persistence) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		owner TEXT NOT NULL,
		key TEXT NOT NULL UNIQUE,
		prefix TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		last_used TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS certificates (
		id TEXT PRIMARY KEY,
		domain TEXT NOT NULL,
		issuer TEXT NOT NULL,
		not_after TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
	CREATE INDEX IF NOT EXISTS idx_certs_domain ON certificates(domain);
	`
	_, err := p.db.Exec(schema)
	return err
}

func (p *Persistence) SaveKey(k *APIKey) error {
	_, err := p.db.Exec(
		`INSERT OR REPLACE INTO api_keys (id, name, owner, key, prefix, created_at, last_used)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.Name, k.Owner, k.Key, k.Prefix, k.CreatedAt, k.LastUsed,
	)
	return err
}

func (p *Persistence) DeleteKey(id string) error {
	_, err := p.db.Exec(`DELETE FROM api_keys WHERE id = ?`, id)
	return err
}

func (p *Persistence) LoadKeys() ([]*APIKey, error) {
	rows, err := p.db.Query(`SELECT id, name, owner, key, prefix, created_at, last_used FROM api_keys`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		var lastUsed sql.NullTime
		if err := rows.Scan(&k.ID, &k.Name, &k.Owner, &k.Key, &k.Prefix, &k.CreatedAt, &lastUsed); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			k.LastUsed = lastUsed.Time
		}
		keys = append(keys, &k)
	}
	return keys, rows.Err()
}

func (p *Persistence) SaveCert(c *Certificate) error {
	_, err := p.db.Exec(
		`INSERT OR REPLACE INTO certificates (id, domain, issuer, not_after, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.Domain, c.Issuer, c.NotAfter, c.CreatedAt,
	)
	return err
}

func (p *Persistence) LoadCerts() ([]*Certificate, error) {
	rows, err := p.db.Query(`SELECT id, domain, issuer, not_after, created_at FROM certificates`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var certs []*Certificate
	for rows.Next() {
		var c Certificate
		if err := rows.Scan(&c.ID, &c.Domain, &c.Issuer, &c.NotAfter, &c.CreatedAt); err != nil {
			return nil, err
		}
		certs = append(certs, &c)
	}
	return certs, rows.Err()
}

func (p *Persistence) Close() error {
	return p.db.Close()
}

func (p *Persistence) UpdateLastUsed(id string, t time.Time) error {
	_, err := p.db.Exec(`UPDATE api_keys SET last_used = ? WHERE id = ?`, t, id)
	return err
}
