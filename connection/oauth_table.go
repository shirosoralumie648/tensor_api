package connection

import (
	"chat/globals"
	"database/sql"
	"fmt"
)

func CreateOAuthTable(db *sql.DB) {
	_, err := globals.ExecDb(db, `
		CREATE TABLE IF NOT EXISTS oauth (
		  id INT PRIMARY KEY AUTO_INCREMENT,
		  provider VARCHAR(32) NOT NULL,
		  open_id VARCHAR(255) NOT NULL,
		  union_id VARCHAR(255),
		  user_id INT NOT NULL,
		  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		  UNIQUE KEY (provider, open_id),
		  FOREIGN KEY (user_id) REFERENCES auth(id)
		);
	`)
	if err != nil {
		fmt.Println(err)
	}
}
