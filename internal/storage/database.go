package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/ryangerardwilson/boxes/internal/core"
	_ "modernc.org/sqlite"
)

func (s Store) SaveHistory(previous core.DayState, current core.DayState, config core.Config) error {
	if s.Paths.DatabasePath == "" || config.LeafCount() == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.Paths.DatabasePath), 0o755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", s.Paths.DatabasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	previousChecked := previous.CheckedSet()
	currentChecked := current.CheckedSet()

	for _, item := range config.Flatten() {
		if item.HasChildren {
			continue
		}

		checked := currentChecked[item.ID]
		if _, err := tx.Exec(
			`insert into box_daily_status (date, item_id, label, checked, updated_at)
			 values (?, ?, ?, ?, ?)
			 on conflict(date, item_id) do update set
			   label = excluded.label,
			   checked = excluded.checked,
			   updated_at = excluded.updated_at`,
			current.Date,
			item.ID,
			item.Label,
			boolInt(checked),
			now,
		); err != nil {
			return err
		}

		if previousChecked[item.ID] == checked {
			continue
		}
		if _, err := tx.Exec(
			`insert into box_events (occurred_at, date, item_id, label, checked)
			 values (?, ?, ?, ?, ?)`,
			now,
			current.Date,
			item.ID,
			item.Label,
			boolInt(checked),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
create table if not exists box_daily_status (
  date text not null,
  item_id text not null,
  label text not null,
  checked integer not null,
  updated_at text not null,
  primary key (date, item_id)
);

create table if not exists box_events (
  id integer primary key autoincrement,
  occurred_at text not null,
  date text not null,
  item_id text not null,
  label text not null,
  checked integer not null
);

create index if not exists idx_box_events_date on box_events(date);
create index if not exists idx_box_events_item_id on box_events(item_id);
`)
	return err
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
