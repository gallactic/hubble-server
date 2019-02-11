package explorer

import (
	"fmt"

	bc "github.com/gallactic/hubble_server/blockchain"
	config "github.com/gallactic/hubble_server/config"
	db "github.com/gallactic/hubble_server/database"
)

var numDots int

//Explorer class for connecting block chain to data base
type Explorer struct {
	BCAdapter bc.Adapter
	DBAdapter db.Adapter
	Config    *config.Config
}

//Init to initialize database and block chain
func (e *Explorer) Init() error {
	//connect to gallactic blockchain by gRPC
	bcAdapter := e.BCAdapter
	clientErr := bcAdapter.CreateGRPCClient()
	if clientErr == nil {
		println("Blockchain client created successfully!")
	} else {
		return clientErr
	}
	bcAdapter.Update()

	//connect to database
	dbAdapter := e.DBAdapter
	connErr := dbAdapter.Connect()
	if connErr != nil {
		return connErr
	}
	println("Connected to database successfully!")

	return nil
}

//Update to Sync database with blockchain
func (e *Explorer) Update() error {

	//defer dbAdapter.Disconnect()
	bcAdapter := e.BCAdapter
	dbAdapter := e.DBAdapter

	//Sync data with blockchain
	updateErr := bcAdapter.Update()
	if updateErr != nil {
		return updateErr
	}

	//Get current height of last block
	currentHeight, getLastHeightErr := bcAdapter.GetBlocksLastHeight()
	if getLastHeightErr != nil {
		return getLastHeightErr
	}

	//Get last block ID that is saved
	lastBlockIDInDB, getLastIDError := dbAdapter.GetBlocksTableLastID()
	if getLastIDError != nil {
		lastBlockIDInDB = 0
	}

	if currentHeight > lastBlockIDInDB {
		d := currentHeight - lastBlockIDInDB
		n := int(d / 1000)
		if d > 1000 {
			println("Saving new blocks number", lastBlockIDInDB, " to ", currentHeight, " in database...")
		}

		var startIndex uint64
		var endIndex uint64

		startBlockID := lastBlockIDInDB + 1
		if lastBlockIDInDB == 0 {
			startBlockID = 0
		}

		for i := 0; i <= n; i++ {
			startIndex = startBlockID + uint64(i*1000)
			endIndex = startIndex + 999
			if endIndex > currentHeight {
				endIndex = currentHeight
			}
			blocks, _ := bcAdapter.GetBlocks(startIndex, endIndex)
			savingErr := e.saveBlocksInDB(blocks, bcAdapter, dbAdapter)
			if savingErr != nil {
				return savingErr
			}

			perc := (int)((float64(i+1) / float64(n+1)) * 100.0)
			fmt.Printf("\r%d%% saved! (%d/%d)", perc, startIndex-lastBlockIDInDB, d)

		}

		if d > 1000 {
			println("\r", d, "new blocks saved!                   ")
			println("Checking new blocks...")
		} else {
			e.writeAnim(currentHeight)
		}
	} else {
		e.writeAnim(currentHeight)
	}

	return nil
}

func (e *Explorer) saveBlocksInDB(blocks []*bc.Block, bcAdapter bc.Adapter, dbAdapter db.Adapter) error {
	l := len(blocks)
	if l <= 0 {
		return fmt.Errorf("Empty Blocks Array")
	}
	for i := l - 1; i >= 0; i-- {
		block := blocks[i]
		err := dbAdapter.InsertBlock(block)
		if err != nil {
			return err
		}
		if block.TxCounts > 0 {
			errTxSave := e.saveBlockTXsInDB(block, bcAdapter, dbAdapter)
			if errTxSave != nil {
				return errTxSave
			}
		}
	}
	return nil
}

func (e *Explorer) saveBlockTXsInDB(block *bc.Block, bcAdapter bc.Adapter, dbAdapter db.Adapter) error {
	l := block.TxCounts
	if l <= 0 {
		return fmt.Errorf("Empty Transactions Array")
	}

	height := uint64(block.Height)
	txs, errTXs := bcAdapter.GetTXs(height)
	if errTXs != nil {
		return errTXs
	}

	for i := l - 1; i >= 0; i-- {
		err := dbAdapter.InsertTx(&txs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Explorer) writeAnim(currentHeight uint64) {
	numDots++
	if numDots > 3 {
		numDots = 0
	}
	dotStr := ""
	for nDot := 1; nDot <= numDots; nDot++ {
		dotStr += "."
	}
	fmt.Printf("\r%d blocks saved"+dotStr+"       ", currentHeight)
}
