package database

import (
	"database/sql"
	"fmt"

	hsBC "github.com/gallactic/hubble_server/blockchain"
	config "github.com/gallactic/hubble_server/config"
	_ "github.com/lib/pq" //dependency for postgre
)

//Postgre adapter
type Postgre struct {
	Config *config.Config
	ObjDB  *sql.DB //Opened DB
}

//Connect to database
func (obe *Postgre) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		obe.Config.DataBase.Host, obe.Config.DataBase.Port, obe.Config.DataBase.User,
		obe.Config.DataBase.Password, obe.Config.DataBase.DBName)

	var err error
	obe.ObjDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		obe.ObjDB.Close()
		return err
	}

	err = obe.ObjDB.Ping()
	if err != nil {
		obe.ObjDB.Close()
		return err
	}
	return nil
}

//Disconnect close connection to database
func (obe *Postgre) Disconnect() error {
	closeError := obe.ObjDB.Close()
	return closeError
}

//InsertAccount add new Account to accounts table
func (obe *Postgre) InsertAccount(acc *hsBC.Account) error {

	sqlStatement := `INSERT INTO accounts (address, balance, permission,sequence,code)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id`
	id := 0
	err := obe.ObjDB.QueryRow(sqlStatement, acc.Address, acc.Balance, acc.Permission, acc.Sequence, acc.Code).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

//UpdateAccount modifies all fields for selected account
func (obe *Postgre) UpdateAccount(id int, acc *hsBC.Account) error {
	sqlStatement := `UPDATE accounts
				SET address = $2, balance = $3, permission = $4, sequence = $5, code = $6
				WHERE id = $1
				RETURNING id, address;`
	var retAddress string
	var retID int
	err := obe.ObjDB.QueryRow(sqlStatement, id, acc.Address, acc.Balance, acc.Permission, acc.Sequence, acc.Code).Scan(&retID, &retAddress)

	if err != nil {
		return err
	}

	return nil
}

//GetAccount finds account in db and returns its data
func (obe *Postgre) GetAccount(id int) (*hsBC.Account, error) {
	sqlStatement := `SELECT * FROM accounts 
					 WHERE id=$1;`
	acc := &hsBC.Account{Address: "", Balance: 0.0, Permission: "", Sequence: 0, Code: ""}
	row := obe.ObjDB.QueryRow(sqlStatement, id)
	err := row.Scan(&acc.Address, &acc.ID, &acc.Balance, &acc.Permission, &acc.Sequence, &acc.Code)
	switch err {
	case sql.ErrNoRows:
		return nil, err
	case nil:
		return acc, nil
	default:
		return nil, err
	}
}

//GetAccountsTableLastID returns last block number
func (obe *Postgre) GetAccountsTableLastID() (uint64, error) {
	sqlStatement := `SELECT MAX(id) FROM accounts;`

	row := obe.ObjDB.QueryRow(sqlStatement)
	var LastID uint64
	err := row.Scan(&LastID)
	switch err {
	case sql.ErrNoRows:
		return 0, err
	case nil:
		return LastID, nil
	default:
		return 0, err
	}
}

//InsertBlock add a block in database
func (obe *Postgre) InsertBlock(b *hsBC.Block) error {
	sqlStatement := `INSERT INTO blocks (id, height, hash, chainID, time, lastblockhash, txcounts)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`
	id := 0
	row := obe.ObjDB.QueryRow(sqlStatement, b.Height, b.Height, b.Hash, b.ChainID, b.Time, b.LastBlockHash, b.TxCounts)
	err := row.Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

//UpdateBlock modifies a block data in database
func (obe *Postgre) UpdateBlock(id int, b *hsBC.Block) error {
	//TODO:
	return nil
}

//GetBlock returns a block
func (obe *Postgre) GetBlock(id int) (*hsBC.Block, error) {
	//TODO:
	return nil, nil
}

//GetBlocksTableLastID returns last block number
func (obe *Postgre) GetBlocksTableLastID() (uint64, error) {
	sqlStatement := `SELECT MAX(id) FROM blocks
					 WHERE id=height;`

	row := obe.ObjDB.QueryRow(sqlStatement)
	var LastID uint64
	err := row.Scan(&LastID)
	switch err {
	case sql.ErrNoRows:
		return 0, err
	case nil:
		return LastID, nil
	default:
		return 0, err
	}
}

//InsertTx add a transaction in database
func (obe *Postgre) InsertTx(b *hsBC.Transaction) error {
	sqlStatement := `INSERT INTO transactions (block_id, txhash, gas_used,gas_wanted,data, time)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
	id := 0
	row := obe.ObjDB.QueryRow(sqlStatement, b.BlockID, b.Hash, b.GasUsed, b.GasWanted, b.Data, b.Time)
	err := row.Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

//UpdateTx modifies a transaction data in database
func (obe *Postgre) UpdateTx(id int, b *hsBC.Transaction) error {
	sqlStatement := `UPDATE transactions
	SET block_id = $2, txhash = $3, gas_used = $4, gas_wanted = $5, data = $6, time = $7
	WHERE id = $1
	RETURNING id, txhash;`
	var retHash string
	var retID int
	err := obe.ObjDB.QueryRow(sqlStatement, id, b.BlockID, b.Hash, b.GasUsed, b.GasWanted, b.Data, b.Time).Scan(&retID, &retHash)

	if err != nil {
		return err
	}

	return nil
}

//GetTx returns a transaction data
func (obe *Postgre) GetTx(hash string) (*hsBC.Transaction, error) {
	sqlStatement := `SELECT block_id,txhash,gas_used,gas_wanted,data,time FROM transactions
					 WHERE txhash=$1;`
	var tx hsBC.Transaction
	obe.ObjDB.QueryRow(sqlStatement, hash).Scan(&tx.BlockID, &tx.Hash, &tx.GasUsed, &tx.GasWanted, &tx.Data, &tx.Time)

	return &tx, nil
}

//GetTXsTableLastID returns last saved transaction number
func (obe *Postgre) GetTXsTableLastID() (uint64, error) {
	sqlStatement := `SELECT MAX(id) FROM transactions;`

	row := obe.ObjDB.QueryRow(sqlStatement)
	var LastID uint64
	err := row.Scan(&LastID)
	switch err {
	case sql.ErrNoRows:
		return 0, err
	case nil:
		return LastID, nil
	default:
		return 0, err
	}
}
