package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"

	config "github.com/gallactic/hubble_server/config"
	pb "github.com/gallactic/hubble_server/proto3"
	"google.golang.org/grpc"
)

//Gallactic class for connecting to Gallactic block chain
type Gallactic struct {
	client   *pb.BlockChainClient
	blocks   *pb.BlocksResponse
	accounts *pb.AccountsResponse
	chain    *pb.BlockchainInfoResponse
	Config   *config.Config
}

//CreateGRPCClient creates a client for communicating with gallactic blockchain
func (g *Gallactic) CreateGRPCClient() error {
	var connURL string
	connURL = g.Config.GRPC.URL + ":" + g.Config.GRPC.Port
	conn, err := grpc.Dial(connURL, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewBlockChainClient(conn)
	g.client = &client
	return nil
}

//Update will refresh all data and sync with block chain
func (g *Gallactic) Update() error {
	client := *g.client
	var chainInfoErr error
	g.chain, chainInfoErr = client.GetBlockchainInfo(context.Background(), &pb.Empty{})
	if chainInfoErr != nil {
		return chainInfoErr
	}

	return nil
}

//GetBlocksLastHeight return last height
func (g *Gallactic) GetBlocksLastHeight() (uint64, error) {
	return g.chain.GetLastBlockHeight(), nil
}

//GetBlockInfo returns specified block
func (g *Gallactic) GetBlockInfo(height uint64) (*BlockInfo, error) {

	client := *g.client
	blockRes, getBlockErr := client.GetBlock(context.Background(), &pb.BlockRequest{Height: height})
	if getBlockErr != nil {
		return nil, getBlockErr
	}

	header := blockRes.GetBlock().GetHeader()
	var inf BlockInfo
	toBlockInfo(header, &inf)
	return &inf, nil
}

//GetBlock returns specified block
func (g *Gallactic) GetBlock(height uint64) (*Block, error) {

	lastID, lastIDErr := g.GetBlocksLastHeight()
	if lastIDErr != nil {
		return nil, lastIDErr
	}
	if height > lastID {
		return nil, fmt.Errorf("block height out of range (max is " + string(lastID) + ")")
	}

	client := *g.client
	blockRes, getBlockErr := client.GetBlock(context.Background(), &pb.BlockRequest{Height: height})
	if getBlockErr != nil {
		return nil, getBlockErr
	}
	var b Block
	toBlock(blockRes, &b)

	return &b, nil
}

//GetAccountsCount returns number of accounts
func (g *Gallactic) GetAccountsCount() int {
	l := len(g.accounts.Accounts)
	return l
}

//GetAccount returns specified account
func (g *Gallactic) GetAccount(id int) (*Account, error) {
	acc := g.accounts.Accounts[id].Account
	ID := uint64(id)
	var retAcc Account
	toAccount(acc, &retAcc)
	retAcc.ID = ID
	return &retAcc, nil
}

//GetAccounts returns all accounts in array of accounts
func (g *Gallactic) GetAccounts() ([]*Account, error) {
	l := len(g.accounts.Accounts)

	retAccounts := make([]*Account, l)

	for i := 0; i < l; i++ {
		acc := g.accounts.Accounts[i].Account
		ID := uint64(i)
		toAccount(acc, retAccounts[i])
		retAccounts[i].ID = ID
	}

	return retAccounts, nil
}

//GetBlocksInfo returns a group of blocks for faster access them
func (g *Gallactic) GetBlocksInfo(from uint64, to uint64) ([]*BlockInfo, error) {
	client := *g.client
	blocks, getBlocksErr := client.GetBlocks(context.Background(), &pb.BlocksRequest{MinHeight: from, MaxHeight: to})
	if getBlocksErr != nil {
		return nil, getBlocksErr
	}

	n := len(blocks.GetBlocks())
	retBlocks := make([]*BlockInfo, 0, n)
	for i := 0; i < n; i++ {
		toBlockInfo(blocks.GetBlocks()[i].GetHeader(), retBlocks[i])
	}

	return retBlocks, nil
}

//GetBlocks returns a group of blocks for faster access them
func (g *Gallactic) GetBlocks(from uint64, to uint64) ([]*Block, error) {
	client := *g.client
	blocks, getBlocksErr := client.GetBlocks(context.Background(), &pb.BlocksRequest{MinHeight: from, MaxHeight: to})
	if getBlocksErr != nil {
		return nil, getBlocksErr
	}

	n := len(blocks.GetBlocks())
	retBlocks := make([]*Block, n)

	for i := 0; i < n; i++ {
		retBlocks[i] = new(Block)
		BlockInfoToBlock(blocks.GetBlocks()[i].GetHeader(), retBlocks[i])
	}

	return retBlocks, nil
}

/*
type BlockTxsResponse struct {
	Count                int32                                         `protobuf:"varint,1,opt,name=Count,proto3" json:"Count,omitempty"`
	Txs                  []github_com_gallactic_gallactic_txs.Envelope `protobuf:"bytes,3,rep,name=Txs,proto3,customtype=github.com/gallactic/gallactic/txs.Envelope" json:"Txs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                      `json:"-"`
	XXX_unrecognized     []byte                                        `json:"-"`
	XXX_sizecache        int32                                         `json:"-"`
}

type Envelope struct {
	ChainID     string             `json:"chainId"`
	Type        tx.Type            `json:"type"`
	Tx          tx.Tx              `json:"tx"`
	Signatories []crypto.Signatory `json:"signatories,omitempty"`
	hash        []byte
}

type Signatory struct {
	PublicKey PublicKey `json:"publicKey"`
	Signature Signature `json:"signature"`
}
type Signature struct {
	data signatureData
}

type signatureData struct {
	Signature []byte `json:"signature"`
}


type Tx interface {
	Signers() []TxInput
	Type() Type
	Amount() uint64
	Fee() uint64
	EnsureValid() error
}


type TxInput struct {
	Address  crypto.Address `json:"address"`
	Amount   uint64         `json:"amount"`
	Sequence uint64         `json:"sequence"`
}


	if txs.Txs[0].Tx.Type() == 1 {
		sndTx := txs.Txs[0].Tx
		println(sndTx.Amount())
	}
	println("TX Count: ", txs.Count)
	println("TX Chain ID: ", txs.Txs[0].ChainID)
	println("TX Hash: ", hex.EncodeToString(txs.Txs[0].Hash()))
	println("TX Signatories Public Key: ", txs.Txs[0].Signatories[0].PublicKey.String())
	println("TX Signatories ACC Address: ", txs.Txs[0].Signatories[0].PublicKey.AccountAddress().String())
	println("TX Signatories VAL Address: ", txs.Txs[0].Signatories[0].PublicKey.ValidatorAddress().String())
	println("TX Signatories Signature: ", txs.Txs[0].Signatories[0].Signature.String())
	println("Num Signers: ", len(txs.Txs[0].Tx.Signers()))
	println("TX Signers Address: ", txs.Txs[0].Tx.Signers()[0].Address.String())
	println("TX Signers Amount: ", txs.Txs[0].Tx.Signers()[0].Amount)
	println("TX Signers Sequence: ", txs.Txs[0].Tx.Signers()[0].Sequence)
	println("TX Amount: ", txs.Txs[0].Tx.Amount())
	println("TX Type: ", txs.Txs[0].Tx.Type())
	println("TX Type Striing: ", txs.Txs[0].Tx.Type().String())
	println("TX Fee: ", txs.Txs[0].Tx.Fee())
	println("TX Err: ", txs.Txs[0].Tx.EnsureValid())
*/

//GetTXsCount returns number of TXs
func (g *Gallactic) GetTXsCount(height uint64) int {
	client := *g.client
	txs, _ := client.GetBlockTxs(context.Background(), &pb.BlockRequest{Height: height})
	n := int(txs.Count)
	return n
}

//GetTx returns specified TX
func (g *Gallactic) GetTx(height uint64, hash []byte) (*Transaction, error) {

	client := *g.client

	blockRes, getBlockErr := client.GetBlock(context.Background(), &pb.BlockRequest{Height: height})
	if getBlockErr != nil {
		return nil, getBlockErr
	}
	var b Block
	toBlock(blockRes, &b)

	findHash := hex.EncodeToString(hash)
	txRes, getTxErr := client.GetTx(context.Background(), &pb.TxRequest{TxHash: findHash})
	if getTxErr != nil {
		return nil, getTxErr
	}

	var tx Transaction
	toTx(txRes, &tx)
	tx.Time = b.Time
	tx.BlockID = int64(height)

	return &tx, nil
}

//GetTXs returns all transaction of specific block
func (g *Gallactic) GetTXs(height uint64) ([]Transaction, error) {
	client := *g.client
	txs, getTXsErr := client.GetBlockTxs(context.Background(), &pb.BlockRequest{Height: height})
	if getTXsErr != nil {
		return nil, getTXsErr
	}

	block, _ := g.GetBlock(height)
	n := int(txs.Count)
	retTXs := make([]Transaction, n)
	for i := 0; i < n; i++ {
		retTXs[i].BlockID = int64(height)
		retTXs[i].Hash = txs.Txs[i].Hash
		retTXs[i].GasUsed = txs.Txs[i].GetGasUsed()
		retTXs[i].GasWanted = txs.Txs[i].GetGasWanted()
		retTXs[i].Data = "" //TODO: fix data
		retTXs[i].Time = block.Time
	}

	return retTXs, nil
}
